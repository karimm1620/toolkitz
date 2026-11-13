import { useState } from "react";
import { imagesToPdf, pdfToImages } from "../services/api";
import { downloadBlob } from "../services/download";

type Mode = "images-to-pdf" | "pdf-to-images";

const MAX_IMAGES = 100;

export default function ImagePdfConverter() {
  const [mode, setMode] = useState<Mode>("images-to-pdf");
  const [loading, setLoading] = useState(false);

  // Gambar -> PDF
  const [images, setImages] = useState<File[]>([]);
  const [page, setPage] = useState<"a4" | "original">("a4");

  // PDF -> Gambar
  const [pdf, setPdf] = useState<File | null>(null);
  const [format, setFormat] = useState<"png" | "jpeg">("png");
  const [dpi, setDpi] = useState(150);

  const addImages = (list: FileList | null) => {
    if (!list) return;
    setImages((prev) => [...prev, ...Array.from(list)].slice(0, MAX_IMAGES));
  };

  const move = (index: number, delta: number) => {
    setImages((prev) => {
      const target = index + delta;
      if (target < 0 || target >= prev.length) return prev;
      const next = [...prev];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  };

  const remove = (index: number) =>
    setImages((prev) => prev.filter((_, i) => i !== index));

  const handleProcess = async () => {
    setLoading(true);
    try {
      if (mode === "images-to-pdf") {
        if (images.length === 0) return alert("Pilih minimal 1 gambar");
        const blob = await imagesToPdf(images, page);
        downloadBlob(blob, "images.pdf");
      } else {
        if (!pdf) return alert("Pilih file PDF terlebih dahulu");
        const blob = await pdfToImages(pdf, format, dpi);
        downloadBlob(blob, pdf.name.replace(/\.[^/.]+$/, "") + "_images.zip");
      }
    } catch (error) {
      alert(
        error instanceof Error ? error.message : "Terjadi kesalahan saat proses",
      );
    } finally {
      setLoading(false);
    }
  };

  const fileInputClass =
    "block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-6";
  const toggle = (active: boolean) =>
    `flex-1 py-2 rounded-lg text-sm font-medium border transition-colors ${active ? "border-blue-600 bg-blue-50 text-blue-700" : "border-gray-200 text-gray-600 hover:bg-gray-50"}`;

  return (
    <div className="max-w-xl mx-auto p-6 bg-white rounded-xl shadow-sm border border-gray-100">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">
        Gambar ↔ PDF
      </h2>

      <div className="flex gap-4 mb-6">
        <button
          onClick={() => setMode("images-to-pdf")}
          className={toggle(mode === "images-to-pdf")}
        >
          Gambar ke PDF
        </button>
        <button
          onClick={() => setMode("pdf-to-images")}
          className={toggle(mode === "pdf-to-images")}
        >
          PDF ke Gambar
        </button>
      </div>

      {mode === "images-to-pdf" ? (
        <>
          <p className="text-gray-500 mb-4 text-sm">
            Gabungkan gambar (JPG, PNG, GIF, BMP, TIFF) menjadi satu PDF.
            Satu gambar satu halaman, urutan mengikuti daftar di bawah.
          </p>

          <input
            type="file"
            accept="image/jpeg,image/png,image/gif,image/bmp,image/tiff"
            multiple
            onChange={(e) => {
              addImages(e.target.files);
              e.target.value = "";
            }}
            className={fileInputClass}
          />

          {images.length > 0 && (
            <ul className="mb-6 divide-y divide-gray-100 border border-gray-200 rounded-lg">
              {images.map((f, i) => (
                <li
                  key={`${f.name}-${i}`}
                  className="flex items-center gap-2 px-3 py-2 text-sm"
                >
                  <span className="w-6 text-gray-400">{i + 1}.</span>
                  <span className="flex-1 truncate text-gray-700">{f.name}</span>
                  <button
                    onClick={() => move(i, -1)}
                    disabled={i === 0}
                    aria-label="Naikkan"
                    className="px-2 text-gray-500 hover:text-blue-600 disabled:opacity-30"
                  >
                    ↑
                  </button>
                  <button
                    onClick={() => move(i, 1)}
                    disabled={i === images.length - 1}
                    aria-label="Turunkan"
                    className="px-2 text-gray-500 hover:text-blue-600 disabled:opacity-30"
                  >
                    ↓
                  </button>
                  <button
                    onClick={() => remove(i)}
                    aria-label="Hapus"
                    className="px-2 text-gray-400 hover:text-red-600"
                  >
                    ✕
                  </button>
                </li>
              ))}
            </ul>
          )}

          <div className="flex gap-4 mb-6">
            <button
              onClick={() => setPage("a4")}
              className={toggle(page === "a4")}
            >
              Ukuran A4
            </button>
            <button
              onClick={() => setPage("original")}
              className={toggle(page === "original")}
            >
              Ukuran asli gambar
            </button>
          </div>
        </>
      ) : (
        <>
          <p className="text-gray-500 mb-4 text-sm">
            Ubah tiap halaman PDF menjadi gambar. Hasilnya diunduh dalam satu
            file ZIP.
          </p>

          <input
            type="file"
            accept="application/pdf"
            onChange={(e) => setPdf(e.target.files?.[0] || null)}
            className={fileInputClass}
          />

          {pdf && (
            <div className="mb-6 text-sm text-gray-600 bg-gray-50 p-3 rounded border border-gray-200">
              <strong>File terpilih:</strong> {pdf.name} (
              {(pdf.size / 1024 / 1024).toFixed(2)} MB)
            </div>
          )}

          <div className="flex gap-4 mb-4">
            <button
              onClick={() => setFormat("png")}
              className={toggle(format === "png")}
            >
              PNG
            </button>
            <button
              onClick={() => setFormat("jpeg")}
              className={toggle(format === "jpeg")}
            >
              JPG
            </button>
          </div>

          <label className="block text-sm text-gray-600 mb-6">
            Resolusi: <strong>{dpi} dpi</strong>
            <input
              type="range"
              min={72}
              max={300}
              step={6}
              value={dpi}
              onChange={(e) => setDpi(Number(e.target.value))}
              className="w-full mt-2"
            />
            <span className="text-xs text-gray-400">
              Makin tinggi makin tajam, tetapi proses dan ukuran file lebih
              besar.
            </span>
          </label>
        </>
      )}

      <button
        onClick={handleProcess}
        disabled={
          loading || (mode === "images-to-pdf" ? images.length === 0 : !pdf)
        }
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2.5 px-4 rounded-lg disabled:opacity-50 transition-colors"
      >
        {loading
          ? "Memproses..."
          : mode === "images-to-pdf"
            ? "Buat PDF"
            : "Ubah ke Gambar"}
      </button>
    </div>
  );
}
