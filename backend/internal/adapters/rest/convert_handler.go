package rest

import (
	"os"
	"path/filepath"

	"toolkitz/backend/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type ConvertHandler struct {
	svc ports.ConverterService
}

func NewConvertHandler(svc ports.ConverterService) *ConvertHandler {
	return &ConvertHandler{svc: svc}
}

func (h *ConvertHandler) handleConversion(c *fiber.Ctx, expectedExt, targetExt string, convertFunc func(string, string) (string, error)) error {
	fileHeader, err := c.FormFile("document")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File dokumen tidak ditemukan"})
	}

	// Setup direktori temporary
	tempDir, err := os.MkdirTemp("", "convert-*")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder sementara"})
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, "input"+expectedExt)
	if err := c.SaveFile(fileHeader, inputPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}

	// Proses konversi
	outputPath, err := convertFunc(inputPath, tempDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Rename file output
	finalName := fileHeader.Filename[:len(fileHeader.Filename)-len(filepath.Ext(fileHeader.Filename))] + targetExt
	c.Set("Content-Disposition", `attachment; filename="`+finalName+`"`)
	
	return c.SendFile(outputPath)
}

func (h *ConvertHandler) WordToPDF(c *fiber.Ctx) error {
	return h.handleConversion(c, ".docx", ".pdf", h.svc.WordToPDF)
}

func (h *ConvertHandler) ExcelToPDF(c *fiber.Ctx) error {
	return h.handleConversion(c, ".xlsx", ".pdf", h.svc.OfficeToPDF)
}

func (h *ConvertHandler) PPTToPDF(c *fiber.Ctx) error {
	return h.handleConversion(c, ".pptx", ".pdf", h.svc.OfficeToPDF)
}

func (h *ConvertHandler) PDFToWord(c *fiber.Ctx) error {
	return h.handleConversion(c, ".pdf", ".docx", h.svc.PDFToWord)
}