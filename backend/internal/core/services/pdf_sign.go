package services

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"toolkitz/backend/internal/core/ports"

	"github.com/disintegration/imaging"
)

const maxSignaturePx = 800

// signatureImage menyimpan piksel RGB dan alpha yang sudah dikompres zlib (FlateDecode).
type signatureImage struct {
	rgbZ   []byte
	alphaZ []byte
	width  int
	height int
}

func zlibBytes(b []byte) []byte {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	_, _ = zw.Write(b)
	_ = zw.Close()
	return buf.Bytes()
}

// loadSignature membaca gambar tanda tangan (EXIF diterapkan), memperkecil bila terlalu besar,
// dan opsional menjadikan latar putih/terang transparan.
func loadSignature(path string, removeWhite bool) (signatureImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return signatureImage{}, err
	}
	defer f.Close()

	src, err := imaging.Decode(f, imaging.AutoOrientation(true))
	if err != nil {
		return signatureImage{}, fmt.Errorf("gambar tanda tangan tidak dapat dibaca (gunakan PNG/JPG): %w", err)
	}
	if src.Bounds().Dx() > maxSignaturePx {
		src = imaging.Resize(src, maxSignaturePx, 0, imaging.Lanczos)
	}

	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return signatureImage{}, errors.New("ukuran gambar tanda tangan tidak valid")
	}

	// NRGBA = warna non-premultiplied, sesuai yang diharapkan PDF (RGB + SMask terpisah)
	nrgba := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(nrgba, nrgba.Bounds(), src, b.Min, draw.Src)

	rgb := make([]byte, 0, w*h*3)
	alpha := make([]byte, 0, w*h)
	for i := 0; i < len(nrgba.Pix); i += 4 {
		r, g, bl, a := nrgba.Pix[i], nrgba.Pix[i+1], nrgba.Pix[i+2], nrgba.Pix[i+3]

		if removeWhite {
			lum := r
			if g < lum {
				lum = g
			}
			if bl < lum {
				lum = bl
			}
			switch {
			case lum >= 245:
				a = 0
			case lum >= 200:
				// tepi halus: makin terang makin transparan
				a = byte(float64(a) * float64(245-int(lum)) / 45.0)
			}
		}

		rgb = append(rgb, r, g, bl)
		alpha = append(alpha, a)
	}

	return signatureImage{rgbZ: zlibBytes(rgb), alphaZ: zlibBytes(alpha), width: w, height: h}, nil
}

// parsePageList menerjemahkan "first" | "last" | "all" | "1,3,5-7" menjadi nomor halaman terurut.
func parsePageList(spec string, total int) ([]int, error) {
	spec = strings.ToLower(strings.TrimSpace(spec))
	switch spec {
	case "", "last":
		return []int{total}, nil
	case "first":
		return []int{1}, nil
	case "all":
		all := make([]int, total)
		for i := range all {
			all[i] = i + 1
		}
		return all, nil
	}

	seen := map[int]bool{}
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lo, hi := 0, 0
		if strings.Contains(part, "-") {
			ab := strings.SplitN(part, "-", 2)
			a, errA := strconv.Atoi(strings.TrimSpace(ab[0]))
			b, errB := strconv.Atoi(strings.TrimSpace(ab[1]))
			if errA != nil || errB != nil {
				return nil, fmt.Errorf("pilihan halaman %q tidak valid", part)
			}
			lo, hi = a, b
			if lo > hi {
				lo, hi = hi, lo
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("pilihan halaman %q tidak valid", part)
			}
			lo, hi = n, n
		}
		if lo < 1 || hi > total {
			return nil, fmt.Errorf("halaman %q di luar jangkauan (PDF punya %d halaman)", part, total)
		}
		for n := lo; n <= hi; n++ {
			seen[n] = true
		}
	}
	if len(seen) == 0 {
		return nil, errors.New("pilihan halaman kosong")
	}

	pages := make([]int, 0, len(seen))
	for n := range seen {
		pages = append(pages, n)
	}
	sort.Ints(pages)
	return pages, nil
}

