import { Button, Card, CardBody, Chip } from "@heroui/react";
import { useMutation } from "@tanstack/react-query";
import type { VideoJob } from "../api/videos";
import { getDownloadUrl } from "../api/videos";

const statusLabel: Record<string, string> = {
  AWAITING_UPLOAD: "Envio incompleto",
  PENDING: "Na fila",
  PROCESSING: "Processando",
  COMPLETED: "Concluído",
  FAILED: "Falhou",
};

export function VideoList({ jobs }: { jobs: VideoJob[] }) {
  const download = useMutation({
    mutationFn: async (jobId: string) => {
      const url = await getDownloadUrl(jobId);
      window.location.assign(url);
    },
  });

  return (
    <Card className="border border-zinc-200 bg-white shadow-none dark:border-zinc-800 dark:bg-zinc-950">
      <CardBody className="gap-5 p-6 sm:p-7">
        <h2 className="text-lg font-semibold">Meus vídeos</h2>
        {jobs.length === 0 ? (
          <p className="text-default-500">Sua fila está vazia.</p>
        ) : (
          jobs.map((job) => (
            <div
              className="flex items-center justify-between gap-4 border-b border-zinc-200 pb-3 dark:border-zinc-800"
              key={job.id}
            >
              <div className="min-w-0">
                <p className="truncate font-medium">{job.originalFilename}</p>
                <p className="text-xs text-default-500">
                  {new Date(job.createdAt).toLocaleString("pt-BR")}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                {job.status === "COMPLETED" && (
                  <Button
                    className="bg-blue-600 font-medium text-white"
                    isLoading={download.isPending && download.variables === job.id}
                    size="sm"
                    onPress={() => download.mutate(job.id)}
                  >
                    Baixar ZIP
                  </Button>
                )}
                <Chip
                  color={
                    job.status === "COMPLETED"
                      ? "success"
                      : job.status === "FAILED"
                        ? "danger"
                        : "warning"
                  }
                  variant="flat"
                >
                  {statusLabel[job.status] ?? job.status}
                </Chip>
              </div>
            </div>
          ))
        )}
        {download.isError && (
          <p className="text-sm text-danger">{download.error.message}</p>
        )}
      </CardBody>
    </Card>
  );
}
