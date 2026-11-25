package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
)

type event struct {
	EventID uuid.UUID `json:"eventId"`
	JobID   uuid.UUID `json:"jobId"`
}

type failureEvent struct {
	EventID    uuid.UUID `json:"eventId"`
	EventType  string    `json:"eventType"`
	Version    int       `json:"version"`
	JobID      uuid.UUID `json:"jobId"`
	OccurredAt time.Time `json:"occurredAt"`
	ErrorCode  string    `json:"errorCode"`
}

type workerMetrics struct {
	started        atomic.Uint64
	completed      atomic.Uint64
	failed         atomic.Uint64
	durationMillis atomic.Uint64
}

var metrics workerMetrics

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	ctx := context.Background()
	go health()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	storage, _ := minio.New(envOrDefault("MINIO_INTERNAL_ENDPOINT", "minio:9000"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), ""),
		Secure: false,
	})
	// Publishing outbox events and consuming processing events use different AMQP
	// channels. RabbitMQ channels are not safe for concurrent publish/consume use.
	go outbox(ctx, db, os.Getenv("RABBITMQ_URL"))
	for {
		connection, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		consume(ctx, logger, db, storage, connection)
		connection.Close()
	}
}

func health() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, "# TYPE fiap_video_worker_jobs_started_total counter\nfiap_video_worker_jobs_started_total %d\n", metrics.started.Load())
		fmt.Fprintf(w, "# TYPE fiap_video_worker_jobs_completed_total counter\nfiap_video_worker_jobs_completed_total %d\n", metrics.completed.Load())
		fmt.Fprintf(w, "# TYPE fiap_video_worker_jobs_failed_total counter\nfiap_video_worker_jobs_failed_total %d\n", metrics.failed.Load())
		fmt.Fprintf(w, "# TYPE fiap_video_worker_processing_duration_milliseconds_total counter\nfiap_video_worker_processing_duration_milliseconds_total %d\n", metrics.durationMillis.Load())
	})
	_ = http.ListenAndServe(envOrDefault("HTTP_ADDR", ":8081"), mux)
}

func consume(ctx context.Context, logger *slog.Logger, db *pgxpool.Pool, storage *minio.Client, connection *amqp.Connection) {
	channel, _ := connection.Channel()
	defer channel.Close()
	_ = channel.ExchangeDeclare("video.processing", "direct", true, false, false, false, nil)
	queue, _ := channel.QueueDeclare("video.processing.jobs", true, false, false, false, nil)
	_ = channel.QueueBind(queue.Name, "video.processing.requested", "video.processing", false, nil)
	_ = channel.Qos(1, 0, false)

	deliveries, _ := channel.Consume(queue.Name, "", false, false, false, false, nil)
	for delivery := range deliveries {
		var message event
		if json.Unmarshal(delivery.Body, &message) != nil {
			_ = delivery.Nack(false, false)
			continue
		}

		startedAt := time.Now()
		claimed, err := process(ctx, db, storage, message.JobID)
		if !claimed {
			_ = delivery.Ack(false)
			continue
		}
		metrics.started.Add(1)
		metrics.durationMillis.Add(uint64(time.Since(startedAt).Milliseconds()))
		if err != nil {
			logger.Error("processing failed", "job_id", message.JobID, "error", err)
			metrics.failed.Add(1)
			if markErr := markFailed(ctx, db, message.JobID); markErr != nil {
				logger.Error("could not persist processing failure", "job_id", message.JobID, "error", markErr)
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Nack(false, false)
			continue
		}
		metrics.completed.Add(1)
		_ = delivery.Ack(false)
	}
}

func outbox(ctx context.Context, db *pgxpool.Pool, rabbitURL string) {
	for {
		connection, err := amqp.Dial(rabbitURL)
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
		_ = channel.ExchangeDeclare("video.processing", "direct", true, false, false, false, nil)
		_ = channel.ExchangeDeclare("video.events", "direct", true, false, false, false, nil)
		if err = channel.Confirm(false); err != nil {
			_ = channel.Close()
			_ = connection.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 10))

		// Keep this publishing channel alive until RabbitMQ drops it. The outer
		// loop will then reconnect and retry only events without published_at.
		for ctx.Err() == nil {
			rows, err := db.Query(ctx, "SELECT id, type, payload FROM outbox_events WHERE published_at IS NULL ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 10")
			if err == nil {
				for rows.Next() {
					var id uuid.UUID
					var eventType string
					var payload []byte
					_ = rows.Scan(&id, &eventType, &payload)
					exchange := "video.processing"
					if eventType == "video.failed" {
						exchange = "video.events"
					}
					err = channel.PublishWithContext(ctx, exchange, eventType, true, false, amqp.Publishing{Body: payload, ContentType: "application/json", DeliveryMode: amqp.Persistent, MessageId: id.String()})
					if err == nil {
						select {
						case confirmation := <-confirms:
							if confirmation.Ack {
								_, _ = db.Exec(ctx, "UPDATE outbox_events SET published_at=NOW(), attempts=attempts+1 WHERE id=$1", id)
							}
						case <-time.After(5 * time.Second):
							// Leave the event unpublished to retry safely on the next loop.
						}
					}
				}
				rows.Close()
			}
			time.Sleep(time.Second)
		}
		_ = channel.Close()
		_ = connection.Close()
	}
}

