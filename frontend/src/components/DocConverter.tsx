import { useState } from "react";
import { convertDocument } from "../services/api";
import { downloadBlob } from "../services/download";

type ConvertType = "word-to-pdf" | "pdf-to-word" | "excel-to-pdf" | "ppt-to-pdf";

const OPTIONS: {
  id: ConvertType;
  label: string;
  accept: string;
  outExt: string;
}[] = [
  { id: "word-to-pdf", label: "Word ke PDF", accept: ".docx,.doc", outExt: ".pdf" },
  { id: "pdf-to-word", label: "PDF ke Word", accept: ".pdf", outExt: ".docx" },
  { id: "excel-to-pdf", label: "Excel ke PDF", accept: ".xlsx", outExt: ".pdf" },
  { id: "ppt-to-pdf", label: "PowerPoint ke PDF", accept: ".pptx", outExt: ".pdf" },
];

export default function DocConverter() {
  const [file, setFile] = useState<File | null>(null);
  const [type, setType] = useState<ConvertType>("word-to-pdf");
  const [loading, setLoading] = useState(false);

  const current = OPTIONS.find((o) => o.id === type) ?? OPTIONS[0];

  const handleProcess = async () => {
    if (!file) return alert("Pilih file terlebih dahulu");

    setLoading(true);
    try {
      const blob = await convertDocument(file, type);
      const newFileName = file.name.replace(/\.[^/.]+$/, "") + current.outExt;
      downloadBlob(blob, newFileName);
    } catch (error) {
      alert(
        error instanceof Error
          ? error.message
          : "Terjadi kesalahan saat konversi",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto p-6 bg-white rounded-xl shadow-sm border border-gray-100">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">
        Konversi Dokumen
      </h2>

      <div className="grid grid-cols-2 gap-3 mb-6">
        {OPTIONS.map((o) => (
          <button
            key={o.id}
            onClick={() => {
              setType(o.id);
              setFile(null);
            }}
            className={`py-2 rounded-lg text-sm font-medium border transition-colors ${type === o.id ? "border-blue-600 bg-blue-50 text-blue-700" : "border-gray-200 text-gray-600 hover:bg-gray-50"}`}
          >
            {o.label}
          </button>
        ))}
      </div>

      <input
        key={type}
        type="file"
        accept={current.accept}
        onChange={(e) => setFile(e.target.files?.[0] || null)}
        className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-6"
      />

      {file && (
        <div className="mb-6 text-sm text-gray-600 bg-gray-50 p-3 rounded border border-gray-200">
          <strong>File terpilih:</strong> {file.name}
        </div>
      )}

      <button
        onClick={handleProcess}
        disabled={!file || loading}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2.5 px-4 rounded-lg disabled:opacity-50 transition-colors"
      >
        {loading ? "Mengkonversi..." : "Mulai Konversi"}
      </button>
    </div>
  );
}
