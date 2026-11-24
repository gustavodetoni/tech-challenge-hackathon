# Fase 4 — Resiliência

Topologia RabbitMQ: `video.processing` recebe eventos; a fila principal é `video.processing.jobs`. Falhas transitórias devem ser encaminhadas para filas de retry com TTL e retornam pela DLX; erros permanentes e mensagens esgotadas vão para `video.processing.dlq`. A DLQ deve ser inspecionada no RabbitMQ Management e republicada manualmente somente após corrigir a causa.

Mailpit é acessível em `http://localhost:8025`. Ele é somente para desenvolvimento local; o serviço de notificações não bloqueia a persistência ou o download do resultado.
