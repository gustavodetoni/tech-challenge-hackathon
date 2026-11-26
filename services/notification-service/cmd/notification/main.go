package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fiap-x/video-platform/notification-service/internal/notification"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	go health(ctx)

	seen := sync.Map{}
	for ctx.Err() == nil {
		connection, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		channel, err := connection.Channel()
		if err != nil {
			_ = connection.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		_ = channel.ExchangeDeclare("video.events", "direct", true, false, false, false, nil)
		queue, _ := channel.QueueDeclare("video.notifications.failed", true, false, false, false, nil)
		_ = channel.QueueBind(queue.Name, "video.failed", "video.events", false, nil)
		deliveries, _ := channel.Consume(queue.Name, "", false, false, false, false, nil)
		for delivery := range deliveries {
			key := delivery.MessageId
			if key == "" {
				key = string(delivery.Body)
			}
			if _, alreadySent := seen.LoadOrStore(key, true); alreadySent {
				_ = delivery.Ack(false)
				continue
			}
			if err := notification.Send(os.Getenv("SMTP_ADDR"), os.Getenv("NOTIFICATION_FROM"), envOrDefault("NOTIFICATION_TO", "admin@fiap-video.local"), delivery.Body); err != nil {
				seen.Delete(key)
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
		_ = channel.Close()
		_ = connection.Close()
	}
}

func health(ctx context.Context) {
	server := &http.Server{Addr: envOrDefault("HTTP_ADDR", ":8082"), Handler: notification.HealthHandler()}
	go server.ListenAndServe()
	<-ctx.Done()
	_ = server.Shutdown(context.Background())
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
