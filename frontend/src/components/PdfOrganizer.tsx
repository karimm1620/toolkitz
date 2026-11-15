import { useRef, useState } from "react";
import { getPdfThumbnails, reorderPdf } from "../services/api";
import { downloadBlob } from "../services/download";

interface Item {
  uid: number; // identitas unik kartu (halaman yang sama bisa muncul dua kali)
  page: number; // nomor halaman asli (mulai 1)
  rotate: number; // 0 | 90 | 180 | 270
}

export default function PdfOrganizer() {
  const [file, setFile] = useState<File | null>(null);
  const [thumbs, setThumbs] = useState<string[]>([]);
  const [items, setItems] = useState<Item[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [dragUid, setDragUid] = useState<number | null>(null);
  const uidRef = useRef(0);

  const nextUid = () => ++uidRef.current;

  const buildItems = (count: number): Item[] =>
    Array.from({ length: count }, (_, i) => ({
      uid: nextUid(),
      page: i + 1,
      rotate: 0,
    }));

  const handleFile = async (selected: File | null) => {
    setFile(selected);
    setThumbs([]);
    setItems([]);
    if (!selected) return;

    setLoading(true);
    try {
      const { thumbnails } = await getPdfThumbnails(selected);
      setThumbs(thumbnails);
      setItems(buildItems(thumbnails.length));
    } catch (error) {
      setFile(null);
      alert(error instanceof Error ? error.message : "Gagal memuat halaman PDF");
    } finally {
      setLoading(false);
    }
  };

  const move = (index: number, delta: number) =>
    setItems((prev) => {
      const target = index + delta;
      if (target < 0 || target >= prev.length) return prev;
      const next = [...prev];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });

  const rotate = (uid: number) =>
    setItems((prev) =>
      prev.map((it) => (it.uid === uid ? { ...it, rotate: (it.rotate + 90) % 360 } : it)),
    );

  const remove = (uid: number) => setItems((prev) => prev.filter((it) => it.uid !== uid));

  const duplicate = (index: number) =>
    setItems((prev) => {
      const next = [...prev];
      next.splice(index + 1, 0, { ...prev[index], uid: nextUid() });
      return next;
    });

  const dropOn = (targetUid: number) => {
    if (dragUid === null || dragUid === targetUid) return;
    setItems((prev) => {
      const from = prev.findIndex((it) => it.uid === dragUid);
      const to = prev.findIndex((it) => it.uid === targetUid);
      if (from < 0 || to < 0) return prev;
      const next = [...prev];
      const [moved] = next.splice(from, 1);
      next.splice(to, 0, moved);
      return next;
    });
    setDragUid(null);
  };

  const handleSave = async () => {
    if (!file || items.length === 0) return;
    setSaving(true);
    try {
      const blob = await reorderPdf(
        file,
        items.map(({ page, rotate }) => ({ page, rotate })),
      );
      downloadBlob(blob, `reordered_${file.name}`);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Gagal menyimpan PDF");
    } finally {
      setSaving(false);
    }
  };

  const btn =
    "px-2 py-1 text-xs rounded border border-gray-200 text-gray-600 hover:bg-gray-50 hover:text-blue-600 disabled:opacity-30";

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white rounded-xl shadow-sm border border-gray-100">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">Susun Halaman PDF</h2>
      <p className="text-gray-500 mb-6 text-sm">
        Seret halaman untuk mengubah urutan (di layar sentuh pakai tombol ◀ ▶).
        Putar, gandakan, atau hapus halaman yang tidak perlu, lalu simpan.
        Maksimal 300 halaman.
      </p>

      <input
        type="file"
        accept="application/pdf"
        onChange={(e) => handleFile(e.target.files?.[0] || null)}
        className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-6"
      />

      {loading && <p className="text-sm text-gray-500 mb-6">Memuat halaman...</p>}

      {items.length > 0 && (
        <>
          <div className="flex items-center justify-between mb-4 text-sm text-gray-600">
            <span>
              {items.length} halaman
              {items.length !== thumbs.length && ` (asli: ${thumbs.length})`}
            </span>
            <button
              type="button"
              onClick={() => setItems(buildItems(thumbs.length))}
              className="text-blue-600 hover:underline"
            >
              Kembalikan urutan awal
            </button>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4 mb-6">
            {items.map((it, index) => (
              <div
                key={it.uid}
                draggable
                onDragStart={() => setDragUid(it.uid)}
                onDragEnd={() => setDragUid(null)}
                onDragOver={(e) => e.preventDefault()}
                onDrop={() => dropOn(it.uid)}
                className={`border rounded-lg p-2 bg-white cursor-grab ${dragUid === it.uid ? "opacity-40 border-blue-400" : "border-gray-200"}`}
              >
                <div className="h-44 flex items-center justify-center bg-gray-50 rounded overflow-hidden">
                  <img
                    src={thumbs[it.page - 1]}
                    alt={`Halaman ${it.page}`}
                    draggable={false}
                    className="max-h-full max-w-full object-contain transition-transform"
                    style={{
                      transform: `rotate(${it.rotate}deg)${it.rotate % 180 !== 0 ? " scale(0.72)" : ""}`,
                    }}
                  />
                </div>
                <div className="text-center text-xs text-gray-500 my-2">
                  Hal. {it.page}
                  {it.rotate !== 0 && ` · ${it.rotate}°`}
                </div>
                <div className="flex flex-wrap justify-center gap-1">
                  <button type="button" className={btn} disabled={index === 0} onClick={() => move(index, -1)} aria-label="Geser ke kiri">◀</button>
                  <button type="button" className={btn} disabled={index === items.length - 1} onClick={() => move(index, 1)} aria-label="Geser ke kanan">▶</button>
                  <button type="button" className={btn} onClick={() => rotate(it.uid)} aria-label="Putar 90 derajat">⟳</button>
                  <button type="button" className={btn} onClick={() => duplicate(index)} aria-label="Gandakan halaman">⧉</button>
                  <button type="button" className={btn} disabled={items.length === 1} onClick={() => remove(it.uid)} aria-label="Hapus halaman">✕</button>
                </div>
              </div>
            ))}
          </div>

          <button
            onClick={handleSave}
            disabled={saving}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2.5 px-4 rounded-lg disabled:opacity-50 transition-colors"
          >
            {saving ? "Menyimpan..." : "Simpan PDF"}
          </button>
        </>
      )}
    </div>
  );
}
