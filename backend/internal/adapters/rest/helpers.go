package rest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// newWorkDir membuat folder kerja unik untuk satu request.
// Semua file input/output request disimpan di sini, sehingga dua pengguna yang
// mengunggah file bernama sama di waktu bersamaan tidak saling menimpa.
// Pemanggil wajib `defer os.RemoveAll(dir)`.
func newWorkDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix+"-*")
}

// safeBase mengambil nama file saja (tanpa path) dari nama unggahan.
func safeBase(name, fallback string) string {
	b := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	if b == "." || b == "/" || b == "" {
		return fallback
	}
	return b
}

// dispositionName membersihkan nama file agar aman dipakai di header Content-Disposition.
func dispositionName(name string) string {
	r := strings.NewReplacer(`"`, "_", "\r", "_", "\n", "_", "\\", "_")
	return r.Replace(name)
}

// sendDownload mengirim file sebagai unduhan.
// File dibuka dulu lalu dialirkan; di Linux file yang sudah terbuka tetap bisa dibaca
// walaupun folder kerjanya dihapus oleh `defer os.RemoveAll` milik handler.
func sendDownload(c *fiber.Ctx, path, downloadName, contentType string) error {
	f, err := os.Open(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca file hasil"})
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca file hasil"})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", `attachment; filename="`+dispositionName(downloadName)+`"`)
	return c.SendStream(f, int(info.Size()))
}
