import { keycloak } from "../../auth/keycloak";
export type VideoJob = {
  id: string;
  originalFilename: string;
  contentType: string;
  sizeBytes: number;
  status: string;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
  errorCode?: string;
  errorMessage?: string;
};
type ApiVideoJob = VideoJob & {
  OriginalFilename?: string;
  ContentType?: string;
  SizeBytes?: number;
  Status?: string;
  CreatedAt?: string;
  StartedAt?: string;
  FinishedAt?: string;
  ErrorCode?: string;
  ErrorMessage?: string;
};
const base = import.meta.env.VITE_API_BASE_URL;
const headers = () => ({ Authorization: `Bearer ${keycloak.token}` });
export async function listVideos(): Promise<VideoJob[]> {
  const r = await fetch(`${base}/api/v1/videos`, { headers: headers() });
  if (!r.ok) throw Error("Não foi possível carregar seus vídeos.");
  const body = await r.json();
  return (body.items ?? []).map(
    (job: ApiVideoJob): VideoJob => ({
      id: job.id,
      originalFilename:
        job.originalFilename ?? job.OriginalFilename ?? "Vídeo sem nome",
      contentType: job.contentType ?? job.ContentType ?? "video",
      sizeBytes: job.sizeBytes ?? job.SizeBytes ?? 0,
      status: job.status ?? job.Status ?? "UNKNOWN",
      createdAt:
        job.createdAt ?? job.CreatedAt ?? new Date(0).toISOString(),
      startedAt: job.startedAt ?? job.StartedAt,
      finishedAt: job.finishedAt ?? job.FinishedAt,
      errorCode: job.errorCode ?? job.ErrorCode,
      errorMessage: job.errorMessage ?? job.ErrorMessage,
    }),
  );
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

export async function getDownloadUrl(jobId: string): Promise<string> {
  const response = await fetch(`${base}/api/v1/videos/${jobId}/download`, {
    headers: headers(),
  });
  if (!response.ok) throw Error("O arquivo ZIP ainda não está disponível.");
  return (await response.json()).downloadUrl as string;
}
