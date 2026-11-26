package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHandlerReportsJobAndOutboxMetrics(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:metrics?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE video_jobs (status TEXT, created_at DATETIME)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE outbox_events (published_at DATETIME)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO video_jobs(status) VALUES ('PENDING'), ('COMPLETED'), ('PENDING')").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO outbox_events(published_at) VALUES (NULL), (CURRENT_TIMESTAMP)").Error; err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	Handler(db).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("got %d", recorder.Code)
	}
	for _, expected := range []string{"fiap_video_jobs{status=\"PENDING\"} 2", "fiap_video_jobs{status=\"COMPLETED\"} 1", "fiap_video_outbox_pending 1"} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("missing %q in %s", expected, recorder.Body.String())
		}
	}
}

func TestHandlerReturnsUnavailableWhenJobsQueryFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:missing?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	Handler(db).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d", recorder.Code)
	}
}
