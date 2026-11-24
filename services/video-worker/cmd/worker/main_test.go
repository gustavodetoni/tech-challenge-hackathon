package main

import "testing"

func TestEnvOrDefault(t *testing.T) { t.Setenv("WORKER_TEST", ""); if got := envOrDefault("WORKER_TEST", "fallback"); got != "fallback" { t.Fatal(got) } }
