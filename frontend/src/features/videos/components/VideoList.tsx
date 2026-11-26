import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useMemo, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import {
  CalendarDays,
  CheckCircle2,
  CircleAlert,
  Clock3,
  Download,
  FileVideo,
  ListOrdered,
  LoaderCircle,
  Timer,
  UploadCloud,
} from "lucide-react";
import type { VideoJob } from "../api/videos";
import { getDownloadUrl } from "../api/videos";

const statusLabel: Record<string, string> = {
  AWAITING_UPLOAD: "Aguardando envio",
  PENDING: "Na fila",
  PROCESSING: "Processando",
  COMPLETED: "Concluído",
  FAILED: "Falhou",
};
const filterLabel: Record<string, string> = {
  ALL: "Todos",
  PENDING: "Na fila",
  PROCESSING: "Processando",
  COMPLETED: "Concluídos",
  FAILED: "Falhas",
};

const statusPresentation: Record<
  string,
  { icon: typeof Clock3; badge: string; iconClassName?: string }
> = {
  COMPLETED: { icon: CheckCircle2, badge: "bg-green-100 text-green-700" },
  FAILED: { icon: CircleAlert, badge: "bg-red-100 text-red-700" },
  PENDING: { icon: ListOrdered, badge: "bg-amber-100 text-amber-700" },
  PROCESSING: {
    icon: LoaderCircle,
    badge: "bg-blue-100 text-blue-700",
    iconClassName: "animate-spin",
  },
  AWAITING_UPLOAD: { icon: UploadCloud, badge: "bg-zinc-100 text-zinc-600" },
};

function formatSize(bytes: number) {
  if (bytes <= 0) return "Tamanho indisponível";
  return new Intl.NumberFormat("pt-BR", {
    maximumFractionDigits: 2,
  }).format(bytes / 1024 / 1024) + " MB";
}

function formatDuration(job: VideoJob) {
  if (!job.startedAt || !job.finishedAt) {
    return job.status === "PROCESSING" ? "Em processamento" : "Não disponível";
  }
  const seconds = Math.max(
    0,
    (new Date(job.finishedAt).getTime() - new Date(job.startedAt).getTime()) / 1000,
  );
  return seconds < 1 ? "< 1 s" : Math.round(seconds) + " s";
}

function formatError(job: VideoJob) {
  if (job.errorCode === "PROCESSING_FAILED") {
    return "Não foi possível processar este vídeo.";
  }
  return job.errorMessage ?? "O processamento não foi concluído.";
}

export function VideoList({ jobs }: { jobs: VideoJob[] }) {
  const [filter, setFilter] = useState("ALL");
  const download = useMutation({
    mutationFn: async (jobId: string) => {
      const url = await getDownloadUrl(jobId);
      window.location.assign(url);
    },
  });
  const filteredJobs = useMemo(
    () => jobs.filter((job) => filter === "ALL" || job.status === filter),
    [filter, jobs],
  );

  return (
    <Card className="border border-zinc-200 bg-white py-0 shadow-sm shadow-zinc-200/50">
      <CardContent className="flex flex-col p-6">
        <div className="flex flex-col gap-3 border-b border-zinc-100 pb-5 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h2 className="text-2xl font-semibold tracking-tight">Meus vídeos</h2>
            <p className="mt-1 text-sm text-zinc-500">Histórico de processamento dos seus vídeos.</p>
          </div>
          <Select
            value={filter}
            onValueChange={(value) => value && setFilter(value)}
          >
            <SelectTrigger aria-label="Filtrar vídeos por status" className="w-36 bg-white">
              <span>{filterLabel[filter] ?? "Todos"}</span>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">Todos</SelectItem>
              <SelectItem value="PENDING">Na fila</SelectItem>
              <SelectItem value="PROCESSING">Processando</SelectItem>
              <SelectItem value="COMPLETED">Concluídos</SelectItem>
              <SelectItem value="FAILED">Falhas</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {filteredJobs.length === 0 ? (
          <div className="grid min-h-48 place-items-center text-center">
            <div>
              <div className="mx-auto mb-3 grid size-11 place-items-center rounded-xl bg-zinc-100 text-zinc-400">
                <FileVideo aria-hidden="true" className="size-5" />
              </div>
              <p className="font-medium text-zinc-700">Nenhum vídeo encontrado</p>
              <p className="mt-1 text-sm text-zinc-500">Envie um vídeo para iniciar o processamento.</p>
            </div>
          </div>
        ) : (
          <div className="divide-y divide-zinc-100">
            {filteredJobs.map((job) => {
              const presentation = statusPresentation[job.status] ?? {
                icon: Clock3,
                badge: "bg-zinc-100 text-zinc-600",
              };
              const StatusIcon = presentation.icon;
              return (
                <div className="grid gap-4 py-5 lg:grid-cols-[minmax(0,1fr)_minmax(11rem,0.55fr)_auto] lg:items-center" key={job.id}>
                  <div className="flex min-w-0 items-start gap-4">
                    <div className="grid size-12 shrink-0 place-items-center rounded-xl bg-blue-50 text-blue-600">
                      <FileVideo aria-hidden="true" className="size-5" />
                    </div>
                    <div className="min-w-0">
                      <p className="truncate font-semibold text-zinc-900">{job.originalFilename}</p>
                      <p className="mt-0.5 text-sm text-zinc-500">
                        {formatSize(job.sizeBytes)} · {job.contentType}
                      </p>
                      <div className="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-zinc-500">
                        <span className="inline-flex items-center gap-1.5">
                          <CalendarDays aria-hidden="true" className="size-3.5" />
                          Criado em: {new Date(job.createdAt).toLocaleString("pt-BR")}
                        </span>
                        <span className="inline-flex items-center gap-1.5">
                          <Timer aria-hidden="true" className="size-3.5" />
                          Tempo de processamento: {formatDuration(job)}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="flex flex-col items-start gap-2 lg:items-start">
                    {job.status === "FAILED" ? (
                      <Tooltip>
                        <TooltipTrigger render={<span className="inline-flex" />}>
                          <Badge className={"h-7 cursor-help gap-1.5 rounded-full px-2.5 font-semibold " + presentation.badge}>
                            <StatusIcon aria-hidden="true" className="size-4" />
                            {statusLabel[job.status]}
                          </Badge>
                        </TooltipTrigger>
                        <TooltipContent className="max-w-64 leading-5">
                          {formatError(job)}
                        </TooltipContent>
                      </Tooltip>
                    ) : (
                      <Badge className={"h-7 gap-1.5 rounded-full px-2.5 font-semibold " + presentation.badge}>
                        <StatusIcon aria-hidden="true" className={"size-4 " + (presentation.iconClassName ?? "")} />
                        {statusLabel[job.status] ?? job.status}
                      </Badge>
                    )}
                  </div>
                  <div className="flex shrink-0 justify-start lg:justify-end">
                    <Button
                      className="border border-zinc-200 bg-white font-semibold text-blue-600 hover:bg-blue-50 disabled:text-zinc-400"
                      size="sm"
                      variant="outline"
                      disabled={job.status !== "COMPLETED" || (download.isPending && download.variables === job.id)}
                      onClick={() => download.mutate(job.id)}
                    >
                      <Download aria-hidden="true" className="size-4" />
                      Baixar ZIP
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
        {download.isError && <p className="mt-5 text-sm font-medium text-red-600">{download.error.message}</p>}
      </CardContent>
    </Card>
  );
}
