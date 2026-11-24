FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/notification-service ./cmd/notification
FROM alpine:3.21
RUN adduser -D -H app && apk add --no-cache wget
COPY --from=build /out/notification-service /notification-service
COPY --from=build /out/notification-service /healthcheck
USER app
EXPOSE 8082
ENTRYPOINT ["/notification-service"]
