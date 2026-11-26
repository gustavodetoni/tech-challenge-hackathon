import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  redirect,
} from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { LoaderCircle } from "lucide-react";
import { keycloak } from "../features/auth/keycloak";
import { LoginPage } from "../pages/LoginPage";
import { DashboardPage } from "../pages/DashboardPage";

function Root() {
  const [ready, setReady] = useState(false);
  useEffect(() => {
    localStorage.removeItem("fiap-video.display-name");
    keycloak
      .init({ onLoad: "check-sso", pkceMethod: "S256" })
      .then(async (authenticated) => {
        if (!authenticated) return;
        const profile = await keycloak.loadUserProfile();
        const displayName = profile.firstName || profile.username;
        if (displayName) localStorage.setItem("fiap-video.display-name", displayName);
      })
      .finally(() => setReady(true));
  }, []);
  return ready ? (
    <Outlet />
  ) : (
    <main className="grid min-h-screen place-items-center bg-zinc-50">
      <LoaderCircle aria-label="Carregando aplicação" className="size-6 animate-spin text-blue-600" />
    </main>
  );
}

const root = createRootRoute({ component: Root });
const login = createRoute({
  getParentRoute: () => root,
  path: "/",
  component: LoginPage,
});

const dashboard = createRoute({
  getParentRoute: () => root,
  path: "/dashboard",
  beforeLoad: () => {
    if (!keycloak.authenticated) throw redirect({ to: "/" });
  },
  component: DashboardPage,
});

const tree = root.addChildren([login, dashboard]);

export const router = createRouter({ routeTree: tree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
