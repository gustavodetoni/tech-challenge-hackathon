import { LoaderCircle } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { AppHeader } from "../components/layout/AppHeader";
import { VideoList } from "../features/videos/components/VideoList";
import { UploadQueue } from "../features/videos/components/UploadQueue";
import { VideoStats } from "../features/videos/components/VideoStats";
import { listVideos } from "../features/videos/api/videos";

export function DashboardPage() {
  const query = useQuery({
    queryKey: ["videos"],
    queryFn: listVideos,
    refetchInterval: 8000,
  });
  return (
    <main className="min-h-screen bg-zinc-50 text-zinc-950">
      <div className="mx-auto max-w-6xl px-5 py-8 sm:px-8 lg:py-10">
        <AppHeader />
        {query.isLoading ? (
          <div className="grid min-h-64 place-items-center">
            <LoaderCircle aria-label="Carregando vídeos" className="size-6 animate-spin text-blue-600" />
          </div>
        ) : (
          <div className="space-y-8">
            <VideoStats jobs={query.data ?? []} />
            <section className="space-y-7">
              <UploadQueue />
              <VideoList jobs={query.data ?? []} />
            </section>
          </div>
        )}
      </div>
    </main>
  );
}
