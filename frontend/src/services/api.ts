// Docker/production: "/api" (di-proxy nginx ke backend, tidak bergantung IP).
// Dev lokal (`npm run dev`): set VITE_API_URL=http://localhost:3000/api di frontend/.env.local
const API_URL = import.meta.env.VITE_API_URL || "/api";

export interface ProcessImageOptions {
  width: number;
  height: number;
  format: string;
  cropWidth: number;
  cropHeight: number;
  rotate: number;
  flipH: boolean;
  flipV: boolean;
  quality: number;
}

export const mergePdfs = async (files: File[]): Promise<Blob> => {
  const formData = new FormData();
  files.forEach((file) => formData.append("pdfs", file));

  const response = await fetch(`${API_URL}/pdf/merge`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) throw new Error("Gagal menggabungkan PDF");
  return await response.blob();
};

export const splitPdf = async (file: File): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);

  const response = await fetch(`${API_URL}/pdf/split`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) throw new Error("Gagal memisahkan PDF");
  return await response.blob();
};

export const processImage = async (
  file: File,
  opts: ProcessImageOptions,
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("image", file);
  formData.append("format", opts.format);
  formData.append("quality", opts.quality.toString());
  if (opts.width > 0) formData.append("width", opts.width.toString());
  if (opts.height > 0) formData.append("height", opts.height.toString());
  if (opts.cropWidth > 0)
    formData.append("crop_width", opts.cropWidth.toString());
  if (opts.cropHeight > 0)
    formData.append("crop_height", opts.cropHeight.toString());
  if (opts.rotate !== 0) formData.append("rotate", opts.rotate.toString());
  if (opts.flipH) formData.append("flip_h", "true");
  if (opts.flipV) formData.append("flip_v", "true");

  const response = await fetch(`${API_URL}/image/resize`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) throw new Error("Gagal memproses gambar");
  return await response.blob();
};

export const convertDocument = async (
  file: File,
  type: "word-to-pdf" | "pdf-to-word",
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("document", file);

  const response = await fetch(`${API_URL}/convert/${type}`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    throw new Error(errorData?.error || "Gagal mengkonversi dokumen");
  }
  return await response.blob();
};

export const securePdf = async (
  file: File,
  password: string,
  action: "encrypt" | "decrypt",
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("password", password);

  const response = await fetch(`${API_URL}/pdf/${action}`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const err = await response.json().catch(() => null);
    throw new Error(err?.error || `Gagal melakukan ${action}`);
  }
  return await response.blob();
};

export const manipulatePdfPages = async (
  file: File,
  action: "extract" | "remove" | "rotate",
  pages: string,
  rotation: number = 90,
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("action", action);
  formData.append("pages", pages);
  if (action === "rotate") {
    formData.append("rotation", rotation.toString());
  }

  const response = await fetch(`${API_URL}/pdf/pages`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const err = await response.json().catch(() => null);
    throw new Error(err?.error || `Gagal melakukan ${action} pada halaman`);
  }
  return await response.blob();
};

export const compressPdf = async (file: File): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);

  const response = await fetch(`${API_URL}/pdf/compress`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) throw new Error("Gagal mengompresi PDF");
  return await response.blob();
};

export const addPdfWatermark = async (
  file: File,
  action: "watermark" | "pagenumbers",
  text?: string,
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("action", action);
  if (text) formData.append("text", text);

  const response = await fetch(`${API_URL}/pdf/watermark`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const err = await response.json().catch(() => null);
    throw new Error(err?.error || "Gagal memproses dokumen");
  }
  return await response.blob();
};
