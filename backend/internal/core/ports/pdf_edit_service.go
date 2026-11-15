package ports

// PageSpec adalah satu halaman pada hasil "Susun Halaman".
// Page memakai nomor halaman asli (mulai 1); Rotate searah jarum jam: 0, 90, 180, atau 270.
type PageSpec struct {
	Page   int `json:"page"`
	Rotate int `json:"rotate"`
}

// SignOptions mengatur penempatan tanda tangan / stempel gambar di PDF.
type SignOptions struct {
	Pages        string // "first" | "last" (default) | "all" | daftar, contoh "1,3,5-7"
	Position     string // tl tc tr / ml mc mr / bl bc br (default "br")
	WidthPercent int    // lebar gambar dalam % lebar halaman (5-80, default 25)
	RemoveWhite  bool   // jadikan latar putih/terang transparan (untuk foto tanda tangan)
}

type PDFEditService interface {
	// OCR membuat PDF yang teksnya bisa dicari/disalin (mode "pdf") atau berkas teks (mode "txt").
	// lang: kode Tesseract, boleh digabung dengan "+", contoh "ind+eng".
	OCR(inputPath, outputPath, lang, mode string) error

	// Sign menempelkan gambar (tanda tangan/stempel) ke halaman terpilih.
	Sign(inputPath, signaturePath, outputPath string, opts SignOptions) error

	// Thumbnails merender thumbnail JPEG tiap halaman ke outputDir dan mengembalikan
	// daftar path berurutan. Error jika halaman lebih banyak dari maxPages.
	Thumbnails(inputPath, outputDir string, maxPages int) ([]string, error)

	// Reorder menyusun ulang, menghapus (dengan tidak menyertakan), menggandakan,
	// dan memutar halaman sesuai urutan pages.
	Reorder(inputPath, outputPath string, pages []PageSpec) error
}
