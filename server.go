package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type Server struct {
	ctx   context.Context
	store *JobStore
	dl    *Downloader
}

func (server *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", server.handleCreateJob)
	mux.HandleFunc("GET /jobs", server.handleListJobs)
	mux.HandleFunc("GET /jobs/{id}", server.handleGetJob)
	mux.Handle("GET /files/", http.StripPrefix("/files/", http.FileServer(http.Dir(server.dl.OutDir))))
	return mux
}

func (server *Server) handleGetJob(writer http.ResponseWriter, request *http.Request) {
	job, ok := server.store.Get(request.PathValue("id"))
	if !ok {
		http.Error(writer, "job not found", http.StatusNotFound)
		return
	}
	writeJSON(writer, http.StatusOK, job)
}

func (server *Server) handleListJobs(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, server.store.List())
}

func (server *Server) handleCreateJob(writer http.ResponseWriter, request *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "invalid JSON", http.StatusBadRequest)
		return
	}
	u, err := url.Parse(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		http.Error(writer, "url must be an http(s) link", http.StatusBadRequest)
	}

	job := &Job{
		ID:        rand.Text(),
		URL:       u.String(),
		Status:    StatusFetching,
		CreatedAt: time.Now(),
	}
	server.store.Add(job)
	snapshot := *job

	go server.runJob(job.ID, job.URL)

	writeJSON(writer, http.StatusAccepted, snapshot)
}

func (server *Server) runJob(id, pageURL string) {
	info, err := fetchInfo(server.ctx, pageURL)
	if err != nil {
		server.fail(id, err)
		return
	}

	server.store.Update(id, func(j *Job) {
		j.Title = info.Title
		j.Status = StatusDownloading
	})
	path, err := server.dl.Download(server.ctx, info, func(p string) {
		server.store.Update(id, func(j *Job) { j.Progress = p })
	})
	if err != nil {
		server.fail(id, err)
		return
	}
	server.store.Update(id, func(j *Job) {
		j.Status = StatusDone
		j.Progress = "100%"
		j.File = "/files/" + filepath.Base(path)
	})
}

func (server *Server) fail(id string, err error) {
	log.Printf("job %s failed: %v", id, err)
	server.store.Update(id, func(j *Job) {
		j.Status = StatusFailed
		j.Error = err.Error()
	})
}

func writeJSON(writer http.ResponseWriter, status int, v any) {
	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(status)
	json.NewEncoder(writer).Encode(v)
}
