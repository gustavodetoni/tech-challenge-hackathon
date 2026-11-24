# MinIO

O serviço `minio-init` usa a AWS CLI somente como cliente S3 compatível e cria idempotentemente os buckets privados `video-inputs` e `video-results` ao subir a stack. Nenhuma política pública é aplicada.
