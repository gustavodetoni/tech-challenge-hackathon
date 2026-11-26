package notification

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderFailureEmailIsPortugueseAndEscapesValues(t *testing.T) {
	email := RenderFailureEmail("no-reply@example.test", "admin@example.test", Failure{JobID: "<job-1>", ErrorCode: "OTHER"})
	for _, expected := range []string{"multipart/alternative", "Não foi possível processar", "O processamento do vídeo foi interrompido.", "&lt;job-1&gt;"} {
		if !strings.Contains(email, expected) {
			t.Fatalf("email should contain %q", expected)
		}
	}
}

func TestSendWritesSMTPConversation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	received := make(chan string, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		data, _ := bufio.NewReader(connection).ReadString('\x00')
		received <- data
	}()
	if err := Send(listener.Addr().String(), "from@example.test", "to@example.test", []byte(`{"jobId":"job-1","errorCode":"PROCESSING_FAILED"}`)); err != nil {
		t.Fatal(err)
	}
	data := <-received
	for _, expected := range []string{"HELO fiap-video", "MAIL FROM:<from@example.test>", "RCPT TO:<to@example.test>", "job-1", "QUIT"} {
		if !strings.Contains(data, expected) {
			t.Fatalf("smtp data should contain %q: %q", expected, data)
		}
	}
}

func TestSendRejectsInvalidEvent(t *testing.T) {
	if err := Send("127.0.0.1:1", "from", "to", []byte("not-json")); err == nil {
		t.Fatal("expected error")
	}
}

func TestHealthHandler(t *testing.T) {
	handler := HealthHandler()
	for _, path := range []string{"/health/live", "/health/ready", "/metrics"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: %d", path, recorder.Code)
		}
	}
}
