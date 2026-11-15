package rest

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"toolkitz/backend/internal/core/ports"
	"toolkitz/backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type PDFHandler struct {
	svc ports.PDFService
}

func NewPDFHandler(svc ports.PDFService) *PDFHandler {
	return &PDFHandler{svc: svc}
}

func (h *PDFHandler) Merge(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid form"})
	}

	files := form.File["pdfs"]
	if len(files) < 2 {
		return c.Status(400).JSON(fiber.Map{"error": "butuh minimal 2 PDF"})
	}

	dir, err := newWorkDir("pdf-merge")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	// Simpan file sementara ke disk (urutan upload = urutan penggabungan)
	var inputPaths []string
	for i, fileHeader := range files {
		tempPath := filepath.Join(dir, fmt.Sprintf("upload_%03d.pdf", i))
		if err := c.SaveFile(fileHeader, tempPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
		}
		inputPaths = append(inputPaths, tempPath)
	}

	tempOut := filepath.Join(dir, "merged_output.pdf")
	if err := h.svc.Merge(inputPaths, tempOut); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return sendDownload(c, tempOut, "merged_output.pdf", "application/pdf")
}

func (h *PDFHandler) Split(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "pdf file is required"})
	}

	dir, err := newWorkDir("pdf-split")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create temp dir"})
	}
	defer os.RemoveAll(dir)

	tempIn := filepath.Join(dir, "input.pdf")
	if err := c.SaveFile(fileHeader, tempIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save file"})
	}

	outDir := filepath.Join(dir, "pages")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create output dir"})
	}

	if err := h.svc.Split(tempIn, outDir, 1); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// ZIP ditaruh di luar outDir supaya tidak ikut terkompres ke dirinya sendiri
	zipPath := filepath.Join(dir, "split_pages.zip")
	if err := utils.ZipDirectory(outDir, zipPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to zip"})
	}

	return sendDownload(c, zipPath, "split_pages.zip", "application/zip")
}

// Security
func (h *PDFHandler) handleSecurity(c *fiber.Ctx, action func(string, string, string) error) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File PDF dibutuhkan"})
	}

	password := c.FormValue("password")
	if password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Password tidak boleh kosong"})
	}

	dir, err := newWorkDir("pdf-security")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	tempIn := filepath.Join(dir, "input.pdf")
	if err := c.SaveFile(fileHeader, tempIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}

	tempOut := filepath.Join(dir, "output.pdf")
	if err := action(tempIn, tempOut, password); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	name := safeBase(fileHeader.Filename, "document.pdf")
	return sendDownload(c, tempOut, "secure_"+name, "application/pdf")
}

func (h *PDFHandler) Encrypt(c *fiber.Ctx) error {
	return h.handleSecurity(c, h.svc.Encrypt)
}

func (h *PDFHandler) Decrypt(c *fiber.Ctx) error {
	return h.handleSecurity(c, h.svc.Decrypt)
}

// Manipulasi page
func (h *PDFHandler) ManipulatePages(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File PDF dibutuhkan"})
	}

	action := c.FormValue("action") // "extract", "remove", atau "rotate"
	pages := c.FormValue("pages")   // contoh: "1-3, 5"

	if pages == "" && action != "rotate" {
		return c.Status(400).JSON(fiber.Map{"error": "Pilihan halaman tidak boleh kosong"})
	}

	dir, err := newWorkDir("pdf-pages")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	tempIn := filepath.Join(dir, "input.pdf")
	if err := c.SaveFile(fileHeader, tempIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}
	tempOut := filepath.Join(dir, "output.pdf")

	var procErr error
	switch action {
	case "extract":
		procErr = h.svc.Extract(tempIn, tempOut, pages)
	case "remove":
		procErr = h.svc.Remove(tempIn, tempOut, pages)
	case "rotate":
		rot, _ := strconv.Atoi(c.FormValue("rotation", "90"))
		procErr = h.svc.Rotate(tempIn, tempOut, pages, rot)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Aksi tidak valid"})
	}

	if procErr != nil {
		return c.Status(500).JSON(fiber.Map{"error": procErr.Error()})
	}

	name := safeBase(fileHeader.Filename, "document.pdf")
	return sendDownload(c, tempOut, "pages_"+name, "application/pdf")
}

// Compress
func (h *PDFHandler) Compress(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File PDF dibutuhkan"})
	}

	// "low" | "medium" | "high" (default medium)
	level := c.FormValue("level", "medium")

	dir, err := newWorkDir("pdf-compress")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	tempIn := filepath.Join(dir, "input.pdf")
	if err := c.SaveFile(fileHeader, tempIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}
	tempOut := filepath.Join(dir, "output.pdf")

	if err := h.svc.CompressWithLevel(tempIn, tempOut, level); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	name := safeBase(fileHeader.Filename, "document.pdf")
	return sendDownload(c, tempOut, "compressed_"+name, "application/pdf")
}

// Watermark
func (h *PDFHandler) Watermark(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File PDF dibutuhkan"})
	}

	action := c.FormValue("action") // "watermark" atau "pagenumbers"
	text := c.FormValue("text")

	if action != "pagenumbers" && text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Teks watermark wajib diisi"})
	}

	dir, err := newWorkDir("pdf-watermark")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	tempIn := filepath.Join(dir, "input.pdf")
	if err := c.SaveFile(fileHeader, tempIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}
	tempOut := filepath.Join(dir, "output.pdf")

	var procErr error
	if action == "pagenumbers" {
		procErr = h.svc.AddPageNumbers(tempIn, tempOut)
	} else {
		procErr = h.svc.AddWatermark(tempIn, tempOut, text)
	}

	if procErr != nil {
		return c.Status(500).JSON(fiber.Map{"error": procErr.Error()})
	}

	name := safeBase(fileHeader.Filename, "document.pdf")
	return sendDownload(c, tempOut, "stamped_"+name, "application/pdf")
}
