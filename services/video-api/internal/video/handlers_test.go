package video

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fiap-x/video-platform/video-api/internal/config"
	"github.com/fiap-x/video-platform/video-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestAPI(t *testing.T) (*API, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Job{}, &OutboxEvent{}); err != nil {
		t.Fatal(err)
	}
	api, err := New(db, config.Config{
		MinioInternalEndpoint: "127.0.0.1:1", MinioPublicEndpoint: "127.0.0.1:1",
		MinioAccessKey: "key", MinioSecretKey: "secret", InputBucket: "input",
		ResultBucket: "result", MaxUploadBytes: 100, PresignTTL: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	return api, db
}

func requestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set(middleware.SubjectKey, "user-1")
	return ctx, recorder
}

func TestCreateRejectsInvalidUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _ := newTestAPI(t)
	ctx, recorder := requestContext(http.MethodPost, "/uploads", []byte("{\"filename\":\"bad.exe\",\"contentType\":\"application/octet-stream\",\"sizeBytes\":1}"))
	api.Create(ctx)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got %d", recorder.Code)
	}
}

func TestCreatePersistsJobAndReturnsPresignedUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newTestAPI(t)
	ctx, recorder := requestContext(http.MethodPost, "/uploads", []byte("{\"filename\":\"movie.mp4\",\"contentType\":\"video/mp4\",\"sizeBytes\":10}"))
	api.Create(ctx)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("body=%s", recorder.Body.String())
	}
	var body struct {
		JobID     string `json:"jobId"`
		UploadURL string `json:"uploadUrl"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.JobID == "" || body.UploadURL == "" {
		t.Fatalf("unexpected response: %#v", body)
	}
	var count int64
	if err := db.Model(&Job{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestListGetAndDownloadState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newTestAPI(t)
	output := "results/job/frames.zip"
	job := Job{ID: uuid.New(), UserID: "user-1", OriginalFilename: "done.mp4", ContentType: "video/mp4", SizeBytes: 9, InputObjectKey: "input/job", OutputObjectKey: &output, Status: "COMPLETED"}
	if err := db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	ctx, recorder := requestContext(http.MethodGet, "/videos?limit=1000", nil)
	ctx.Request.URL.RawQuery = "limit=1000"
	api.List(ctx)
	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte("done.mp4")) {
		t.Fatalf("list=%s", recorder.Body.String())
	}
	ctx, recorder = requestContext(http.MethodGet, "/videos/"+job.ID.String(), nil)
	ctx.Params = gin.Params{{Key: "id", Value: job.ID.String()}}
	api.Get(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("get=%s", recorder.Body.String())
	}
	ctx, recorder = requestContext(http.MethodGet, "/videos/"+job.ID.String()+"/download", nil)
	ctx.Params = gin.Params{{Key: "id", Value: job.ID.String()}}
	api.Download(ctx)
	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte("downloadUrl")) {
		t.Fatalf("download=%s", recorder.Body.String())
	}
}

func TestGetAndDownloadRejectUnknownOrIncompleteJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newTestAPI(t)
	pending := Job{ID: uuid.New(), UserID: "user-1", OriginalFilename: "pending.mp4", ContentType: "video/mp4", SizeBytes: 1, InputObjectKey: "input/pending", Status: "PENDING"}
	if err := db.Create(&pending).Error; err != nil {
		t.Fatal(err)
	}
	ctx, recorder := requestContext(http.MethodGet, "/videos/bad", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "bad"}}
	api.Get(ctx)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unknown=%d", recorder.Code)
	}
	ctx, recorder = requestContext(http.MethodGet, "/videos/"+pending.ID.String()+"/download", nil)
	ctx.Params = gin.Params{{Key: "id", Value: pending.ID.String()}}
	api.Download(ctx)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("pending=%d", recorder.Code)
	}
}

func TestConfirmRejectsMissingObjectAndInvalidState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newTestAPI(t)
	awaiting := Job{ID: uuid.New(), UserID: "user-1", OriginalFilename: "movie.mp4", ContentType: "video/mp4", SizeBytes: 1, InputObjectKey: "missing", Status: "AWAITING_UPLOAD"}
	failed := Job{ID: uuid.New(), UserID: "user-1", OriginalFilename: "bad.mp4", ContentType: "video/mp4", SizeBytes: 1, InputObjectKey: "bad", Status: "FAILED"}
	if err := db.Create(&awaiting).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&failed).Error; err != nil {
		t.Fatal(err)
	}
	ctx, recorder := requestContext(http.MethodPost, "/", nil)
	ctx.Params = gin.Params{{Key: "id", Value: awaiting.ID.String()}}
	api.Confirm(ctx)
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("awaiting=%d", recorder.Code)
	}
	ctx, recorder = requestContext(http.MethodPost, "/", nil)
	ctx.Params = gin.Params{{Key: "id", Value: failed.ID.String()}}
	api.Confirm(ctx)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("failed=%d", recorder.Code)
	}
}

func TestConfirmQueuesValidUploadAndIsIdempotent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newTestAPI(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodHead {
			t.Errorf("unexpected method %s", request.Method)
		}
		w.Header().Set("Content-Length", "10")
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client, err := minio.New(strings.TrimPrefix(server.URL, "http://"), &minio.Options{Creds: credentials.NewStaticV4("key", "secret", ""), Secure: false, Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	api.internal = client
	info, err := client.StatObject(context.Background(), "input", "uploads/movie", minio.StatObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != 10 || info.ContentType != "video/mp4" {
		t.Fatalf("stat=%+v", info)
	}
	job := Job{ID: uuid.New(), UserID: "user-1", OriginalFilename: "movie.mp4", ContentType: "video/mp4", SizeBytes: 10, InputObjectKey: "uploads/movie", Status: "AWAITING_UPLOAD"}
	if err := db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	ctx, recorder := requestContext(http.MethodPost, "/confirm", nil)
	ctx.Params = gin.Params{{Key: "id", Value: job.ID.String()}}
	api.Confirm(ctx)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "PENDING") {
		t.Fatalf("confirm=%d %s", recorder.Code, recorder.Body.String())
	}
	var events int64
	if err := db.Model(&OutboxEvent{}).Count(&events).Error; err != nil || events != 1 {
		t.Fatalf("events=%d err=%v", events, err)
	}
	ctx, recorder = requestContext(http.MethodPost, "/confirm", nil)
	ctx.Params = gin.Params{{Key: "id", Value: job.ID.String()}}
	api.Confirm(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("idempotent=%d", recorder.Code)
	}
}

func TestListClampsLimitAndFiltersStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newTestAPI(t)
	for _, status := range []string{"PENDING", "COMPLETED"} {
		job := Job{ID: uuid.New(), UserID: "user-1", OriginalFilename: status + ".mp4", ContentType: "video/mp4", SizeBytes: 1, InputObjectKey: status, Status: status}
		if err := db.Create(&job).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctx, recorder := requestContext(http.MethodGet, "/videos?status=COMPLETED&limit=-1", nil)
	ctx.Request.URL.RawQuery = "status=COMPLETED&limit=-1"
	api.List(ctx)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "COMPLETED.mp4") || strings.Contains(recorder.Body.String(), "PENDING.mp4") || !strings.Contains(recorder.Body.String(), `"limit":1`) {
		t.Fatalf("list=%s", recorder.Body.String())
	}
}
