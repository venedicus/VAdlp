package service

import (
	"context"
	"strings"
	"time"

	"vadlp/internal/applog"
	"vadlp/internal/browse"
	"vadlp/internal/core"
	"vadlp/internal/downloader"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

type DownloadResult struct {
	LogOutput   string
	DurationSec int
}

func (s *Service) Download(
	ctx context.Context,
	cfg core.Config,
	jobID string,
	onEvent func(downloader.Event),
) (DownloadResult, error) {
	if err := cfg.ValidateForDownload(); err != nil {
		return DownloadResult{}, err
	}
	start := time.Now()
	applog.Info("download start", "job", jobID)
	logs, err := downloader.RunCtx(ctx, cfg, jobID, onEvent)
	res := DownloadResult{
		LogOutput:   logs,
		DurationSec: int(time.Since(start).Seconds()),
	}
	if err != nil {
		applog.Info("download end", "job", jobID, "err", err.Error())
		return res, err
	}
	applog.Info("download end", "job", jobID, "ok", true)
	return res, nil
}

func (s *Service) Probe(ctx context.Context, cfg core.Config) (downloader.ProbeResult, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return downloader.ProbeResult{}, core.ValidationError{Key: "err.queue_no_url"}
	}
	if err := cfg.Validate(); err != nil {
		return downloader.ProbeResult{}, err
	}
	return downloader.ProbeCtx(ctx, cfg)
}

// BrowseSearch lists videos matching a free-text query, or opens the listing
// directly when the query is a URL.
func (s *Service) BrowseSearch(ctx context.Context, cfg core.Config, query string, limit int) (browse.Result, error) {
	return browse.Search(ctx, cfg, query, limit)
}

// BrowseOpen lists the contents of a channel or playlist URL.
func (s *Service) BrowseOpen(ctx context.Context, cfg core.Config, target string) (browse.Result, error) {
	return browse.Open(ctx, cfg, target)
}
