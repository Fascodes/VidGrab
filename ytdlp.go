package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type Format struct {
	FormatID string  `json:"format_id"`
	Ext      string  `json:"ext"`
	Protocol string  `json:"protocol"`
	VCodec   string  `json:"vcodec"`
	ACodec   string  `json:"acodec"`
	Height   int     `json:"height"`
	TBR      float64 `json:"tbr"` // Total bitrate kbit/s
}

type VideoInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Extractor   string   `json:"extractor"`
	Duration    float64  `json:"duration"`
	WebpageUrl  string   `json:"webpage_url"`
	OriginalUrl string   `json:"original_url"`
	Formats     []Format `json:"formats"`
}

const defaultYtdlpPath string = "./.venv/Scripts/yt-dlp.exe"

func fetchInfo(ctx context.Context, url string) (*VideoInfo, error) {
	out, err := ytdlp(ctx, "-J", "--no-warnings", url).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("yt-dlp failed: %w\n%s", err, exitErr.Stderr)
		}
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}

	var info VideoInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	return &info, nil
}

func ytdlp(ctx context.Context, args ...string) *exec.Cmd {
	bin := defaultYtdlpPath
	if bin == "" {
		bin = "yt-dlp"
	}
	return exec.CommandContext(ctx, bin, args...)
}

func bestVideo(formats []Format) (Format, bool) {
	var best Format
	found := false

	for _, f := range formats {
		if f.VCodec == "none" || f.ACodec != "none" {
			continue
		}
		if !found || f.Height > best.Height ||
			(f.Height == best.Height && f.TBR > best.TBR) {
			best = f
			found = true
		}
	}
	return best, found
}

func bestAudio(formats []Format) (Format, bool) {
	var best Format
	found := false
	for _, f := range formats {
		if f.ACodec == "none" || f.VCodec != "none" {
			continue
		}
		if !found || f.TBR > best.TBR {
			best, found = f, true
		}
	}
	return best, found
}
