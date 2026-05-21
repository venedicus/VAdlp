package service

import (
	"context"
	"testing"

	"vadlp/internal/core"
)

func TestDownloadValidateForDownload(t *testing.T) {
	svc := New()
	cfg := core.DefaultConfig()
	_, err := svc.Download(context.Background(), cfg, "job", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if v, ok := core.AsValidation(err); !ok || v.Key != "err.queue_no_url" {
		t.Fatalf("got %v", err)
	}
}

func TestProbeRequiresURL(t *testing.T) {
	svc := New()
	_, err := svc.Probe(core.DefaultConfig())
	if err == nil {
		t.Fatal("expected error")
	}
	if v, ok := core.AsValidation(err); !ok || v.Key != "err.queue_no_url" {
		t.Fatalf("got %v", err)
	}
}

func TestProbeValidatesConfig(t *testing.T) {
	svc := New()
	cfg := core.DefaultConfig()
	cfg.URL = "https://example.com/v"
	cfg.Retries = -1
	_, err := svc.Probe(cfg)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
