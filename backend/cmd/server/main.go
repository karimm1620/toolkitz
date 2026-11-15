package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"toolkitz/backend/internal/adapters/rest"
	"toolkitz/backend/internal/core/services"
)

func main() {
	// Inisialisasi Fiber dengan batas ukuran body 100MB untuk file besar
	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024,
	})

	// Mengaktifkan CORS agar frontend React (port 5173) bisa menembak API (port 3000)
	// app.Use(cors.New()) -> ini diganti
	app.Use(cors.New(cors.Config{
	AllowOrigins: "*",
	AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
}))

	// Fitur Gambar
	imgSvc := services.NewImageService()
	imgHandler := rest.NewImageHandler(imgSvc)

	// Fitur PDF (Merge & Split)
	pdfSvc := services.NewPDFService()
	pdfHandler := rest.NewPDFHandler(pdfSvc)

	// Fitur Konversi (Word <-> PDF)
	convSvc := services.NewConverterService()
	convHandler := rest.NewConvertHandler(convSvc)

	api := app.Group("/api")

	// Health check (dipakai healthcheck Docker Compose)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Endpoint Gambar
	api.Post("/image/resize", imgHandler.Resize)

	// Endpoint PDF
	api.Post("/pdf/merge", pdfHandler.Merge)
	api.Post("/pdf/split", pdfHandler.Split)

	// Endpoint Konversi Dokumen
	api.Post("/convert/word-to-pdf", convHandler.WordToPDF)
	api.Post("/convert/pdf-to-word", convHandler.PDFToWord)
	api.Post("/convert/excel-to-pdf", convHandler.ExcelToPDF)
	api.Post("/convert/ppt-to-pdf", convHandler.PPTToPDF)

	// Endpoint Gambar <-> PDF
	imgPdfSvc := services.NewPDFImageService()
	imgPdfHandler := rest.NewPDFImageHandler(imgPdfSvc)
	api.Post("/pdf/from-images", imgPdfHandler.ImagesToPDF)
	api.Post("/pdf/to-images", imgPdfHandler.PDFToImages)

	// Endpoint Encrypt Dokumen
	api.Post("/pdf/encrypt", pdfHandler.Encrypt)
	api.Post("/pdf/decrypt", pdfHandler.Decrypt)

	// Endpoint Manipulasi Page
	api.Post("/pdf/pages", pdfHandler.ManipulatePages)

	// Endpoint PDF Kompres
	api.Post("/pdf/compress", pdfHandler.Compress)

	// Endpoint Watermark
	api.Post("/pdf/watermark", pdfHandler.Watermark)

	// Hormati env PORT (di-set di docker-compose.yml), default 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di port %s", port)
	log.Fatal(app.Listen(":" + port))
}