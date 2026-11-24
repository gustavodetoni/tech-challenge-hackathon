# ADR 0001 — Arquitetura local do processamento

## Decisão

A API pública usa Gin; worker e notificações usam apenas Go e `net/http` para healthchecks. Keycloak fornece OIDC/JWT local, PostgreSQL é a fonte de verdade, MinIO guarda objetos privados e RabbitMQ transporta eventos de processamento.

O padrão Transactional Outbox garante que a confirmação do upload persista o job e o evento juntos. A entrega é at-least-once; o consumidor protege a transição `PENDING` para `PROCESSING` no banco.

## Consequências

Não há API Gateway, serviço de nuvem ou segredo de produção. Workers podem ser escalados com `docker compose up -d --build --scale video-worker=3`.
