package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Downloader struct {
	YtdlpPath  string
	FfmpegPath string // pusty = szukaj w PATH
	OutDir     string
}

func (d *Downloader) Download(ctx context.Context, info *VideoInfo) (string, error) {
	if err := os.MkdirAll(d.OutDir, 0o755); err != nil {
		return "", fmt.Errorf("create out dir: %w", err)
	}

	selector := "b"
	video, okV := bestVideo(info.Formats)
	audio, okA := bestAudio(info.Formats)
	if okV && okA {
		selector = video.FormatID + "+" + audio.FormatID
	}

	base := fmt.Sprintf("%s_%s", info.Extractor, info.ID)
	outPath := filepath.Join(d.OutDir, base+".mp4")

	args := []string{
		"-f", selector,
		"--merge-output-format", "mp4",
		"-o", filepath.Join(d.OutDir, base+".%(ext)s"),
		"--newline",
		"--progress-template", "download:PROGRESS %(progress._percent_str)s %(progress._speed_str)s",
	}
	if d.FfmpegPath != "" {
		args = append(args, "--ffmpeg-location", d.FfmpegPath)
	}
	pageUrl := info.WebpageUrl
	if pageUrl == "" {
		pageUrl = info.OriginalUrl
	}

	args = append(args, "--", pageUrl)

	cmd := exec.CommandContext(ctx, d.YtdlpPath, args...)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if p, ok := strings.CutPrefix(line, "PROGRESS "); ok {
			fmt.Printf("\r  %s   ", strings.TrimSpace(p))
		} else {
			fmt.Println("\n" + line)
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("yt-dlp: %w", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("expected output missing: %w", err)
	}
	return outPath, nil
}
