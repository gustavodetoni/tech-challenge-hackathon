import { Card, CardContent } from "@/components/ui/card";
import { CircleAlert, CircleCheck, Video } from "lucide-react";
import type { VideoJob } from "../api/videos";

type StatCardProps = {
  label: string;
  value: number;
  icon: typeof Video;
};

function StatCard({ label, value, icon: Icon }: StatCardProps) {
  return (
    <Card className="border border-zinc-200/90 bg-white py-0 shadow-sm shadow-zinc-200/50">
      <CardContent className="flex flex-row items-center gap-4 p-5">
        <div className="grid size-14 place-items-center rounded-2xl bg-blue-50 text-blue-600">
          <Icon aria-hidden="true" className="size-7" strokeWidth={2.2} />
        </div>
        <div>
          <p className="text-3xl font-semibold tracking-tight text-zinc-950">{value}</p>
          <p className="text-sm font-medium text-zinc-500">{label}</p>
        </div>
      </CardContent>
    </Card>
  );
}

export function VideoStats({ jobs }: { jobs: VideoJob[] }) {
  return (
    <section aria-label="Resumo dos processamentos" className="grid gap-5 md:grid-cols-3">
      <StatCard icon={Video} label="Total de vídeos" value={jobs.length} />
      <StatCard
        icon={CircleCheck}
        label="Concluídos"
        value={jobs.filter((job) => job.status === "COMPLETED").length}
      />
      <StatCard
        icon={CircleAlert}
        label="Falhas"
        value={jobs.filter((job) => job.status === "FAILED").length}
      />
    </section>
  );
}
