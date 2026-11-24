import { Button, Card, CardBody } from "@heroui/react";
import { keycloak } from "../features/auth/keycloak";

export function LoginPage() {
  return (
    <main className="grid min-h-screen place-items-center bg-zinc-50 p-6 text-zinc-950 dark:bg-black dark:text-zinc-50">
      <Card className="w-full max-w-md border border-zinc-200 bg-white shadow-none dark:border-zinc-800 dark:bg-zinc-950">
        <CardBody className="gap-7 p-8 sm:p-10">
          <p className="text-xs font-semibold tracking-[.18em] text-blue-600">
            FIAP X
          </p>
          <h1 className="text-3xl font-semibold tracking-tight">Vídeos em imagens.</h1>
          <p className="leading-6 text-zinc-500 dark:text-zinc-400">
            Envie, acompanhe e baixe seus resultados com segurança.
          </p>
          <Button className="bg-blue-600 font-medium text-white" size="lg" onPress={() => keycloak.login()}>
            Entrar na plataforma
          </Button>
        </CardBody>
      </Card>
    </main>
  );
}
