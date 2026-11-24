# Decisões da Fundação

- O repositório estava vazio, sem projeto-base disponível localmente. Por isso, nenhum algoritmo de extração foi migrado; ele será incorporado somente quando a etapa de processamento permitir.
- O navegador fala apenas com `video-api`. Worker e serviço de notificações têm health endpoints internos para operação, sem portas publicadas.
- PostgreSQL é acessado pela API via GORM. RabbitMQ, MinIO e Keycloak são inicializados para os próximos fluxos, mas nenhum job, mensagem, upload ou autenticação de negócio é criado nesta etapa.
- O cliente Keycloak `fiap-video-spa` é público, usa Authorization Code Flow com PKCE S256 e não contém segredo de cliente.
- `VITE_API_BASE_URL` é uma variável de build: Vite a incorpora no JavaScript gerado. Alterá-la requer reconstruir o container do frontend.
- A imagem comunitária oficial do MinIO deixou de estar disponível no Docker Hub durante a implementação. A stack usa a imagem Bitnami Legacy MinIO, com versão fixada, para manter uma inicialização local reproduzível e compatível com S3.
