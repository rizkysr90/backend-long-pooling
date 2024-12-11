package repository

import (
	"log"
	"sync"
	"time"
)

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

type Job struct {
	JobID              string
	StatusInPercentage int
	Status             JobStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Error              string
}

type JobManager struct {
	Jobs      sync.Map
	Listeners map[string][]chan *Job
	Mu        sync.RWMutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		Listeners: make(map[string][]chan *Job),
	}
}

func (jm *JobManager) ProcessJob(jobID string) {
	progress := 0
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for progress < 100 {
		getTick := <-ticker.C
		if !getTick.IsZero() {
			progress += 10
			log.Println("processing job id", jobID, ":", " ", progress)
			jm.updateJobProgress(jobID, progress)
		}
	}
}

func (jm *JobManager) updateJobProgress(jobID string, progress int) {
	if jobValue, exists := jm.Jobs.Load(jobID); exists {
		job := jobValue.(*Job)
		job.StatusInPercentage = progress
		job.UpdatedAt = time.Now()

		if progress >= 100 {
			job.Status = StatusCompleted
		} else {
			job.Status = StatusRunning
		}

		jm.Jobs.Store(jobID, job)
		jm.notifyListeners(jobID, job)
	}
}
func (jm *JobManager) notifyListeners(jobID string, job *Job) {
	jm.Mu.RLock() // Read lock since we're only reading from listeners map
	defer jm.Mu.RUnlock()

	if listeners, exists := jm.Listeners[jobID]; exists {
		for _, listener := range listeners {
			select {
			case listener <- job:
				// Successfully sent update
			default:
				// Channel is blocked, skip this listener
			}
		}
	}
}

func (jm *JobManager) Subscribe(jobID string) chan *Job {
	jm.Mu.Lock() // Write lock since we're modifying the listeners map
	defer jm.Mu.Unlock()

	ch := make(chan *Job, 1)
	jm.Listeners[jobID] = append(jm.Listeners[jobID], ch)
	return ch
}
func (jm *JobManager) Unsubscribe(jobID string, ch chan *Job) {
	jm.Mu.Lock() // Write lock since we're modifying the listeners map
	defer jm.Mu.Unlock()

	if listeners, exists := jm.Listeners[jobID]; exists {
		for i, listener := range listeners {
			if listener == ch {
				// Remove this channel from Listeners
				jm.Listeners[jobID] = append(jm.Listeners[jobID][:i], jm.Listeners[jobID][i+1:]...)
				close(ch)
				break
			}
		}
		// If no more Listeners, clean up the map entry
		if len(jm.Listeners[jobID]) == 0 {
			delete(jm.Listeners, jobID)
		}
	}
}
