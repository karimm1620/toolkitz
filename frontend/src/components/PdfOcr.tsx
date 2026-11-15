import { useState } from "react";
import { ocrPdf } from "../services/api";
import { downloadBlob } from "../services/download";

const LANGS = [
  { id: "ind+eng", label: "Indonesia + Inggris" },
  { id: "ind", label: "Indonesia" },
  { id: "eng", label: "Inggris" },
];

export default function PdfOcr() {
  const [file, setFile] = useState<File | null>(null);
  const [lang, setLang] = useState("ind+eng");
  const [output, setOutput] = useState<"pdf" | "txt">("pdf");
  const [loading, setLoading] = useState(false);

  const handleProcess = async () => {
    if (!file) return alert("Pilih file PDF terlebih dahulu");

    setLoading(true);
    try {
      const blob = await ocrPdf(file, lang, output);
      const base = file.name.replace(/\.[^/.]+$/, "");
      downloadBlob(blob, `${base}_ocr.${output}`);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Terjadi kesalahan saat OCR");
    } finally {
      setLoading(false);
    }
  };

  const toggle = (active: boolean) =>
    `flex-1 py-2 rounded-lg text-sm font-medium border transition-colors ${active ? "border-blue-600 bg-blue-50 text-blue-700" : "border-gray-200 text-gray-600 hover:bg-gray-50"}`;

  return (
    <div className="max-w-xl mx-auto p-6 bg-white rounded-xl shadow-sm border border-gray-100">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">OCR PDF</h2>
      <p className="text-gray-500 mb-6 text-sm">
        Ubah PDF hasil scan (berisi gambar saja) menjadi PDF yang teksnya bisa
        dicari dan disalin, atau ambil teksnya saja. Maksimal 100 halaman per
        proses; makin banyak halaman, makin lama.
      </p>

      <label className="block text-sm text-gray-600 mb-2">Bahasa dokumen</label>
      <div className="grid grid-cols-3 gap-3 mb-6">
        {LANGS.map((l) => (
          <button key={l.id} type="button" onClick={() => setLang(l.id)} className={toggle(lang === l.id)}>
            {l.label}
          </button>
        ))}
      </div>

      <label className="block text-sm text-gray-600 mb-2">Hasil</label>
      <div className="flex gap-4 mb-6">
        <button type="button" onClick={() => setOutput("pdf")} className={toggle(output === "pdf")}>
          PDF yang bisa dicari
        </button>
        <button type="button" onClick={() => setOutput("txt")} className={toggle(output === "txt")}>
          Teks saja (.txt)
        </button>
      </div>

      <input
        type="file"
        accept="application/pdf"
        onChange={(e) => setFile(e.target.files?.[0] || null)}
        className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-6"
      />

      {file && (
        <div className="mb-6 text-sm text-gray-600 bg-gray-50 p-3 rounded border border-gray-200">
          <strong>File terpilih:</strong> {file.name} ({(file.size / 1024 / 1024).toFixed(2)} MB)
        </div>
      )}

      <button
        onClick={handleProcess}
        disabled={!file || loading}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2.5 px-4 rounded-lg disabled:opacity-50 transition-colors"
      >
        {loading ? "Memproses OCR (bisa beberapa menit)..." : "Mulai OCR"}
      </button>
    </div>
  );
}
