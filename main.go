package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run . <url>")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	info, err := fetchInfo(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("%s [%s]\n", info.Title, info.Extractor)

	ytdlpPath := os.Getenv("YTDLP_PATH") // This dictates which ytdlp is used, default is env variable, fallback is hardcoded .venv path
	if ytdlpPath == "" {
		ytdlpPath = defaultYtdlpPath
	}
	d := &Downloader{YtdlpPath: ytdlpPath, OutDir: "downloads"}

	path, err := d.Download(ctx, info)
	if err != nil {
		fmt.Fprintln(os.Stderr, "\nerror:", err)
		os.Exit(1)
	}
	fmt.Println("\nzapisano:", path)
}
