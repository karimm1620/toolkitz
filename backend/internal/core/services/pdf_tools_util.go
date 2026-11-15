package services

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// runTool menjalankan program eksternal dengan batas waktu dan mengembalikan error
// yang menyertakan output program, supaya penyebab gagal terlihat.
func runTool(timeout time.Duration, env []string, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = env
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("%s melewati batas waktu %s", name, timeout)
		}
		return fmt.Errorf("%s gagal: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// runQpdf menjalankan qpdf. Exit code 3 berarti "selesai dengan peringatan"
// (file hasil tetap dibuat), jadi dianggap berhasil.
func runQpdf(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	out, err := exec.CommandContext(ctx, "qpdf", args...).CombinedOutput()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 3 {
			return nil
		}
		return fmt.Errorf("qpdf gagal: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// pdfPageCount mengembalikan jumlah halaman PDF.
func pdfPageCount(path string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	out, err := exec.CommandContext(ctx, "qpdf", "--show-npages", path).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("tidak bisa membaca PDF (rusak atau terkunci password?): %s", strings.TrimSpace(string(out)))
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || n < 1 {
		return 0, errors.New("jumlah halaman PDF tidak valid")
	}
	return n, nil
}

// pageGeom adalah ukuran halaman (poin) dan rotasi bawaannya.
type pageGeom struct {
	Width  float64
	Height float64
	Rot    int
}

var (
	reSize = regexp.MustCompile(`(?m)^Page\s+(\d+)\s+size:\s+([\d.]+)\s+x\s+([\d.]+)\s+pts`)
	reRot  = regexp.MustCompile(`(?m)^Page\s+(\d+)\s+rot:\s+(\d+)`)
)

// pdfPageGeoms membaca ukuran & rotasi tiap halaman memakai pdfinfo (poppler-utils).
// Hasil diindeks dari halaman 1 (index 0).
func pdfPageGeoms(path string, total int) ([]pageGeom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	out, err := exec.CommandContext(ctx, "pdfinfo", "-f", "1", "-l", strconv.Itoa(total), path).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pdfinfo gagal: %w: %s", err, strings.TrimSpace(string(out)))
	}
	text := string(out)

	geoms := make([]pageGeom, total)
	found := 0
	for _, m := range reSize.FindAllStringSubmatch(text, -1) {
		n, _ := strconv.Atoi(m[1])
		w, _ := strconv.ParseFloat(m[2], 64)
		h, _ := strconv.ParseFloat(m[3], 64)
		if n >= 1 && n <= total {
			geoms[n-1].Width, geoms[n-1].Height = w, h
			found++
		}
	}
	if found != total {
		return nil, fmt.Errorf("ukuran halaman tidak lengkap (%d dari %d)", found, total)
	}
	for _, m := range reRot.FindAllStringSubmatch(text, -1) {
		n, _ := strconv.Atoi(m[1])
		r, _ := strconv.Atoi(m[2])
		if n >= 1 && n <= total {
			geoms[n-1].Rot = r
		}
	}
	return geoms, nil
}
