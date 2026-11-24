import { keycloak } from "../../auth/keycloak";
export type VideoJob = {
  id: string;
  originalFilename: string;
  status: string;
  createdAt: string;
};
const base = import.meta.env.VITE_API_BASE_URL;
const headers = () => ({ Authorization: `Bearer ${keycloak.token}` });
export async function listVideos() {
  const r = await fetch(`${base}/api/v1/videos`, { headers: headers() });
  if (!r.ok) throw Error("Não foi possível carregar seus vídeos.");
  return ((await r.json()).items ?? []) as VideoJob[];
}
export async function uploadVideo(
  file: File,
  onProgress: (value: number) => void,
) {
  const r = await fetch(`${base}/api/v1/videos/uploads`, {
    method: "POST",
    headers: { ...headers(), "Content-Type": "application/json" },
    body: JSON.stringify({
      filename: file.name,
      contentType: file.type,
      sizeBytes: file.size,
    }),
  });
  if (!r.ok) throw Error("Arquivo inválido ou indisponível.");
  const data = await r.json();
  await new Promise<void>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", data.uploadUrl);
    xhr.setRequestHeader("Content-Type", file.type);
    xhr.upload.onprogress = (e) =>
      e.lengthComputable && onProgress(Math.round((e.loaded / e.total) * 100));
    xhr.onload = () =>
      xhr.status < 300 ? resolve() : reject(Error("Falha no armazenamento."));
    xhr.onerror = () => reject(Error("Falha de rede."));
    xhr.send(file);
  });
  const confirm = await fetch(`${base}/api/v1/videos/${data.jobId}/confirm`, {
    method: "POST",
    headers: headers(),
  });
  if (!confirm.ok) throw Error("Não foi possível confirmar o vídeo.");
}
