import { Button } from "@heroui/react";
import { keycloak } from "../../features/auth/keycloak";

export function AppHeader() {
  return (
    <header className="mb-10 flex flex-col justify-between gap-5 border-b border-zinc-200 pb-8 dark:border-zinc-800 sm:flex-row sm:items-start">
      <div>
        <p className="text-xs font-semibold tracking-[.18em] text-blue-600">
          PAINEL DE PROCESSAMENTO
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">
          Olá,{" "}
          {(keycloak.tokenParsed?.preferred_username as string) || "usuário"}.
        </h1>
        <p className="mt-2 text-zinc-500 dark:text-zinc-400">
          Adicione vários vídeos e acompanhe a fila em tempo real.
        </p>
      </div>
      <Button
        variant="bordered"
        onPress={() => keycloak.logout({ redirectUri: location.origin })}
      >
        Sair
      </Button>
    </header>
  );
}
