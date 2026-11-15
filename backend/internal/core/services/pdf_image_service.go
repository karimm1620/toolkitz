package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"toolkitz/backend/internal/core/ports"

	"github.com/disintegration/imaging"
)

type pdfImageService struct{}

func NewPDFImageService() ports.PDFImageService {
	return &pdfImageService{}
}

// pdfPageImage adalah satu gambar yang sudah di-encode JPEG, siap ditanam ke halaman PDF.
type pdfPageImage struct {
	jpegData []byte
	width    int
	height   int
}

const (
	a4WidthPt  = 595.28
	a4HeightPt = 841.89
	a4MarginPt = 24.0
	pxToPt     = 0.75 // 96 dpi -> 72 pt
)

// ImagesToPDF menyusun gambar menjadi satu PDF (satu gambar per halaman).
// Ditulis dengan pustaka standar Go supaya tidak bergantung pada API pdfcpu.
func (s *pdfImageService) ImagesToPDF(inputPaths []string, outputPath, pageMode string) error {
	if len(inputPaths) == 0 {
		return errors.New("tidak ada gambar yang diberikan")
	}

	imgs := make([]pdfPageImage, 0, len(inputPaths))
	for i, p := range inputPaths {
		pi, err := loadPageImage(p)
		if err != nil {
			return fmt.Errorf("gambar ke-%d tidak dapat diproses (format didukung: JPG, PNG, GIF, BMP, TIFF): %w", i+1, err)
		}
		imgs = append(imgs, pi)
	}

	data, err := buildImagesPDF(imgs, pageMode)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o644)
}

// loadPageImage membaca gambar (EXIF orientation diterapkan), meratakan transparansi ke
// latar putih, lalu meng-encode ke JPEG 3-komponen (RGB) untuk ditanam di PDF.
func loadPageImage(path string) (pdfPageImage, error) {
	f, err := os.Open(path)
	if err != nil {
		return pdfPageImage{}, err
	}
	defer f.Close()

	src, err := imaging.Decode(f, imaging.AutoOrientation(true))
	if err != nil {
		return pdfPageImage{}, err
	}

	b := src.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return pdfPageImage{}, errors.New("ukuran gambar tidak valid")
	}

	// Selalu ubah ke RGBA di atas latar putih -> JPEG selalu 3 komponen warna
	// (gambar grayscale/transparan tidak membuat ColorSpace PDF tidak cocok).
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(rgba, rgba.Bounds(), src, b.Min, draw.Over)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 90}); err != nil {
		return pdfPageImage{}, err
	}

	return pdfPageImage{jpegData: buf.Bytes(), width: b.Dx(), height: b.Dy()}, nil
}

// buildImagesPDF menulis dokumen PDF 1.4 minimal: tiap gambar satu halaman.
//
// Susunan objek: 1=Catalog, 2=Pages, lalu per gambar i: Page(3+3i), Content(4+3i), Image(5+3i).
func buildImagesPDF(imgs []pdfPageImage, pageMode string) ([]byte, error) {
	if len(imgs) == 0 {
		return nil, errors.New("tidak ada gambar")
	}
	original := strings.EqualFold(strings.TrimSpace(pageMode), "original")

	var buf bytes.Buffer
	var offsets []int // offsets[n-1] = posisi byte objek nomor n

	begin := func() {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n", len(offsets))
	}

	buf.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")

	// 1: Catalog
	begin()
	buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// 2: Pages
	kids := make([]string, 0, len(imgs))
	for i := range imgs {
		kids = append(kids, fmt.Sprintf("%d 0 R", 3+3*i))
	}
	begin()
	fmt.Fprintf(&buf, "<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", strings.Join(kids, " "), len(imgs))

	for i, im := range imgs {
		pageNum := 3 + 3*i
		contentNum := pageNum + 1
		imageNum := pageNum + 2

		w, h := float64(im.width), float64(im.height)

		var pageW, pageH, drawW, drawH, x, y float64
		if original {
			pageW, pageH = w*pxToPt, h*pxToPt
			drawW, drawH = pageW, pageH
		} else {
			pageW, pageH = a4WidthPt, a4HeightPt
			if w > h { // landscape
				pageW, pageH = a4HeightPt, a4WidthPt
			}
			availW, availH := pageW-2*a4MarginPt, pageH-2*a4MarginPt
			scale := availW / w
			if s := availH / h; s < scale {
				scale = s
			}
			drawW, drawH = w*scale, h*scale
			x, y = (pageW-drawW)/2, (pageH-drawH)/2
		}

		// Page
		begin()
		fmt.Fprintf(&buf,
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /XObject << /Im0 %d 0 R >> >> /Contents %d 0 R >>\nendobj\n",
			pageW, pageH, imageNum, contentNum)

		// Content stream
		content := fmt.Sprintf("q %.2f 0 0 %.2f %.2f %.2f cm /Im0 Do Q", drawW, drawH, x, y)
		begin()
		fmt.Fprintf(&buf, "<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(content), content)

		// Image XObject (JPEG / DCTDecode)
		begin()
		fmt.Fprintf(&buf,
			"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n",
			im.width, im.height, len(im.jpegData))
		buf.Write(im.jpegData)
		buf.WriteString("\nendstream\nendobj\n")
	}

	// Cross-reference table (tiap entri tepat 20 byte)
	xrefPos := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(offsets)+1)
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xrefPos)

	return buf.Bytes(), nil
}

// PDFToImages merender tiap halaman PDF ke gambar memakai pdftoppm (poppler-utils).
// Hasil: outputDir/page-1.png, page-2.png, ... (atau .jpg)
func (s *pdfImageService) PDFToImages(inputPath, outputDir, format string, dpi int) error {
	if dpi < 72 || dpi > 300 {
		dpi = 150
	}

	formatFlag := "-png"
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "jpeg", "jpg":
		formatFlag = "-jpeg"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{formatFlag, "-r", strconv.Itoa(dpi), inputPath, filepath.Join(outputDir, "page")}
	out, err := exec.CommandContext(ctx, "pdftoppm", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("gagal mengubah PDF menjadi gambar (pdftoppm): %w: %s", err, strings.TrimSpace(string(out)))
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil || len(entries) == 0 {
		return errors.New("tidak ada gambar yang dihasilkan dari PDF")
	}
	return nil
}
