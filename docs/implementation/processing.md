# Fase 3 — Processamento

- A confirmação de upload grava a transição para `PENDING` e um evento de outbox na mesma transação PostgreSQL.
- O worker publica eventos usando a exchange durável `video.processing`, routing key `video.processing.requested`, e consome `video.processing.jobs` com mensagens JSON versionadas que contêm apenas identificadores.
- A extração usa `ffprobe` antes do `ffmpeg`, a 1 FPS. Esta configuração explícita foi adotada porque o projeto-base não estava presente no repositório. Os frames JPG são ordenados e gravados em ZIP por streaming.
