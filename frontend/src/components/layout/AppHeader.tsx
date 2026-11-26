import { Button } from "@/components/ui/button";
import { LogOut } from "lucide-react";
import { keycloak } from "../../features/auth/keycloak";

export function AppHeader() {
  const claims = keycloak.tokenParsed as { name?: string; given_name?: string; preferred_username?: string } | undefined;
  const displayName =
    localStorage.getItem("fiap-video.display-name") ||
    claims?.name ||
    claims?.given_name ||
    claims?.preferred_username ||
    "usuário";

  return (
    <header className="mb-7 flex flex-col justify-between gap-5 border-b border-zinc-200 pb-7 sm:flex-row sm:items-start">
      <div>
        <p className="text-xs font-semibold tracking-[.18em] text-blue-600">
          PAINEL DE PROCESSAMENTO
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">
          Olá, {displayName}.
        </h1>
        <p className="mt-1.5 text-zinc-500">
          Gerencie seus vídeos e acompanhe seus processamentos.
        </p>
      </div>
      <Button
        variant="outline"
        onClick={() => {
          localStorage.removeItem("fiap-video.display-name");
          void keycloak.logout({ redirectUri: location.origin });
        }}
      >
        <LogOut aria-hidden="true" className="size-4" />
        Sair
      </Button>
    </header>
  );
}
