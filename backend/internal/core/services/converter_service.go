package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"toolkitz/backend/internal/core/ports"
)

type converterService struct{}

func NewConverterService() ports.ConverterService {
	return &converterService{}
}

// runSoffice menjalankan LibreOffice headless dan mengembalikan error yang
// menyertakan output soffice, sehingga penyebab gagal terlihat (bukan hanya "exit status 1").
//
// Profil LibreOffice dibuat di dalam outputDir (folder temp per-request). Tanpa ini semua
// request berbagi satu profil, dan request kedua yang berjalan bersamaan akan gagal
// atau hang karena profil sedang terkunci.
func runSoffice(outputDir string, args ...string) error {
	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("path output tidak valid: %w", err)
	}

	profile := "-env:UserInstallation=file://" + filepath.ToSlash(filepath.Join(absDir, ".lo-profile"))
	fullArgs := append([]string{profile, "--headless"}, args...)

	out, err := exec.Command("soffice", fullArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("soffice gagal: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *converterService) WordToPDF(inputPath, outputDir string) (string, error) {
	// eksekusi: soffice --headless --convert-to pdf --outdir [dir] [file]
	if err := runSoffice(outputDir, "--convert-to", "pdf", "--outdir", outputDir, inputPath); err != nil {
		return "", fmt.Errorf("gagal konversi ke PDF. Pastikan LibreOffice terinstall: %w", err)
	}

	// LibreOffice akan membuat file dengan nama yang sama tapi ekstensi .pdf
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(outputDir, baseName+".pdf")
	if _, err := os.Stat(outputPath); err != nil {
		return "", fmt.Errorf("konversi selesai tetapi file hasil tidak ditemukan: %w", err)
	}
	return outputPath, nil
}

func (s *converterService) PDFToWord(inputPath, outputDir string) (string, error) {
	// Menggunakan infilter khusus agar PDF dibaca sebagai dokumen teks
	if err := runSoffice(outputDir, "--infilter=writer_pdf_import", "--convert-to", "docx", "--outdir", outputDir, inputPath); err != nil {
		return "", fmt.Errorf("gagal konversi ke Word: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(outputDir, baseName+".docx")
	if _, err := os.Stat(outputPath); err != nil {
		return "", fmt.Errorf("konversi selesai tetapi file hasil tidak ditemukan: %w", err)
	}
	return outputPath, nil
}
