import { useState } from "react";
import PdfMerger from "./components/PdfMerger";
import PdfSplitter from "./components/PdfSplitter";
import PdfSecurity from "./components/PdfSecurity";
import PdfPages from "./components/PdfPages";
import PdfCompressor from "./components/PdfCompressor";
import PdfWatermark from "./components/PdfWatermark";
import DocConverter from "./components/DocConverter";
import ImageResizer from "./components/ImageResizer";
import ImagePdfConverter from "./components/ImagePdfConverter";

type ToolId =
  | "home"
  | "pdf-merge"
  | "pdf-split"
  | "pdf-pages"
  | "pdf-security"
  | "pdf-compress"
  | "pdf-watermark"
  | "doc-convert"
  | "image-pdf"
  | "image-resize";

// Data konfigurasi untuk Grid Beranda
const TOOLS = [
  {
    id: "pdf-merge" as ToolId,
    title: "Gabungkan PDF",
    desc: "Gabungkan banyak file PDF menjadi satu file secara berurutan.",
    icon: (
      <svg
        className="w-10 h-10 text-red-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
        />
      </svg>
    ),
  },
  {
    id: "pdf-split" as ToolId,
    title: "Pisahkan PDF",
    desc: "Ekstrak setiap halaman dari PDF menjadi file terpisah (zip).",
    icon: (
      <svg
        className="w-10 h-10 text-orange-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2"
        />
      </svg>
    ),
  },
  {
    id: "pdf-compress" as ToolId,
    title: "Kompresi PDF",
    desc: "Perkecil ukuran file PDF dengan tiga tingkat kompresi.",
    icon: (
      <svg
        className="w-10 h-10 text-green-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
        />
      </svg>
    ),
  },
  {
    id: "doc-convert" as ToolId,
    title: "Konversi Dokumen",
    desc: "Ubah Word, Excel, atau PowerPoint ke PDF, atau PDF ke Word.",
    icon: (
      <svg
        className="w-10 h-10 text-blue-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
        />
      </svg>
    ),
  },
  {
    id: "pdf-pages" as ToolId,
    title: "Atur Halaman",
    desc: "Hapus, ekstrak, atau putar halaman tertentu pada PDF Anda.",
    icon: (
      <svg
        className="w-10 h-10 text-purple-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
        />
      </svg>
    ),
  },
  {
    id: "pdf-security" as ToolId,
    title: "Keamanan PDF",
    desc: "Kunci PDF Anda dengan password atau hapus password lama.",
    icon: (
      <svg
        className="w-10 h-10 text-gray-700 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
        />
      </svg>
    ),
  },
  {
    id: "pdf-watermark" as ToolId,
    title: "Watermark",
    desc: "Tambahkan teks watermark diagonal atau nomor halaman.",
    icon: (
      <svg
        className="w-10 h-10 text-teal-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"
        />
      </svg>
    ),
  },
  {
    id: "image-pdf" as ToolId,
    title: "Gambar ↔ PDF",
    desc: "Gabungkan gambar menjadi PDF, atau ubah halaman PDF jadi gambar.",
    icon: (
      <svg
        className="w-10 h-10 text-indigo-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M4 16l4-4 3 3 4-5 5 6M4 4h16a1 1 0 011 1v14a1 1 0 01-1 1H4a1 1 0 01-1-1V5a1 1 0 011-1z"
        />
      </svg>
    ),
  },
  {
    id: "image-resize" as ToolId,
    title: "Toolkit Gambar",
    desc: "Crop, resize, kompresi, putar, dan ubah format gambar.",
    icon: (
      <svg
        className="w-10 h-10 text-pink-500 mb-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.5}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
        />
      </svg>
    ),
  },
];

function App() {
  const [activeTool, setActiveTool] = useState<ToolId>("home");

  return (
    <div className="min-h-screen bg-[#f7f9fc] font-sans">
      {/* Header Statis */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10 shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <button
            onClick={() => setActiveTool("home")}
            className="flex items-center gap-2 hover:opacity-80 transition-opacity focus:outline-none"
          >
            {/* Logo Simple */}
            <div className="w-8 h-8 bg-red-600 rounded flex items-center justify-center">
              <span className="text-white font-bold text-xl">T</span>
            </div>
            <h1 className="text-2xl font-bold text-gray-800 tracking-tight">
              Toolkit<span className="text-red-600">z</span>
            </h1>
          </button>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="max-w-6xl mx-auto px-4 py-10">
        {/* Tampilan Grid Beranda */}
        {activeTool === "home" && (
          <div className="animate-in fade-in duration-500">
            <div className="text-center max-w-2xl mx-auto mb-12">
              <h2 className="text-4xl font-extrabold text-gray-900 mb-4 tracking-tight">
                Setiap alat yang Anda butuhkan
              </h2>
              <p className="text-lg text-gray-500">
                Semua fitur pemrosesan PDF dan Gambar di satu tempat. 100%
                gratis, cepat, dan aman. Pilih alat di bawah untuk memulai.
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
              {TOOLS.map((tool) => (
                <button
                  key={tool.id}
                  onClick={() => setActiveTool(tool.id)}
                  className="group flex flex-col items-center text-center bg-white p-8 rounded-2xl shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 border border-transparent hover:border-red-100 focus:outline-none"
                >
                  <div className="group-hover:scale-110 transition-transform duration-300">
                    {tool.icon}
                  </div>
                  <h3 className="text-lg font-bold text-gray-800 mb-2">
                    {tool.title}
                  </h3>
                  <p className="text-sm text-gray-500 leading-relaxed">
                    {tool.desc}
                  </p>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Tampilan Halaman Fitur (Alat) */}
        {activeTool !== "home" && (
          <div className="animate-in slide-in-from-bottom-4 fade-in duration-300">
            {/* Tombol Back */}
            <button
              onClick={() => setActiveTool("home")}
              className="flex items-center gap-2 text-sm font-medium text-gray-500 hover:text-red-600 mb-8 transition-colors"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M10 19l-7-7m0 0l7-7m-7 7h18"
                />
              </svg>
              Kembali ke Beranda
            </button>

            {/* Render Komponen Sesuai Pilihan */}
            <div className="pb-12">
              {activeTool === "pdf-merge" && <PdfMerger />}
              {activeTool === "pdf-split" && <PdfSplitter />}
              {activeTool === "pdf-pages" && <PdfPages />}
              {activeTool === "pdf-security" && <PdfSecurity />}
              {activeTool === "pdf-compress" && <PdfCompressor />}
              {activeTool === "pdf-watermark" && <PdfWatermark />}
              {activeTool === "doc-convert" && <DocConverter />}
              {activeTool === "image-pdf" && <ImagePdfConverter />}
              {activeTool === "image-resize" && <ImageResizer />}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