// placeSignature menghitung posisi & ukuran gambar pada halaman (koordinat PDF, titik asal kiri-bawah).
func placeSignature(pageW, pageH, imgW, imgH float64, position string, widthPct float64) (x, y, w, h float64) {
	w = pageW * widthPct / 100
	h = w * imgH / imgW
	if maxH := pageH * 0.5; h > maxH { // jangan sampai menutupi lebih dari setengah halaman
		h = maxH
		w = h * imgW / imgH
	}

	m := 0.04 * math.Min(pageW, pageH)

	position = strings.ToLower(strings.TrimSpace(position))
	if len(position) != 2 || !strings.ContainsRune("tmb", rune(position[0])) || !strings.ContainsRune("lcr", rune(position[1])) {
		position = "br"
	}

	switch position[1] {
	case 'l':
		x = m
	case 'c':
		x = (pageW - w) / 2
	default:
		x = pageW - w - m
	}
	switch position[0] {
	case 't':
		y = pageH - h - m
	case 'm':
		y = (pageH - h) / 2
	default:
		y = m
	}
	return x, y, w, h
}

// buildSignatureOverlay membuat PDF overlay: halaman ke-i berukuran sama dengan halaman target ke-i
// dan berisi gambar tanda tangan di posisi yang diminta.
//
// Objek: 1=Catalog, 2=Pages, 3=SMask (alpha), 4=Image (RGB), lalu per halaman: Page(5+2i), Content(6+2i).
func buildSignatureOverlay(sig signatureImage, sizes [][2]float64, position string, widthPct float64) []byte {
	var buf bytes.Buffer
	var offsets []int

	begin := func() {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n", len(offsets))
	}

	buf.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")

	begin() // 1
	buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	kids := make([]string, 0, len(sizes))
	for i := range sizes {
		kids = append(kids, fmt.Sprintf("%d 0 R", 5+2*i))
	}
	begin() // 2
	fmt.Fprintf(&buf, "<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", strings.Join(kids, " "), len(sizes))

	begin() // 3: SMask
	fmt.Fprintf(&buf,
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceGray /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n",
		sig.width, sig.height, len(sig.alphaZ))
	buf.Write(sig.alphaZ)
	buf.WriteString("\nendstream\nendobj\n")

	begin() // 4: Image
	fmt.Fprintf(&buf,
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /SMask 3 0 R /Length %d >>\nstream\n",
		sig.width, sig.height, len(sig.rgbZ))
	buf.Write(sig.rgbZ)
	buf.WriteString("\nendstream\nendobj\n")

	for i, sz := range sizes {
		pageNum := 5 + 2*i
		contentNum := pageNum + 1

		x, y, w, h := placeSignature(sz[0], sz[1], float64(sig.width), float64(sig.height), position, widthPct)

		begin() // Page
		fmt.Fprintf(&buf,
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /XObject << /Sig 4 0 R >> >> /Contents %d 0 R >>\nendobj\n",
			sz[0], sz[1], contentNum)

		content := fmt.Sprintf("q %.2f 0 0 %.2f %.2f %.2f cm /Sig Do Q", w, h, x, y)
		begin() // Content
		fmt.Fprintf(&buf, "<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(content), content)
	}

	xrefPos := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(offsets)+1)
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xrefPos)

	return buf.Bytes()
}

// Sign menempelkan tanda tangan ke halaman terpilih dengan qpdf --overlay.
func (s *pdfEditService) Sign(inputPath, signaturePath, outputPath string, opts ports.SignOptions) error {
	total, err := pdfPageCount(inputPath)
	if err != nil {
		return err
	}
	targets, err := parsePageList(opts.Pages, total)
	if err != nil {
		return err
	}

	widthPct := opts.WidthPercent
	if widthPct < 5 || widthPct > 80 {
		widthPct = 25
	}

	sig, err := loadSignature(signaturePath, opts.RemoveWhite)
	if err != nil {
		return err
	}

	geoms, err := pdfPageGeoms(inputPath, total)
	if err != nil {
		return err
	}

	// Ukuran overlay = ukuran halaman sebagaimana TAMPAK (halaman terotasi 90/270 ditukar).
	sizes := make([][2]float64, 0, len(targets))
	csv := make([]string, 0, len(targets))
	for _, n := range targets {
		g := geoms[n-1]
		w, h := g.Width, g.Height
		if g.Rot == 90 || g.Rot == 270 {
			w, h = h, w
		}
		sizes = append(sizes, [2]float64{w, h})
		csv = append(csv, strconv.Itoa(n))
	}

	overlayPath := filepath.Join(filepath.Dir(outputPath), "signature-overlay.pdf")
	defer os.Remove(overlayPath)
	if err := os.WriteFile(overlayPath, buildSignatureOverlay(sig, sizes, opts.Position, float64(widthPct)), 0o644); err != nil {
		return err
	}

	return runQpdf(
		inputPath,
		"--overlay", overlayPath,
		fmt.Sprintf("--from=1-%d", len(targets)),
		"--to="+strings.Join(csv, ","),
		"--",
		outputPath,
	)
}
