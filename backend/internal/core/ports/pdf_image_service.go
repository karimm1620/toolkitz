package ports

type PDFImageService interface {
	// ImagesToPDF menyusun gambar (urutan = urutan slice) menjadi satu PDF.
	// pageMode: "a4" (default, gambar dimuat di halaman A4) atau "original" (ukuran halaman = ukuran gambar).
	ImagesToPDF(inputPaths []string, outputPath, pageMode string) error

	// PDFToImages merender tiap halaman PDF menjadi gambar di outputDir.
	// format: "png" (default) atau "jpeg". dpi: 72-300 (default 150).
	PDFToImages(inputPath, outputDir, format string, dpi int) error
}
