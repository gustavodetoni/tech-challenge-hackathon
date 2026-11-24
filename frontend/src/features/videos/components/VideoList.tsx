import { Card, CardBody, Chip } from "@heroui/react";
import type { VideoJob } from "../api/videos";

export function VideoList({ jobs }: { jobs: VideoJob[] }) {
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
                {job.status}
              </Chip>
            </div>
          ))
        )}
      </CardBody>
    </Card>
  );
}
