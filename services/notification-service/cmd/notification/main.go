package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"mime"
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

	message := renderFailureEmail(from, to, event)
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

func renderFailureEmail(from, to string, event failure) string {
	const boundary = "fiap-video-notification-boundary"
	reason := "Não foi possível processar o arquivo enviado."
	if event.ErrorCode != "PROCESSING_FAILED" && event.ErrorCode != "" {
		reason = "O processamento do vídeo foi interrompido."
	}
	subject := mime.QEncoding.Encode("UTF-8", "Não foi possível processar seu vídeo")
	plainText := fmt.Sprintf(
		"Olá!\r\n\r\nNão foi possível processar o seu vídeo. Você pode tentar enviá-lo novamente.\r\n\r\nIdentificação do processamento: %s\r\nMotivo: %s\r\n\r\nEquipe FIAP Video\r\n",
		event.JobID, reason,
	)
	htmlBody := fmt.Sprintf(
		"<!doctype html><html lang=\"pt-BR\"><body style=\"margin:0;background:#f4f4f5;font-family:Arial,sans-serif;color:#18181b\">"+
			"<table role=\"presentation\" width=\"100%%\" cellpadding=\"0\" cellspacing=\"0\"><tr><td align=\"center\" style=\"padding:32px 16px\">"+
			"<table role=\"presentation\" width=\"100%%\" cellpadding=\"0\" cellspacing=\"0\" style=\"max-width:560px;background:#fff;border:1px solid #e4e4e7;border-radius:12px;overflow:hidden\">"+
			"<tr><td style=\"padding:24px 28px;background:#18181b;color:#fff;font-weight:700;font-size:18px\">FIAP Video</td></tr>"+
			"<tr><td style=\"padding:28px\"><div style=\"display:inline-block;padding:6px 10px;background:#fef2f2;color:#b91c1c;border-radius:999px;font-size:12px;font-weight:700\">PROCESSAMENTO NÃO CONCLUÍDO</div>"+
			"<h1 style=\"margin:20px 0 10px;font-size:24px;line-height:1.25\">Não foi possível processar seu vídeo</h1>"+
			"<p style=\"margin:0 0 20px;color:#52525b;line-height:1.6\">Encontramos um problema ao transformar o seu vídeo em imagens. Tente enviá-lo novamente.</p>"+
			"<div style=\"padding:16px;background:#f4f4f5;border-radius:8px\"><div style=\"font-size:12px;color:#71717a;margin-bottom:4px\">Identificação do processamento</div>"+
			"<div style=\"font-family:monospace;font-size:13px;word-break:break-all\">%s</div><div style=\"font-size:12px;color:#71717a;margin:16px 0 4px\">Motivo</div><div style=\"font-size:14px\">%s</div></div>"+
			"<p style=\"margin:24px 0 0;color:#52525b;font-size:14px;line-height:1.6\">Se o problema persistir, verifique se o arquivo é um vídeo válido e tente novamente.</p></td></tr>"+
			"<tr><td style=\"padding:18px 28px;border-top:1px solid #e4e4e7;color:#71717a;font-size:12px\">FIAP Video · Notificação automática</td></tr></table></td></tr></table></body></html>",
		html.EscapeString(event.JobID), html.EscapeString(reason),
	)

	return fmt.Sprintf(
		"Subject: %s\r\nFrom: <%s>\r\nTo: <%s>\r\nDate: %s\r\n"+
			"MIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n"+
			"--%s\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n"+
			"--%s\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n"+
			"--%s--\r\n.\r\n",
		subject, from, to, time.Now().Format(time.RFC1123Z), boundary, boundary, plainText, boundary, htmlBody, boundary,
	)
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
