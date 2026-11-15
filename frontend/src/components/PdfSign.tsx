import { useEffect, useState } from "react";
import { signPdf, type SignPosition } from "../services/api";
import { downloadBlob } from "../services/download";

const POSITIONS: SignPosition[] = ["tl", "tc", "tr", "ml", "mc", "mr", "bl", "bc", "br"];

const PAGE_PRESETS = [
  { id: "last", label: "Halaman terakhir" },
  { id: "first", label: "Halaman pertama" },
  { id: "all", label: "Semua halaman" },
  { id: "custom", label: "Pilih sendiri" },
];

export default function PdfSign() {
  const [pdf, setPdf] = useState<File | null>(null);
  const [signature, setSignature] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [pagePreset, setPagePreset] = useState("last");
  const [customPages, setCustomPages] = useState("1");
  const [position, setPosition] = useState<SignPosition>("br");
  const [width, setWidth] = useState(25);
  const [removeWhite, setRemoveWhite] = useState(true);
  const [loading, setLoading] = useState(false);

  // Pratinjau dibuat di event handler; effect ini hanya membebaskan URL lama
  // saat pratinjau berganti atau komponen ditutup.
  useEffect(() => {
    return () => {
      if (preview) URL.revokeObjectURL(preview);
    };
  }, [preview]);

  const handleSignature = (file: File | null) => {
    setSignature(file);
    setPreview(file ? URL.createObjectURL(file) : null);
  };

  const handleProcess = async () => {
    if (!pdf) return alert("Pilih file PDF terlebih dahulu");
    if (!signature) return alert("Pilih gambar tanda tangan terlebih dahulu");

    const pages = pagePreset === "custom" ? customPages.trim() : pagePreset;
    if (!pages) return alert("Isi nomor halaman, contoh: 1,3,5-7");

    setLoading(true);
    try {
      const blob = await signPdf(pdf, signature, { pages, position, width, removeWhite });
      downloadBlob(blob, `signed_${pdf.name}`);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Terjadi kesalahan saat menambahkan tanda tangan");
    } finally {
      setLoading(false);
    }
  };

  const toggle = (active: boolean) =>
    `py-2 rounded-lg text-sm font-medium border transition-colors ${active ? "border-blue-600 bg-blue-50 text-blue-700" : "border-gray-200 text-gray-600 hover:bg-gray-50"}`;
  const fileInputClass =
    "block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-6";

  return (
    <div className="max-w-xl mx-auto p-6 bg-white rounded-xl shadow-sm border border-gray-100">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">Tanda Tangan PDF</h2>
      <p className="text-gray-500 mb-6 text-sm">
        Tempelkan gambar tanda tangan atau stempel ke PDF. Foto tanda tangan di
        kertas putih bisa dipakai: latar putihnya otomatis dibuat transparan.
      </p>

      <label className="block text-sm text-gray-600 mb-2">1. File PDF</label>
      <input
        type="file"
        accept="application/pdf"
        onChange={(e) => setPdf(e.target.files?.[0] || null)}
        className={fileInputClass}
      />

      <label className="block text-sm text-gray-600 mb-2">2. Gambar tanda tangan (PNG / JPG)</label>
      <input
        type="file"
        accept="image/png,image/jpeg"
        onChange={(e) => handleSignature(e.target.files?.[0] || null)}
        className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-3"
      />
      {preview && (
        <div className="mb-6 p-3 rounded border border-gray-200 bg-[repeating-conic-gradient(#f3f4f6_0%_25%,#fff_0%_50%)] bg-[length:16px_16px]">
          <img src={preview} alt="Pratinjau tanda tangan" className="max-h-24 mx-auto" />
        </div>
      )}

      <label className="block text-sm text-gray-600 mb-2">3. Halaman</label>
      <div className="grid grid-cols-2 gap-3 mb-3">
        {PAGE_PRESETS.map((p) => (
          <button key={p.id} type="button" onClick={() => setPagePreset(p.id)} className={toggle(pagePreset === p.id)}>
            {p.label}
          </button>
        ))}
      </div>
      {pagePreset === "custom" && (
        <input
          type="text"
          value={customPages}
          onChange={(e) => setCustomPages(e.target.value)}
          placeholder="Contoh: 1,3,5-7"
          className="w-full mb-3 px-3 py-2 border border-gray-200 rounded-lg text-sm"
        />
      )}

      <label className="block text-sm text-gray-600 mt-3 mb-2">4. Posisi</label>
      <div className="grid grid-cols-3 gap-2 w-40 mb-6">
        {POSITIONS.map((p) => (
          <button
            key={p}
            type="button"
            onClick={() => setPosition(p)}
            aria-label={`Posisi ${p}`}
            className={`h-9 rounded border transition-colors ${position === p ? "border-blue-600 bg-blue-600" : "border-gray-300 bg-gray-50 hover:bg-gray-100"}`}
          />
        ))}
      </div>

      <label className="block text-sm text-gray-600 mb-6">
        Ukuran: <strong>{width}%</strong> dari lebar halaman
        <input
          type="range"
          min={5}
          max={80}
          step={1}
          value={width}
          onChange={(e) => setWidth(Number(e.target.value))}
          className="w-full mt-2"
        />
      </label>

      <label className="flex items-center gap-2 text-sm text-gray-600 mb-6">
        <input type="checkbox" checked={removeWhite} onChange={(e) => setRemoveWhite(e.target.checked)} />
        Jadikan latar putih transparan
      </label>

      <button
        onClick={handleProcess}
        disabled={!pdf || !signature || loading}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2.5 px-4 rounded-lg disabled:opacity-50 transition-colors"
      >
        {loading ? "Memproses..." : "Tempel Tanda Tangan"}
      </button>
    </div>
  );
}
