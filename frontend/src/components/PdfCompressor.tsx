import { useState } from "react";
import { compressPdf, type CompressLevel } from "../services/api";
import { downloadBlob } from "../services/download";

const LEVELS: { id: CompressLevel; title: string; desc: string }[] = [
  { id: "low", title: "Ringan", desc: "Kualitas tinggi, ukuran turun sedikit" },
  { id: "medium", title: "Seimbang", desc: "Disarankan untuk kebanyakan file" },
  { id: "high", title: "Maksimal", desc: "Ukuran terkecil, kualitas gambar turun" },
];

const formatSize = (bytes: number) => {
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
  return `${(bytes / 1024).toFixed(1)} KB`;
};

export default function PdfCompressor() {
  const [file, setFile] = useState<File | null>(null);
  const [level, setLevel] = useState<CompressLevel>("medium");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ before: number; after: number } | null>(
    null,
  );

  const handleProcess = async () => {
    if (!file) return alert("Pilih file PDF terlebih dahulu");

    setLoading(true);
    setResult(null);
    try {
      const blob = await compressPdf(file, level);
      setResult({ before: file.size, after: blob.size });
      downloadBlob(blob, `compressed_${file.name}`);
    } catch (error) {
      alert(
        error instanceof Error
          ? error.message
          : "Terjadi kesalahan saat mengompresi PDF",
      );
    } finally {
      setLoading(false);
    }
  };

  const saved = result
    ? Math.round(((result.before - result.after) / result.before) * 100)
    : 0;

  return (
    <div className="max-w-xl mx-auto p-6 bg-white rounded-xl shadow-sm border border-gray-100">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">Kompresi PDF</h2>
      <p className="text-gray-500 mb-6 text-sm">
        Perkecil ukuran file PDF. Pilih tingkat kompresi: makin tinggi, makin
        kecil ukurannya, tetapi kualitas gambar di dalam PDF ikut turun.
      </p>

      <div className="grid grid-cols-3 gap-3 mb-6">
        {LEVELS.map((l) => (
          <button
            key={l.id}
            type="button"
            onClick={() => setLevel(l.id)}
            className={`p-3 rounded-lg text-left border transition-colors ${level === l.id ? "border-blue-600 bg-blue-50" : "border-gray-200 hover:bg-gray-50"}`}
          >
            <div
              className={`text-sm font-semibold ${level === l.id ? "text-blue-700" : "text-gray-700"}`}
            >
              {l.title}
            </div>
            <div className="text-xs text-gray-500 mt-1">{l.desc}</div>
          </button>
        ))}
      </div>

      <input
        type="file"
        accept="application/pdf"
        onChange={(e) => {
          setFile(e.target.files?.[0] || null);
          setResult(null);
        }}
        className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-6"
      />

      {file && (
        <div className="mb-6 text-sm text-gray-600 bg-gray-50 p-3 rounded border border-gray-200">
          <strong>File terpilih:</strong> {file.name} ({formatSize(file.size)})
        </div>
      )}

      {result && (
        <div className="mb-6 text-sm bg-green-50 text-green-800 p-3 rounded border border-green-200">
          {saved > 0 ? (
            <>
              Ukuran turun <strong>{saved}%</strong>:{" "}
              {formatSize(result.before)} → {formatSize(result.after)}
            </>
          ) : (
            <>
              File ini sudah cukup ringkas ({formatSize(result.before)} →{" "}
              {formatSize(result.after)}). Coba tingkat yang lebih tinggi.
            </>
          )}
        </div>
      )}

      <button
        onClick={handleProcess}
        disabled={!file || loading}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2.5 px-4 rounded-lg disabled:opacity-50 transition-colors"
      >
        {loading ? "Mengompresi File..." : "Mulai Kompresi"}
      </button>
    </div>
  );
}