// process claims a PENDING job. claimed=false means a duplicate message that
// has already been handled, so the consumer can acknowledge it safely.
func process(ctx context.Context, db *pgxpool.Pool, storage *minio.Client, id uuid.UUID) (claimed bool, err error) {
	var key, name string
	err = db.QueryRow(ctx, "UPDATE video_jobs SET status='PROCESSING', started_at=NOW(), attempt_count=attempt_count+1 WHERE id=$1 AND status='PENDING' RETURNING input_object_key, original_filename", id).Scan(&key, &name)
	if err != nil {
		return false, nil
	}
	if prefix := os.Getenv("WORKER_FORCE_FAILURE_PREFIX"); prefix != "" && strings.HasPrefix(name, prefix) {
		return true, errors.New("forced processing failure for local notification test")
	}

	directory, err := os.MkdirTemp("", "video-")
	if err != nil {
		return true, err
	}
	defer os.RemoveAll(directory)
	input := filepath.Join(directory, "source"+filepath.Ext(name))
	if err = storage.FGetObject(ctx, "video-inputs", key, input, minio.GetObjectOptions{}); err != nil {
		return true, err
	}
	if err = exec.CommandContext(ctx, "ffprobe", "-v", "error", input).Run(); err != nil {
		return true, err
	}
	pattern := filepath.Join(directory, "frame-%06d.jpg")
	if err = exec.CommandContext(ctx, "ffmpeg", "-y", "-i", input, "-vf", "fps=1", pattern).Run(); err != nil {
		return true, err
	}
	frames, _ := filepath.Glob(filepath.Join(directory, "frame-*.jpg"))
	if len(frames) == 0 {
		return true, os.ErrNotExist
	}
	sort.Strings(frames)
	archivePath := filepath.Join(directory, "frames.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		return true, err
	}
	archive := zip.NewWriter(file)
	for _, frame := range frames {
		entry, entryErr := archive.Create(filepath.Base(frame))
		if entryErr != nil {
			_ = archive.Close()
			_ = file.Close()
			return true, entryErr
		}
		source, sourceErr := os.Open(frame)
		if sourceErr != nil {
			_ = archive.Close()
			_ = file.Close()
			return true, sourceErr
		}
		_, _ = io.Copy(entry, source)
		_ = source.Close()
	}
	if err = archive.Close(); err != nil {
		_ = file.Close()
		return true, err
	}
	if err = file.Close(); err != nil {
		return true, err
	}
	output := "results/" + id.String() + "/frames.zip"
	if _, err = storage.FPutObject(ctx, "video-results", output, archivePath, minio.PutObjectOptions{ContentType: "application/zip"}); err != nil {
		return true, err
	}
	_, err = db.Exec(ctx, "UPDATE video_jobs SET status='COMPLETED', output_object_key=$2, finished_at=NOW() WHERE id=$1 AND status='PROCESSING'", id, output)
	return true, err
}

func markFailed(ctx context.Context, db *pgxpool.Pool, jobID uuid.UUID) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, "UPDATE video_jobs SET status='FAILED', error_code='PROCESSING_FAILED', error_message='Processing could not be completed', finished_at=NOW() WHERE id=$1 AND status='PROCESSING'", jobID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return nil
	}
	eventID := uuid.New()
	payload, err := json.Marshal(failureEvent{
		EventID: eventID, EventType: "video.failed", Version: 1, JobID: jobID,
		OccurredAt: time.Now().UTC(), ErrorCode: "PROCESSING_FAILED",
	})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO outbox_events (id, aggregate_id, type, payload) VALUES ($1, $2, $3, $4)", eventID, jobID, "video.failed", payload); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
