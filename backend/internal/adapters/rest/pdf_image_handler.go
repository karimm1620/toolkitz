package rest

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

const maxHEICFiles = 50

// HEICToJPEG: field "images" (satu atau banyak file HEIC/HEIF), opsional "quality" (1-100, default 90).
// Satu file -> JPG langsung; banyak file -> ZIP berisi JPG dengan nama asli.
func (h *PDFImageHandler) HEICToJPEG(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid form"})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Minimal 1 file HEIC dibutuhkan"})
	}
	if len(files) > maxHEICFiles {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Maksimal %d file sekali proses", maxHEICFiles)})
	}

	quality, _ := strconv.Atoi(c.FormValue("quality", "90"))

	dir, err := newWorkDir("heic2jpg")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(dir)

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder output"})
	}

	used := map[string]int{}
	var outPaths, outNames []string
	for i, fh := range files {
		orig := safeBase(fh.Filename, fmt.Sprintf("foto_%d.heic", i+1))
		in := filepath.Join(dir, fmt.Sprintf("in_%03d.heic", i))
		if err := c.SaveFile(fh, in); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
		}

		// Nama keluaran = nama asli + .jpg; beri akhiran bila ada nama kembar
		name := strings.TrimSuffix(orig, filepath.Ext(orig))
		used[name]++
		if used[name] > 1 {
			name = fmt.Sprintf("%s_%d", name, used[name])
		}
		outName := name + ".jpg"
		outPath := filepath.Join(outDir, fmt.Sprintf("out_%03d.jpg", i))

		if err := h.svc.HEICToJPEG(in, outPath, quality); err != nil {
			return c.Status(422).JSON(fiber.Map{"error": fmt.Sprintf("%s: %s", orig, err.Error())})
		}
		outPaths = append(outPaths, outPath)
		outNames = append(outNames, outName)
	}

	if len(outPaths) == 1 {
		return sendDownload(c, outPaths[0], outNames[0], "image/jpeg")
	}

	// Ganti nama berkas hasil ke nama asli sebelum di-ZIP
	for i := range outPaths {
		if err := os.Rename(outPaths[i], filepath.Join(outDir, outNames[i])); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyiapkan hasil"})
		}
	}
	zipPath := filepath.Join(dir, "heic_to_jpg.zip")
	if err := utils.ZipDirectory(outDir, zipPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat ZIP"})
	}
	return sendDownload(c, zipPath, "heic_to_jpg.zip", "application/zip")
}
