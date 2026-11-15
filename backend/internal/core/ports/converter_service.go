package ports

type ConverterService interface {
	// Mengembalikan path file hasil konversi
	WordToPDF(inputPath, outputDir string) (string, error)
	// OfficeToPDF: Word/Excel/PowerPoint -> PDF (format dikenali dari ekstensi file)
	OfficeToPDF(inputPath, outputDir string) (string, error)
	PDFToWord(inputPath, outputDir string) (string, error)
}
