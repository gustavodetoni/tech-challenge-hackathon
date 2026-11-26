import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useState, type ChangeEvent, type DragEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { FileVideo, Upload, UploadCloud } from "lucide-react";
import { uploadVideo } from "../api/videos";

type Item = { id: string; file: File; progress: number; state: string };

function newItems(files: FileList | File[]) {
  return Array.from(files).map((file) => ({
    id: crypto.randomUUID(),
    file,
    progress: 0,
    state: "Pronto",
  }));
}

export function UploadQueue() {
  const [items, setItems] = useState<Item[]>([]);
  const [inputKey, setInputKey] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const qc = useQueryClient();

  const addFiles = (files: FileList | File[]) => {
    setItems((queue) => [...queue, ...newItems(files)]);
  };

  const mutation = useMutation({
    mutationFn: async () => {
      for (const item of items.filter((entry) => entry.state === "Pronto")) {
        setItems((queue) =>
          queue.map((entry) =>
            entry.id === item.id ? { ...entry, state: "Enviando" } : entry,
          ),
        );
        try {
          await uploadVideo(item.file, (progress) =>
            setItems((queue) =>
              queue.map((entry) =>
                entry.id === item.id ? { ...entry, progress } : entry,
              ),
            ),
          );
          setItems((queue) => queue.filter((entry) => entry.id !== item.id));
          setInputKey((key) => key + 1);
          await qc.invalidateQueries({ queryKey: ["videos"] });
        } catch (error) {
          setItems((queue) =>
            queue.map((entry) =>
              entry.id === item.id
                ? { ...entry, state: "Erro ao enviar" }
                : entry,
            ),
          );
          throw error;
        }
      }
    },
  });

  const onInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (event.target.files?.length) addFiles(event.target.files);
  };

  const onDrop = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    setIsDragging(false);
    if (event.dataTransfer.files.length) addFiles(event.dataTransfer.files);
  };

  return (
    <Card className="border border-zinc-200 bg-white py-0 shadow-sm shadow-zinc-200/50">
      <CardContent className="flex flex-col gap-5 p-6">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Novo processamento</h2>
          <p className="mt-1 text-sm text-zinc-500">
            Adicione os vídeos que deseja processar.
          </p>
        </div>

        <label
          className={[
            "group grid min-h-52 cursor-pointer place-items-center rounded-2xl border-2 border-dashed px-5 py-7 text-center transition",
            isDragging
              ? "border-blue-500 bg-blue-50"
              : "border-blue-200 bg-blue-50/40 hover:border-blue-400 hover:bg-blue-50",
          ].join(" ")}
          onDragEnter={(event) => {
            event.preventDefault();
            setIsDragging(true);
          }}
          onDragLeave={() => setIsDragging(false)}
          onDragOver={(event) => event.preventDefault()}
          onDrop={onDrop}
        >
          <input
            key={inputKey}
            className="sr-only"
            accept="video/*,.mkv"
            multiple
            type="file"
            onChange={onInputChange}
          />
          <div>
            <div className="mx-auto mb-4 grid size-14 place-items-center rounded-2xl bg-white text-blue-600 shadow-sm">
              <UploadCloud aria-hidden="true" className="size-8" strokeWidth={2} />
            </div>
            <p className="font-semibold text-zinc-900">Arraste e solte seus vídeos aqui</p>
            <p className="mt-1 text-sm font-medium text-blue-600">
              ou clique para selecionar arquivos
            </p>
            <span className="mt-4 inline-flex rounded-full bg-blue-100 px-3 py-1 text-xs font-semibold text-blue-700">
              MP4, MOV, AVI, MKV e WebM
            </span>
          </div>
        </label>

        {items.length > 0 && (
          <div className="divide-y divide-zinc-100">
            {items.map((item) => (
              <div className="py-3" key={item.id}>
                <div className="flex items-center justify-between gap-3 text-sm">
                  <div className="flex min-w-0 items-center gap-2">
                    <FileVideo aria-hidden="true" className="size-4 shrink-0 text-blue-600" />
                    <span className="truncate font-medium text-zinc-800">{item.file.name}</span>
                  </div>
                  <span className="shrink-0 text-zinc-500">{item.state}</span>
                </div>
                {(item.state === "Enviando" || item.progress > 0) && (
                  <Progress
                    aria-label={"Progresso de " + item.file.name}
                    className={
                      "mt-2 " +
                      (item.state === "Erro ao enviar"
                        ? "[&_[data-slot=progress-indicator]]:bg-red-600"
                        : "")
                    }
                    value={item.progress}
                  />
                )}
              </div>
            ))}
          </div>
        )}

        <Button
          className="h-12 bg-blue-600 text-base font-semibold text-white shadow-lg shadow-blue-600/20"
          disabled={!items.some((item) => item.state === "Pronto") || mutation.isPending}
          onClick={() => mutation.mutate()}
        >
          <Upload aria-hidden="true" className="size-5" />
          Enviar para processamento
        </Button>
        {mutation.isError && (
          <p className="text-sm font-medium text-red-600">{mutation.error.message}</p>
        )}
      </CardContent>
    </Card>
  );
}
