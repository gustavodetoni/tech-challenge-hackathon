# Fase 2 — Autenticação e vídeos

- A API valida access tokens RS256 por JWKS interno do Keycloak e compara o `iss` com a URL pública que o navegador recebe. O cache de chaves dura cinco minutos e é renovado imediatamente se um `kid` novo aparecer.
- URLs do MinIO possuem dois endpoints: `minio:9000` para operações internas e `localhost:9000` para URLs presigned usadas pelo navegador. A assinatura, portanto, aponta para um host que o browser consegue alcançar.
- A migração SQL é aplicada uma única vez, registrada em `schema_migrations`; a API não usa `AutoMigrate`.
- O usuário de desenvolvimento do realm é `demo` com senha `demo-local-only`. Ele é exclusivamente local e não deve ser reutilizado fora deste ambiente.
- Depois de alterar o JSON do realm em um ambiente já iniciado, execute `docker compose down -v` para que o Keycloak importe novamente o realm. Isso remove dados locais.
