package main

import (
	"sync"
	"time"
)

type JobStatus string

const (
	StatusFetching    JobStatus = "fetching"
	StatusDownloading JobStatus = "downloading"
	StatusDone        JobStatus = "done"
	StatusFailed      JobStatus = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    JobStatus `json:"status"`
	Title     string    `json:"title,omitempty"`
	Progress  string    `json:"progress,omitempty"`
	File      string    `json:"file,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TODO: Add a DELETE method for jobs - each job/request needs a separate context
type JobStore struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*Job)}
}

func (store *JobStore) Add(job *Job) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[job.ID] = job
}

func (store *JobStore) Get(id string) (Job, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	job, ok := store.jobs[id]
	if !ok {
		return Job{}, false
	}
	return *job, true
}

func (store *JobStore) List() []Job {
	store.mu.Lock()
	defer store.mu.Unlock()
	out := make([]Job, 0, len(store.jobs))
	for _, job := range store.jobs {
		out = append(out, *job)
	}
	return out
}

func (store *JobStore) Update(id string, fn func(job *Job)) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if job, ok := store.jobs[id]; ok {
		fn(job)
	}
}
