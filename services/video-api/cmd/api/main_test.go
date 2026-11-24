package main

import "testing"

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("TEST_HTTP_ADDR", "")
	if got := envOrDefault("TEST_HTTP_ADDR", ":9999"); got != ":9999" { t.Fatalf("got %q", got) }
}
