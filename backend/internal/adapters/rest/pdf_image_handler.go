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

const maxImagesPerPDF = 100

type PDFImageHandler struct {
	svc ports.PDFImageService
}

func NewPDFImageHandler(svc ports.PDFImageService) *PDFImageHandler {
	return &PDFImageHandler{svc: svc}
}

// ImagesToPDF: field "images" (banyak file, urutan upload = urutan halaman),
// opsional "page" = "a4" (default) | "original".
func (h *PDFImageHandler) ImagesToPDF(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid form"})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Minimal 1 gambar dibutuhkan"})
	}
	if len(files) > maxImagesPerPDF {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Maksimal %d gambar sekali proses", maxImagesPerPDF)})
	}

	pageMode := c.FormValue("page", "a4")

	dir, err := newWorkDir("img2pdf")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	var inputPaths []string
	for i, fh := range files {
		p := filepath.Join(dir, fmt.Sprintf("img_%03d", i))
		if err := c.SaveFile(fh, p); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan gambar"})
		}
		inputPaths = append(inputPaths, p)
	}

	out := filepath.Join(dir, "images.pdf")
	if err := h.svc.ImagesToPDF(inputPaths, out, pageMode); err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}

	return sendDownload(c, out, "images.pdf", "application/pdf")
}

// PDFToImages: field "pdf", opsional "format" = "png" (default) | "jpeg", "dpi" = 72-300 (default 150).
// Hasil berupa ZIP berisi satu gambar per halaman.
func (h *PDFImageHandler) PDFToImages(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File PDF dibutuhkan"})
	}

	format := c.FormValue("format", "png")
	dpi, _ := strconv.Atoi(c.FormValue("dpi", "150"))

	dir, err := newWorkDir("pdf2img")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	tempIn := filepath.Join(dir, "input.pdf")
	if err := c.SaveFile(fileHeader, tempIn); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}

	imgDir := filepath.Join(dir, "images")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder output"})
	}

	if err := h.svc.PDFToImages(tempIn, imgDir, format, dpi); err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}

	// ZIP di luar imgDir agar tidak ikut terkompres ke dirinya sendiri
	zipPath := filepath.Join(dir, "pdf_images.zip")
	if err := utils.ZipDirectory(imgDir, zipPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat ZIP"})
	}

	return sendDownload(c, zipPath, "pdf_images.zip", "application/zip")
}
