# Toolkitz

Toolkitz adalah platform pemrosesan dokumen dan gambar lintas platform (Cross-Platform) berkinerja tinggi. Proyek ini dibangun menggunakan arsitektur **Clean Architecture** dan menawarkan dua cara penggunaan: melalui antarmuka web modern bergaya kartu (Web UI) atau langsung dari Terminal Anda (CLI).

## Fitur Utama

### 📄 Manipulasi PDF
*   **Merge:** Menggabungkan banyak PDF menjadi satu.
*   **Split:** Memisahkan setiap halaman PDF ke dalam file berbeda (Output ZIP).
*   **Pages:** Mengekstrak, menghapus, atau memutar halaman spesifik (misal: 1, 3, 5-10).
*   **Security:** Kunci PDF dengan enkripsi AES-256 (Encrypt) atau hapus proteksinya (Decrypt).
*   **Compress:** Optimalisasi struktur internal PDF tanpa mengurangi kualitas visual.
*   **Watermark & Page Numbers:** Stempel teks diagonal atau nomor halaman otomatis.

### 🖼️ Image Processor & Doc Converter
*   **Resize & Crop:** Pemotongan tengah presisi dan pengubahan resolusi gambar.
*   **Format & Kompresi:** Dukungan kompresi ke JPEG, PNG, dan WEBP dengan slider kualitas (1-100%).
*   **Flip & Rotate:** Cerminkan gambar atau putar (90, 180, 270 derajat).
*   **Doc Converter:** Konversi cepat dari format dokumen lama ke PDF.

## Tech Stack

*   **Backend:** Go (Golang), Fiber (REST API), Cobra (CLI Engine)
*   **Frontend:** React, TypeScript, Tailwind CSS, Vite
*   **Core Libraries:** `pdfcpu` (Manipulasi PDF), `disintegration/imaging` & `go-webp` (Gambar)

## Instalasi & Menjalankan via Docker (Direkomendasikan)

Cara termudah untuk menjalankan Toolkitz adalah dengan Docker Compose.

1. Clone repositori ini:
   ```bash
   git clone https://github.com/karimm1620/toolkitz.git
   cd toolkitz

2. ```bash
    docker-compose up -d --build

3. Buka browser dan akses http://localhost untuk Web UI.
(Backend API akan berjalan di http://localhost:3000).

> **Tidak perlu mengubah IP di `docker-compose.yml`.** Web UI memanggil `/api` di alamat yang sama,
> lalu nginx meneruskannya ke container backend. Dari device lain di jaringan yang sama cukup buka
> `http://IP-LAPTOP-KAMU` (port 80). Kalau gagal terhubung, cek log: `docker compose logs backend`
> dan `docker compose ps` (backend harus berstatus `healthy`).

---

# Penggunaan CLI (Command Line Interface)
Toolkitz memiliki CLI mandiri yang sangat cepat. Anda bisa mem-build CLI-nya dengan masuk ke folder backend lalu eksekusi:
1. ``` go
    go build -o pdf-tool ./cmd/cli
2. ``` bash
    ./pdf-tool --help
Contoh command:

``` bash
# Menggabungkan PDF
./pdf-tool merge output.pdf bab1.pdf bab2.pdf bab3.pdf

# Menambahkan password
./pdf-tool encrypt rahasia.pdf rahasia_locked.pdf --password "superaman"

# Mengekstrak halaman 1, 3, dan 5
./pdf-tool extract buku.pdf bab1.pdf --pages "1,3,5"

# Resize, putar, dan ubah format gambar ke WEBP
./pdf-tool resize foto.jpg hasil.webp --width 800 --format webp