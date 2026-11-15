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
  type: "word-to-pdf" | "pdf-to-word" | "excel-to-pdf" | "ppt-to-pdf",
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

export type CompressLevel = "low" | "medium" | "high";

export const compressPdf = async (
  file: File,
  level: CompressLevel = "medium",
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("level", level);

  const response = await fetch(`${API_URL}/pdf/compress`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const err = await response.json().catch(() => null);
    throw new Error(err?.error || "Gagal mengompresi PDF");
  }
  return await response.blob();
};

export const imagesToPdf = async (
  files: File[],
  page: "a4" | "original" = "a4",
): Promise<Blob> => {
  const formData = new FormData();
  files.forEach((file) => formData.append("images", file));
  formData.append("page", page);

  const response = await fetch(`${API_URL}/pdf/from-images`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const err = await response.json().catch(() => null);
    throw new Error(err?.error || "Gagal membuat PDF dari gambar");
  }
  return await response.blob();
};

export const pdfToImages = async (
  file: File,
  format: "png" | "jpeg" = "png",
  dpi: number = 150,
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("format", format);
  formData.append("dpi", dpi.toString());

  const response = await fetch(`${API_URL}/pdf/to-images`, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const err = await response.json().catch(() => null);
    throw new Error(err?.error || "Gagal mengubah PDF menjadi gambar");
  }
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

// ---------- OCR, Tanda Tangan, Susun Halaman, HEIC ----------

const readError = async (response: Response, fallback: string) => {
  const err = await response.json().catch(() => null);
  return err?.error || fallback;
};

export const ocrPdf = async (
  file: File,
  lang: string,
  output: "pdf" | "txt",
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("lang", lang);
  formData.append("output", output);

  const response = await fetch(`${API_URL}/pdf/ocr`, {
    method: "POST",
    body: formData,
  });
  if (!response.ok) throw new Error(await readError(response, "Gagal menjalankan OCR"));
  return await response.blob();
};

export type SignPosition =
  | "tl" | "tc" | "tr"
  | "ml" | "mc" | "mr"
  | "bl" | "bc" | "br";

export interface SignOptions {
  pages: string; // "first" | "last" | "all" | "1,3,5-7"
  position: SignPosition;
  width: number; // % lebar halaman
  removeWhite: boolean;
}

export const signPdf = async (
  pdf: File,
  signature: File,
  opts: SignOptions,
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", pdf);
  formData.append("signature", signature);
  formData.append("pages", opts.pages);
  formData.append("position", opts.position);
  formData.append("width", opts.width.toString());
  formData.append("remove_white", opts.removeWhite ? "true" : "false");

  const response = await fetch(`${API_URL}/pdf/sign`, {
    method: "POST",
    body: formData,
  });
  if (!response.ok) throw new Error(await readError(response, "Gagal menambahkan tanda tangan"));
  return await response.blob();
};

export const getPdfThumbnails = async (
  file: File,
): Promise<{ total: number; thumbnails: string[] }> => {
  const formData = new FormData();
  formData.append("pdf", file);

  const response = await fetch(`${API_URL}/pdf/thumbnails`, {
    method: "POST",
    body: formData,
  });
  if (!response.ok) throw new Error(await readError(response, "Gagal memuat halaman PDF"));
  return await response.json();
};

export interface PageSpec {
  page: number; // nomor halaman asli (mulai 1)
  rotate: number; // 0 | 90 | 180 | 270
}

export const reorderPdf = async (
  file: File,
  order: PageSpec[],
): Promise<Blob> => {
  const formData = new FormData();
  formData.append("pdf", file);
  formData.append("order", JSON.stringify(order));

  const response = await fetch(`${API_URL}/pdf/reorder`, {
    method: "POST",
    body: formData,
  });
  if (!response.ok) throw new Error(await readError(response, "Gagal menyimpan susunan halaman"));
  return await response.blob();
};

// Satu file -> JPG, banyak file -> ZIP. Jenis hasil dibaca dari blob.type.
export const heicToJpg = async (
  files: File[],
  quality: number = 90,
): Promise<Blob> => {
  const formData = new FormData();
  files.forEach((file) => formData.append("images", file));
  formData.append("quality", quality.toString());

  const response = await fetch(`${API_URL}/convert/heic-to-jpg`, {
    method: "POST",
    body: formData,
  });
  if (!response.ok) throw new Error(await readError(response, "Gagal mengonversi HEIC"));
  return await response.blob();
};
