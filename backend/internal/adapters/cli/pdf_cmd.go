package cli

import (
	"fmt"
	"os"

	"toolkitz/backend/internal/core/services"

	"github.com/spf13/cobra"
)

func AddPDFCommands(rootCmd *cobra.Command) {
	pdfSvc := services.NewPDFService()

	// ==========================================
	// 1. COMMAND: PENGGABUNGAN & PEMISAHAN
	// ==========================================

	var mergeCmd = &cobra.Command{
		Use:   "merge [output.pdf] [input1.pdf] [input2.pdf]...",
		Short: "Merge beberapa file PDF menjadi satu file",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputPath := args[0]
			inputPaths := args[1:]
			
			fmt.Println("⌛︎ Sedang menggabungkan PDF...")
			if err := pdfSvc.Merge(inputPaths, outputPath); err != nil {
				return err
			}
			
			fmt.Printf("✔ Berhasil menggabungkan %d file ke %s\n", len(inputPaths), outputPath)
			return nil
		},
	}

	var splitCmd = &cobra.Command{
		Use:   "split [input.pdf] [output_dir]",
		Short: "Pisahkan PDF menjadi halaman tunggal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]
			outDir := args[1]
			
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return fmt.Errorf("gagal membuat direktori output: %w", err)
			}
			
			fmt.Println("⌛︎ Sedang memisahkan PDF...")
			if err := pdfSvc.Split(inputPath, outDir, 1); err != nil {
				return err
			}
			
			fmt.Printf("✔ Berhasil memisahkan PDF %s ke folder %s\n", inputPath, outDir)
			return nil
		},
	}

	// ==========================================
	// 2. COMMAND: KEAMANAN (PASSWORD)
	// ==========================================

	var password string

	var encryptCmd = &cobra.Command{
		Use:   "encrypt [input.pdf] [output.pdf]",
		Short: "Menambahkan password ke PDF",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if password == "" {
				return fmt.Errorf("flag --password wajib diisi (contoh: -p rahasia123)")
			}
			fmt.Println("⌛︎ Sedang mengunci PDF...")
			if err := pdfSvc.Encrypt(args[0], args[1], password); err != nil {
				return err
			}
			fmt.Println("🔒︎ PDF berhasil dikunci!")
			return nil
		},
	}
	encryptCmd.Flags().StringVarP(&password, "password", "p", "", "Password untuk PDF")

	var decryptCmd = &cobra.Command{
		Use:   "decrypt [input.pdf] [output.pdf]",
		Short: "Menghapus password dari PDF",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if password == "" {
				return fmt.Errorf("flag --password wajib diisi (contoh: -p rahasia123)")
			}
			fmt.Println("⌛︎ Sedang membuka kunci PDF...")
			if err := pdfSvc.Decrypt(args[0], args[1], password); err != nil {
				return err
			}
			fmt.Println("ꗃ Password PDF berhasil dihapus!")
			return nil
		},
	}
	decryptCmd.Flags().StringVarP(&password, "password", "p", "", "Password lama PDF")

	// ==========================================
	// 3. COMMAND: MANIPULASI HALAMAN
	// ==========================================

	var pages string
	var rotation int

	var extractCmd = &cobra.Command{
		Use:   "extract [input.pdf] [output.pdf]",
		Short: "Ekstrak halaman tertentu dari PDF",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if pages == "" {
				return fmt.Errorf("flag --pages wajib diisi (contoh: -p 1-3,5)")
			}
			fmt.Println("⌛︎ Sedang mengekstrak halaman...")
			if err := pdfSvc.Extract(args[0], args[1], pages); err != nil {
				return err
			}
			fmt.Println("✔ Halaman berhasil diekstrak!")
			return nil
		},
	}
	extractCmd.Flags().StringVarP(&pages, "pages", "p", "", "Halaman (contoh: 1,3-5) [Wajib]")

	var removeCmd = &cobra.Command{
		Use:   "remove [input.pdf] [output.pdf]",
		Short: "Hapus halaman tertentu dari PDF",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if pages == "" {
				return fmt.Errorf("flag --pages wajib diisi (contoh: -p 1-3,5)")
			}
			fmt.Println("⌛︎ Sedang menghapus halaman...")
			if err := pdfSvc.Remove(args[0], args[1], pages); err != nil {
				return err
			}
			fmt.Println("✔ Halaman berhasil dihapus!")
			return nil
		},
	}
	removeCmd.Flags().StringVarP(&pages, "pages", "p", "", "Halaman yang dihapus (contoh: 1,3-5) [Wajib]")

	var rotateCmd = &cobra.Command{
		Use:   "rotate [input.pdf] [output.pdf]",
		Short: "Putar halaman PDF (90, 180, 270)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⌛︎ Sedang memutar halaman...")
			if err := pdfSvc.Rotate(args[0], args[1], pages, rotation); err != nil {
				return err
			}
			fmt.Println("🗘 Halaman berhasil diputar!")
			return nil
		},
	}
	rotateCmd.Flags().StringVarP(&pages, "pages", "p", "", "Halaman yang diputar (kosongkan untuk memutar semua)")
	rotateCmd.Flags().IntVarP(&rotation, "rotation", "r", 90, "Derajat putaran (90, 180, 270)")

	// ==========================================
	// 4. COMMAND: OPTIMALISASI / KOMPRESI
	// ==========================================

	var compressLevel string

	var compressCmd = &cobra.Command{
		Use:   "compress [input.pdf] [output.pdf]",
		Short: "Mengompresi dan mengoptimalkan ukuran file PDF",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⌛︎ Mengompresi PDF...")
			if err := pdfSvc.CompressWithLevel(args[0], args[1], compressLevel); err != nil {
				return err
			}
			fmt.Println("✔ PDF berhasil dikompresi!")
			return nil
		},
	}
	compressCmd.Flags().StringVarP(&compressLevel, "level", "l", "medium", "Tingkat kompresi: low, medium, high (butuh Ghostscript untuk hasil maksimal)")

	// ==========================================
	// 5. WATERMARK
	// ==========================================
	var watermarkText string

	var watermarkCmd = &cobra.Command{
		Use:   "watermark [input.pdf] [output.pdf]",
		Short: "Tambahkan teks watermark ke seluruh halaman PDF",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if watermarkText == "" {
				return fmt.Errorf("flag --text wajib diisi")
			}
			fmt.Println("⌛︎ Menambahkan watermark...")
			if err := pdfSvc.AddWatermark(args[0], args[1], watermarkText); err != nil {
				return err
			}
			fmt.Println("✔ Watermark berhasil ditambahkan!")
			return nil
		},
	}
	watermarkCmd.Flags().StringVarP(&watermarkText, "text", "t", "", "Teks watermark")

	var pageNumCmd = &cobra.Command{
		Use:   "pagenumber [input.pdf] [output.pdf]",
		Short: "Tambahkan nomor halaman (Page X of Y) di bawah tengah dokumen",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⌛︎ Menambahkan nomor halaman...")
			if err := pdfSvc.AddPageNumbers(args[0], args[1]); err != nil {
				return err
			}
			fmt.Println("✔ Nomor halaman berhasil ditambahkan!")
			return nil
		},
	}

	// ==========================================
	// DAFTARKAN SEMUA KE ROOT COMMAND
	// ==========================================
	
	rootCmd.AddCommand(
		mergeCmd, 
		splitCmd, 
		encryptCmd, 
		decryptCmd, 
		extractCmd, 
		removeCmd, 
		rotateCmd, 
		compressCmd,
		watermarkCmd,
		pageNumCmd,
	)
}