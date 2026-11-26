package middleware

import (
	"github.com/fiap-x/video-platform/video-api/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRejectsMissingAndInvalidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := auth.New("issuer", "audience", "http://127.0.0.1:1")
	router := gin.New()
	router.GET("/", Auth(validator), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	for _, header := range []string{"", "Basic x", "Bearer invalid"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("header %q: got %d", header, recorder.Code)
		}
	}
}

func TestSubjectReadsAuthenticatedRequestValue(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(SubjectKey, "subject-1")
	if got := Subject(ctx); got != "subject-1" {
		t.Fatal(got)
	}
}
