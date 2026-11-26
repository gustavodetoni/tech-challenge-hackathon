import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { keycloak } from "../features/auth/keycloak";

export function LoginPage() {
  const navigate = useNavigate();

  useEffect(() => {
    if (keycloak.authenticated) {
      void navigate({ to: "/dashboard", replace: true });
    }
  }, [navigate]);

  return (
    <main className="grid min-h-screen place-items-center bg-zinc-50 p-6 text-zinc-950">
      <Card className="w-full max-w-md border border-zinc-200 bg-white py-0 shadow-sm">
        <CardContent className="flex flex-col gap-7 p-8 sm:p-10">
          <p className="text-xs font-semibold tracking-[.18em] text-blue-600">
            FIAP X
          </p>
          <h1 className="text-3xl font-semibold tracking-tight">Vídeos em imagens.</h1>
          <p className="leading-6 text-zinc-500">
            Envie, acompanhe e baixe seus resultados com segurança.
          </p>
          <Button
            className="bg-blue-600 font-medium text-white"
            size="lg"
            onClick={() =>
              keycloak.authenticated
                ? void navigate({ to: "/dashboard" })
                : keycloak.login()
            }
          >
            Entrar na plataforma
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}
