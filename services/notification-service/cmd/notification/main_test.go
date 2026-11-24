package main

import "testing"

func TestEnvOrDefault(t *testing.T) { t.Setenv("NOTIFICATION_TEST", ""); if got := envOrDefault("NOTIFICATION_TEST", "fallback"); got != "fallback" { t.Fatal(got) } }
