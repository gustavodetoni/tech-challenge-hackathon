package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type failure struct {
	EventID   string `json:"eventId"`
	JobID     string `json:"jobId"`
	ErrorCode string `json:"errorCode"`
}

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
			if err := send(os.Getenv("SMTP_ADDR"), os.Getenv("NOTIFICATION_FROM"), envOrDefault("NOTIFICATION_TO", "admin@fiap-video.local"), delivery.Body); err != nil {
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

func send(addr, from, to string, body []byte) error {
	var event failure
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("invalid video.failed event: %w", err)
	}

	connection, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return err
	}
	defer connection.Close()

	message := fmt.Sprintf(
		"Subject: Video processing failed\r\n"+
			"From: <%s>\r\n"+
			"To: <%s>\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n"+
			"Your video could not be processed.\r\n"+
			"Job ID: %s\r\n"+
			"Reason: %s\r\n\r\n"+
			"You can upload the video again.\r\n.\r\n",
		from, to, event.JobID, event.ErrorCode,
	)
	for _, command := range []string{
		"HELO fiap-video\r\n",
		"MAIL FROM:<" + from + ">\r\n",
		"RCPT TO:<" + to + ">\r\n",
		"DATA\r\n",
		message,
		"QUIT\r\n",
	} {
		if _, err = fmt.Fprint(connection, command); err != nil {
			return err
		}
	}
	return nil
}

func health(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "notification_service_up 1\n") })
	server := &http.Server{Addr: envOrDefault("HTTP_ADDR", ":8082"), Handler: mux}
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
