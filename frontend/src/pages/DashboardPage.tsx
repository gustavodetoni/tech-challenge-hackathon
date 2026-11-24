import { Spinner } from "@heroui/react";
import { useQuery } from "@tanstack/react-query";
import { AppHeader } from "../components/layout/AppHeader";
import { VideoList } from "../features/videos/components/VideoList";
import { UploadQueue } from "../features/videos/components/UploadQueue";
import { listVideos } from "../features/videos/api/videos";

export function DashboardPage() {
  const query = useQuery({
    queryKey: ["videos"],
    queryFn: listVideos,
    refetchInterval: (q) =>
      q.state.data?.some((j) =>
        ["AWAITING_UPLOAD", "PENDING", "PROCESSING"].includes(j.status),
      )
        ? 8000
        : false,
  });
  return (
    <main className="min-h-screen bg-zinc-50 text-zinc-950 dark:bg-black dark:text-zinc-50">
      <div className="mx-auto max-w-5xl px-5 py-12 sm:px-8">
        <AppHeader />
        {query.isLoading ? (
          <Spinner />
        ) : (
          <section className="grid gap-6 lg:grid-cols-2">
            <UploadQueue />
            <VideoList jobs={query.data ?? []} />
          </section>
        )}
      </div>
    </main>
  );
}
