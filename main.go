package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ytdlpPath := os.Getenv("YTDLP_PATH")
	if ytdlpPath == "" {
		ytdlpPath = defaultYtdlpPath
	}

	srv := &Server{
		ctx:   ctx,
		store: NewJobStore(),
		dl:    &Downloader{YtdlpPath: ytdlpPath, OutDir: "downloads"},
	}
	httpSrv := &http.Server{Addr: "127.0.0.1:8080", Handler: srv.routes()}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpSrv.Shutdown(shutdownCtx)
	}()

	log.Println("listening on %w", httpSrv.Addr)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
