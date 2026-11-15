package cli

import (
	"fmt"
	"os"

	"toolkitz/backend/internal/core/services"

	"github.com/spf13/cobra"
)

// AddImagePDFCommands mendaftarkan perintah gambar <-> PDF.
func AddImagePDFCommands(rootCmd *cobra.Command) {
	imgPdfSvc := services.NewPDFImageService()

	// ---- COMMAND: Gambar -> PDF ----
	var pageMode string

	var img2pdfCmd = &cobra.Command{
		Use:   "img2pdf [output.pdf] [gambar1] [gambar2]...",
		Short: "Gabungkan gambar (JPG, PNG, GIF, BMP, TIFF) menjadi satu PDF",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputPath := args[0]
			inputPaths := args[1:]

			fmt.Println("⌛︎ Sedang membuat PDF dari gambar...")
			if err := imgPdfSvc.ImagesToPDF(inputPaths, outputPath, pageMode); err != nil {
				return err
			}

			fmt.Printf("✔ Berhasil membuat %s dari %d gambar\n", outputPath, len(inputPaths))
			return nil
		},
	}
	img2pdfCmd.Flags().StringVar(&pageMode, "page", "a4", "Ukuran halaman: a4 atau original")

	// ---- COMMAND: PDF -> Gambar ----
	var imgFormat string
	var imgDPI int

	var pdf2imgCmd = &cobra.Command{
		Use:   "pdf2img [input.pdf] [output_dir]",
		Short: "Ubah tiap halaman PDF menjadi gambar (butuh pdftoppm / poppler-utils)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir := args[1]
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return fmt.Errorf("gagal membuat direktori output: %w", err)
			}

			fmt.Println("⌛︎ Sedang mengubah PDF menjadi gambar...")
			if err := imgPdfSvc.PDFToImages(args[0], outDir, imgFormat, imgDPI); err != nil {
				return err
			}

			fmt.Printf("✔ Berhasil! Gambar tersimpan di folder %s\n", outDir)
			return nil
		},
	}
	pdf2imgCmd.Flags().StringVarP(&imgFormat, "format", "f", "png", "Format gambar: png atau jpeg")
	pdf2imgCmd.Flags().IntVarP(&imgDPI, "dpi", "d", 150, "Resolusi 72-300 dpi")

	rootCmd.AddCommand(img2pdfCmd, pdf2imgCmd)
}
