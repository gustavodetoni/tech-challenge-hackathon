FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/video-api ./cmd/api

FROM alpine:3.21
RUN adduser -D -H app && apk add --no-cache wget
COPY --from=build /out/video-api /video-api
COPY --from=build /out/video-api /healthcheck
USER app
EXPOSE 8080
ENTRYPOINT ["/video-api"]
