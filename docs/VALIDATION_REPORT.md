# Validation report

## Executed

| Command | Result |
| --- | --- |
| `go test ./...` in `services/video-api` | Passed |
| `go test ./...` in `services/video-worker` | Passed |
| `go test ./...` in `services/notification-service` | Passed |
| `npm run build` in `frontend` | Passed |
| `docker compose config --quiet` | Passed |

## Integrated validation

An initial Docker Compose bring-up was attempted. The environment could not pull the then-configured MinIO images from their registries; this is recorded as an unresolved external image-registry limitation. Therefore no end-to-end success, retry/DLQ, Mailpit or three-worker claim is made in this report.

## Manual acceptance procedure

1. `cp .env.example .env && docker compose up -d --build`
2. Open the frontend and sign in as `demo` / `demo-local-only`.
3. Upload two short supported videos and wait for `COMPLETED`; download and inspect each ZIP.
4. Open Mailpit at `http://localhost:8025` after an invalid-media processing failure.
5. Run `docker compose up -d --scale video-worker=3` and observe queues in RabbitMQ Management.
