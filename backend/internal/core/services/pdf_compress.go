package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ghostscriptPreset memetakan level kompresi ke preset Ghostscript.
//
//	low    -> /printer (300 dpi): kualitas tinggi, penghematan kecil
//	medium -> /ebook   (150 dpi): seimbang (default)
//	high   -> /screen  (72 dpi) : file terkecil, kualitas turun
func ghostscriptPreset(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "low":
		return "/printer"
	case "high":
		return "/screen"
	default:
		return "/ebook"
	}
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CompressWithLevel mengompres PDF dengan Ghostscript.
// Kalau gs tidak terpasang, gagal, atau hasilnya justru tidak lebih kecil dari aslinya,
// dipakai Compress (optimasi struktur pdfcpu) sebagai cadangan.
func (s *pdfService) CompressWithLevel(inputPath, outputPath, level string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.5",
		"-dPDFSETTINGS=" + ghostscriptPreset(level),
		"-dNOPAUSE",
		"-dBATCH",
		"-dQUIET",
		"-sOutputFile=" + outputPath,
		inputPath,
	}

	gsErr := exec.CommandContext(ctx, "gs", args...).Run()
	if gsErr == nil {
		inSize, errIn := fileSize(inputPath)
		outSize, errOut := fileSize(outputPath)
		if errIn == nil && errOut == nil && outSize > 0 && outSize < inSize {
			return nil
		}
	}

	// Cadangan: optimasi struktur (tanpa menurunkan kualitas gambar)
	if err := s.Compress(inputPath, outputPath); err != nil {
		if gsErr != nil {
			return fmt.Errorf("gagal mengompres PDF (ghostscript: %v; optimasi: %w)", gsErr, err)
		}
		return err
	}
	return nil
}
