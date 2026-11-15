package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"toolkitz/backend/internal/core/ports"
)

const maxOCRPages = 100

type pdfEditService struct{}

func NewPDFEditService() ports.PDFEditService {
	return &pdfEditService{}
}

// Kode bahasa Tesseract, mis. "eng", "ind", "chi_sim"; boleh digabung "ind+eng".
var reLang = regexp.MustCompile(`^[a-z][a-z_]{1,10}(\+[a-z][a-z_]{1,10}){0,2}$`)

// OCR merender tiap halaman ke gambar (200 dpi), lalu menjalankan Tesseract
// untuk semua halaman sekaligus. mode "pdf" -> PDF yang bisa dicari; "txt" -> teks biasa.
func (s *pdfEditService) OCR(inputPath, outputPath, lang, mode string) error {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = "eng"
	}
	if !reLang.MatchString(lang) {
		return errors.New("kode bahasa OCR tidak valid")
	}
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "txt" {
		mode = "pdf"
	}

	total, err := pdfPageCount(inputPath)
	if err != nil {
		return err
	}
	if total > maxOCRPages {
		return fmt.Errorf("OCR dibatasi %d halaman per proses (file ini %d halaman); pisahkan dulu PDF-nya", maxOCRPages, total)
	}

	work, err := os.MkdirTemp(filepath.Dir(outputPath), "ocr-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	// 1) PDF -> gambar halaman
	if err := runTool(5*time.Minute, nil, "pdftoppm", "-png", "-r", "200", inputPath, filepath.Join(work, "p")); err != nil {
		return fmt.Errorf("gagal merender halaman untuk OCR: %w", err)
	}
	images, err := filepath.Glob(filepath.Join(work, "p-*.png"))
	if err != nil || len(images) == 0 {
		return errors.New("tidak ada halaman yang bisa dirender untuk OCR")
	}
	sort.Strings(images) // pdftoppm memberi nomor dengan zero-padding

	// 2) daftar gambar -> satu panggilan Tesseract (menghasilkan satu dokumen multi-halaman)
	listPath := filepath.Join(work, "images.txt")
	if err := os.WriteFile(listPath, []byte(strings.Join(images, "\n")+"\n"), 0o644); err != nil {
		return err
	}

	outBase := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))

	// OMP_THREAD_LIMIT=1: di container, thread OpenMP Tesseract saling berebut CPU dan malah lebih lambat.
	env := append(os.Environ(), "OMP_THREAD_LIMIT=1")
	if err := runTool(10*time.Minute, env, "tesseract", listPath, outBase, "-l", lang, mode); err != nil {
		return fmt.Errorf("OCR gagal (pastikan paket bahasa %q terpasang): %w", lang, err)
	}

	produced := outBase + "." + mode
	if _, err := os.Stat(produced); err != nil {
		return fmt.Errorf("OCR selesai tetapi file hasil tidak ditemukan: %w", err)
	}
	if produced != outputPath {
		return os.Rename(produced, outputPath)
	}
	return nil
}

// Thumbnails merender tiap halaman menjadi JPEG kecil (sisi terpanjang 240 px).
func (s *pdfEditService) Thumbnails(inputPath, outputDir string, maxPages int) ([]string, error) {
	total, err := pdfPageCount(inputPath)
	if err != nil {
		return nil, err
	}
	if total > maxPages {
		return nil, fmt.Errorf("PDF terlalu besar untuk fitur ini (%d halaman, maksimal %d)", total, maxPages)
	}

	args := []string{"-jpeg", "-jpegopt", "quality=70", "-scale-to", "240", inputPath, filepath.Join(outputDir, "t")}
	if err := runTool(3*time.Minute, nil, "pdftoppm", args...); err != nil {
		return nil, fmt.Errorf("gagal membuat thumbnail: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "t-*.jpg"))
	if err != nil || len(files) != total {
		return nil, fmt.Errorf("thumbnail tidak lengkap (%d dari %d halaman)", len(files), total)
	}
	sort.Strings(files)
	return files, nil
}

// Reorder menyusun ulang halaman dengan qpdf:
//  1. --pages menyalin halaman sesuai urutan (boleh dihapus/digandakan),
//  2. --rotate memutar halaman hasil (nomor sudah berdasarkan urutan baru).
func (s *pdfEditService) Reorder(inputPath, outputPath string, pages []ports.PageSpec) error {
	if len(pages) == 0 {
		return errors.New("minimal satu halaman harus dipertahankan")
	}
	if len(pages) > 1000 {
		return errors.New("terlalu banyak halaman (maksimal 1000)")
	}

	total, err := pdfPageCount(inputPath)
	if err != nil {
		return err
	}

	order := make([]string, 0, len(pages))
	rotations := map[int][]string{} // sudut -> nomor halaman hasil
	for i, p := range pages {
		if p.Page < 1 || p.Page > total {
			return fmt.Errorf("halaman %d tidak ada (PDF punya %d halaman)", p.Page, total)
		}
		if p.Rotate%90 != 0 || p.Rotate < 0 || p.Rotate >= 360 {
			return fmt.Errorf("rotasi %d tidak valid (gunakan 0, 90, 180, atau 270)", p.Rotate)
		}
		order = append(order, strconv.Itoa(p.Page))
		if p.Rotate != 0 {
			rotations[p.Rotate] = append(rotations[p.Rotate], strconv.Itoa(i+1))
		}
	}

	stage1 := outputPath
	if len(rotations) > 0 {
		stage1 = outputPath + ".stage1.pdf"
		defer os.Remove(stage1)
	}

	if err := runQpdf(inputPath, "--pages", ".", strings.Join(order, ","), "--", stage1); err != nil {
		return err
	}
	if len(rotations) == 0 {
		return nil
	}

	args := []string{stage1}
	angles := make([]int, 0, len(rotations))
	for a := range rotations {
		angles = append(angles, a)
	}
	sort.Ints(angles)
	for _, a := range angles {
		args = append(args, fmt.Sprintf("--rotate=+%d:%s", a, strings.Join(rotations[a], ",")))
	}
	args = append(args, "--", outputPath)
	return runQpdf(args...)
}
