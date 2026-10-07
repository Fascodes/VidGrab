package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Extractor string   `json:"extractor"`
	Duration  float64  `json:"duration"`
	Formats   []Format `json:"formats"`
}

const ytdlpPath string = "./.venv/Scripts/yt-dlp.exe"

func fetchInfo(url string) (*VideoInfo, error) {
	out, err := ytdlp("-J", "--no-warnings", url).Output()
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

func ytdlp(args ...string) *exec.Cmd {
	bin := ytdlpPath
	if bin == "" {
		bin = "yt-dlp"
	}
	return exec.Command(bin, args...)
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run . <url>")
		os.Exit(1)
	}

	info, err := fetchInfo(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}

	fmt.Printf("%s [%s] %.0fs\n", info.Title, info.Extractor, info.Duration)
	fmt.Printf("FormatID	Ext		Protocol	VCodec	ACodec	Height	TBR\n")
	for _, f := range info.Formats {
		fmt.Printf("%-12s %-5s %-18s v=%-14s a=%-10s %4dp %7.0fk\n",
			f.FormatID, f.Ext, f.Protocol, f.VCodec, f.ACodec, f.Height, f.TBR)
	}
	if video, ok := bestVideo(info.Formats); ok {
		fmt.Printf("best video: %s %dp %s\n", video.FormatID, video.Height, video.VCodec)
	} else {
		fmt.Println("no video codec available")
	}
}
