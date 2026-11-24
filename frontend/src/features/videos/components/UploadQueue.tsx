import { Button, Card, CardBody, Input, Progress } from "@heroui/react";
import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { uploadVideo } from "../api/videos";

type Item = { id: string; file: File; progress: number; state: string };

export function UploadQueue() {
  const [items, setItems] = useState<Item[]>([]);
  const qc = useQueryClient();
  const mutation = useMutation({
    mutationFn: async () => {
      for (const item of items.filter((x) => x.state === "Pronto")) {
        setItems((q) =>
          q.map((x) => (x.id === item.id ? { ...x, state: "Enviando" } : x)),
        );
        await uploadVideo(item.file, (n) =>
          setItems((q) =>
            q.map((x) => (x.id === item.id ? { ...x, progress: n } : x)),
          ),
        );
        setItems((q) =>
          q.map((x) =>
            x.id === item.id ? { ...x, state: "Na fila", progress: 100 } : x,
          ),
        );
      }
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["videos"] }),
  });
  return (
    <Card className="border border-zinc-200 bg-white shadow-none dark:border-zinc-800 dark:bg-zinc-950">
      <CardBody className="gap-5 p-6 sm:p-7">
        <h2 className="text-lg font-semibold">Novo processamento</h2>
        <Input
          type="file"
          multiple
          accept="video/*,.mkv"
          label="Selecione seus vídeos"
          onChange={(e) =>
            e.target.files &&
            setItems((q) => [
              ...q,
              ...Array.from(e.target.files!).map((file) => ({
                id: crypto.randomUUID(),
                file,
                progress: 0,
                state: "Pronto",
              })),
            ])
          }
        />
        {items.map((item) => (
          <div
            className="space-y-2 border-b border-zinc-200 py-3 dark:border-zinc-800"
            key={item.id}
          >
            <div className="flex justify-between gap-3 text-sm">
              <span className="truncate font-medium">{item.file.name}</span>
              <span className="text-default-500">{item.state}</span>
            </div>
            <Progress value={item.progress} />
          </div>
        ))}
        <Button
          className="bg-blue-600 font-medium text-white"
          isLoading={mutation.isPending}
          isDisabled={!items.some((i) => i.state === "Pronto")}
          onPress={() => mutation.mutate()}
        >
          Enviar fila
        </Button>
        {mutation.isError && (
          <p className="text-danger">{mutation.error.message}</p>
        )}
      </CardBody>
    </Card>
  );
}
