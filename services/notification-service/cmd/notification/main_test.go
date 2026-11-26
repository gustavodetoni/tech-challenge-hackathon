package main

import (
	"strings"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("NOTIFICATION_TEST", "")
	if got := envOrDefault("NOTIFICATION_TEST", "fallback"); got != "fallback" {
		t.Fatal(got)
	}
}

func TestRenderFailureEmailUsesPortugueseMultipartContent(t *testing.T) {
	email := renderFailureEmail("no-reply@fiap-video.local", "admin@fiap-video.local", failure{
		JobID:     "job-123",
		ErrorCode: "PROCESSING_FAILED",
	})
	for _, expected := range []string{
		"multipart/alternative",
		"Não foi possível processar o seu vídeo",
		"PROCESSAMENTO NÃO CONCLUÍDO",
		"job-123",
	} {
		if !strings.Contains(email, expected) {
			t.Fatalf("email should contain %q", expected)
		}
	}
}
