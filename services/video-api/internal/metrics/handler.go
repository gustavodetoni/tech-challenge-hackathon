// Package metrics exposes a small, dependency-free Prometheus endpoint.
package metrics

import (
	"fmt"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

// Handler reports the persisted state of the video workflow. Keeping these
// values database-backed means a dashboard remains accurate after an API or
// worker restart.
func Handler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintln(w, "# HELP fiap_video_jobs Number of video jobs grouped by workflow status.")
		fmt.Fprintln(w, "# TYPE fiap_video_jobs gauge")

		var rows []struct {
			Status string
			Count  int64
		}
		if err := db.Table("video_jobs").Select("status, COUNT(*) AS count").Group("status").Scan(&rows).Error; err != nil {
			http.Error(w, "unable to collect video metrics", http.StatusServiceUnavailable)
			return
		}
		for _, row := range rows {
			fmt.Fprintf(w, "fiap_video_jobs{status=%s} %d\n", strconv.Quote(row.Status), row.Count)
		}

		var pendingOutbox int64
		if err := db.Table("outbox_events").Where("published_at IS NULL").Count(&pendingOutbox).Error; err == nil {
			fmt.Fprintln(w, "# HELP fiap_video_outbox_pending Events waiting to be published to RabbitMQ.")
			fmt.Fprintln(w, "# TYPE fiap_video_outbox_pending gauge")
			fmt.Fprintf(w, "fiap_video_outbox_pending %d\n", pendingOutbox)
		}

		var oldestPendingSeconds float64
		if err := db.Raw(`SELECT COALESCE(EXTRACT(EPOCH FROM NOW() - MIN(created_at)), 0)
			FROM video_jobs WHERE status IN ('PENDING', 'PROCESSING')`).Scan(&oldestPendingSeconds).Error; err == nil {
			fmt.Fprintln(w, "# HELP fiap_video_oldest_active_job_seconds Age of the oldest queued or processing job.")
			fmt.Fprintln(w, "# TYPE fiap_video_oldest_active_job_seconds gauge")
			fmt.Fprintf(w, "fiap_video_oldest_active_job_seconds %.3f\n", oldestPendingSeconds)
		}
	}
}
