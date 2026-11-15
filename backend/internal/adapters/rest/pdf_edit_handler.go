package rest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"toolkitz/backend/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

const maxThumbnailPages = 300

type PDFEditHandler struct {
	svc ports.PDFEditService
}

func NewPDFEditHandler(svc ports.PDFEditService) *PDFEditHandler {
	return &PDFEditHandler{svc: svc}
}

func trimExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// savePDFUpload menyimpan unggahan "pdf" ke folder kerja baru.
// Mengembalikan folder kerja, path input, dan nama asli (aman).
// Pemanggil wajib `defer os.RemoveAll(dir)` jika err == nil.
func savePDFUpload(c *fiber.Ctx, prefix string) (dir, inPath, origName string, status int, err error) {
	fileHeader, ferr := c.FormFile("pdf")
	if ferr != nil {
		return "", "", "", 400, fmt.Errorf("File PDF dibutuhkan")
	}
	d, derr := newWorkDir(prefix)
	if derr != nil {
		return "", "", "", 500, fmt.Errorf("Gagal membuat folder sementara")
	}
	in := filepath.Join(d, "input.pdf")
	if serr := c.SaveFile(fileHeader, in); serr != nil {
		os.RemoveAll(d)
		return "", "", "", 500, fmt.Errorf("Gagal menyimpan file")
	}
	return d, in, safeBase(fileHeader.Filename, "document.pdf"), 0, nil
}

// OCR: field "pdf", "lang" (default "eng", boleh "ind+eng"), "output" = "pdf" (default) | "txt".
func (h *PDFEditHandler) OCR(c *fiber.Ctx) error {
	lang := c.FormValue("lang", "eng")
	output := strings.ToLower(c.FormValue("output", "pdf"))
	if output != "txt" {
		output = "pdf"
	}

	dir, in, name, status, err := savePDFUpload(c, "pdf-ocr")
	if err != nil {
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	defer os.RemoveAll(dir)

	out := filepath.Join(dir, "ocr_output."+output)
	if err := h.svc.OCR(in, out, lang, output); err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}

	if output == "txt" {
		return sendDownload(c, out, trimExt(name)+"_ocr.txt", "text/plain; charset=utf-8")
	}
	return sendDownload(c, out, trimExt(name)+"_ocr.pdf", "application/pdf")
}

// Sign: field "pdf", "signature" (gambar), opsional "pages", "position", "width", "remove_white".
func (h *PDFEditHandler) Sign(c *fiber.Ctx) error {
	sigHeader, err := c.FormFile("signature")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Gambar tanda tangan dibutuhkan"})
	}

	width, _ := strconv.Atoi(c.FormValue("width", "25"))
	opts := ports.SignOptions{
		Pages:        c.FormValue("pages", "last"),
		Position:     c.FormValue("position", "br"),
		WidthPercent: width,
		RemoveWhite:  c.FormValue("remove_white", "true") != "false",
	}

	dir, in, name, status, err := savePDFUpload(c, "pdf-sign")
	if err != nil {
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	defer os.RemoveAll(dir)

	sigPath := filepath.Join(dir, "signature")
	if err := c.SaveFile(sigHeader, sigPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan gambar tanda tangan"})
	}

	out := filepath.Join(dir, "signed.pdf")
	if err := h.svc.Sign(in, sigPath, out, opts); err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return sendDownload(c, out, "signed_"+name, "application/pdf")
}

// Thumbnails: field "pdf". Balasan JSON: {"total": N, "thumbnails": ["data:image/jpeg;base64,...", ...]}.
func (h *PDFEditHandler) Thumbnails(c *fiber.Ctx) error {
	dir, in, _, status, err := savePDFUpload(c, "pdf-thumbs")
	if err != nil {
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	defer os.RemoveAll(dir)

	thumbDir := filepath.Join(dir, "thumbs")
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat folder output"})
	}

	files, err := h.svc.Thumbnails(in, thumbDir, maxThumbnailPages)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}

	uris := make([]string, 0, len(files))
	for _, f := range files {
		data, rerr := os.ReadFile(f)
		if rerr != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca thumbnail"})
		}
		uris = append(uris, "data:image/jpeg;base64,"+base64.StdEncoding.EncodeToString(data))
	}
	return c.JSON(fiber.Map{"total": len(uris), "thumbnails": uris})
}

// Reorder: field "pdf" dan "order" = JSON, contoh [{"page":3,"rotate":90},{"page":1,"rotate":0}].
// Halaman yang tidak disebut dihapus; halaman yang disebut dua kali digandakan.
func (h *PDFEditHandler) Reorder(c *fiber.Ctx) error {
	var pages []ports.PageSpec
	if err := json.Unmarshal([]byte(c.FormValue("order")), &pages); err != nil || len(pages) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Urutan halaman tidak valid"})
	}

	dir, in, name, status, err := savePDFUpload(c, "pdf-reorder")
	if err != nil {
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	defer os.RemoveAll(dir)

	out := filepath.Join(dir, "reordered.pdf")
	if err := h.svc.Reorder(in, out, pages); err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return sendDownload(c, out, "reordered_"+name, "application/pdf")
}
