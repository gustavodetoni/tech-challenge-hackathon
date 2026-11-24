FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/video-worker ./cmd/worker
FROM alpine:3.21
RUN adduser -D -H app && apk add --no-cache wget
COPY --from=build /out/video-worker /video-worker
COPY --from=build /out/video-worker /healthcheck
USER app
EXPOSE 8081
ENTRYPOINT ["/video-worker"]
