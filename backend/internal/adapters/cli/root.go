package cli

import (
	"fmt"
	"os"

	"toolkitz/backend/internal/core/ports"
	"toolkitz/backend/internal/core/services"

	"github.com/spf13/cobra"
)

var (
	width  int
	height int
	format string
)

const banner = "\033[36m" + `
████████╗  ██████╗   ██████╗  ██╗      ██╗  ██╗ ██╗ ████████╗ ███████╗
╚══██╔══╝ ██╔═══██╗ ██╔═══██╗ ██║      ██║ ██╔╝ ██║ ╚══██╔══╝ ╚════██║
   ██║    ██║   ██║ ██║   ██║ ██║      █████╔╝  ██║    ██║      ███╔═╝
   ██║    ██║   ██║ ██║   ██║ ██║      ██╔═██╗  ██║    ██║    ██╔══╝  
   ██║    ╚██████╔╝ ╚██████╔╝ ███████╗ ██║  ██╗ ██║    ██║    ███████╗
   ╚═╝     ╚═════╝   ╚═════╝  ╚══════╝ ╚═╝  ╚═╝ ╚═╝    ╚═╝    ╚══════╝` + "\033[0m"

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "pdf-tool",
		Short: "Toolkit lintas platform untuk pemrosesan dokumen dan gambar",
		// 'Long' to see ./pdf-tool --help
		Long: fmt.Sprintf("%s\n\nToolkitz CLI memungkinkan Anda memproses PDF dan Gambar langsung dari terminal.\n", banner),
		
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(banner)
			fmt.Print("\033[90mCross-Platform Document & Image Processing CLI\033[0m\n\n")
			
			fmt.Println("Available Commands (Top 5):")
			fmt.Println("  merge       Gabungkan beberapa PDF menjadi satu")
			fmt.Println("  split       Pisahkan PDF menjadi halaman-halaman tunggal")
			fmt.Println("  compress    Kompresi dan optimalkan ukuran PDF")
			fmt.Println("  word2pdf    Konversi dokumen Word (.docx) ke PDF")
			fmt.Println("  resize      Ubah ukuran, crop, dan format gambar")
			
			fmt.Println("\n\033[33mKetik 'pdf-tool --help' untuk melihat seluruh command yang tersedia.\033[0m")
		},
	}

	// 1. Registrasi Command PDF
	AddPDFCommands(rootCmd)

	// 2. Registrasi Command Konversi
	AddConvertCommands(rootCmd)

	// 2b. Registrasi Command Gambar <-> PDF
	AddImagePDFCommands(rootCmd)

	// 3. Command Gambar (Resize)
	imgSvc := services.NewImageService()
	var resizeCmd = &cobra.Command{
		Use:   "resize [input] [output]",
		Short: "Ubah ukuran dan format gambar",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]
			outputPath := args[1]

			in, err := os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("gagal membuka file input: %w", err)
			}
			defer in.Close()

			out, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("gagal membuat file output: %w", err)
			}
			defer out.Close()

			opts := ports.ImageOptions{
				Width: width, Height: height, Format: format,
				Quality: 85,
			}
			
			fmt.Println("⌛︎ Memproses gambar...")
			if err := imgSvc.ProcessImage(in, out, opts); err != nil {
				return err
			}

			fmt.Printf("✔ Gambar berhasil diproses: %s\n", outputPath)
			return nil
		},
	}

	resizeCmd.Flags().IntVarP(&width, "width", "W", 0, "Target lebar")
	resizeCmd.Flags().IntVarP(&height, "height", "H", 0, "Target tinggi")
	resizeCmd.Flags().StringVarP(&format, "format", "f", "jpeg", "Format output (jpeg, png, webp)")
	
	rootCmd.AddCommand(resizeCmd)

	// CLI
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}