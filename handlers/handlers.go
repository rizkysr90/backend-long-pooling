package handlers

import (
	"long-pool-demo/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Job struct {
	manager *repository.JobManager
}

func NewJob(jobManager *repository.JobManager) *Job {
	return &Job{
		manager: jobManager,
	}
}
func (j *Job) CreateJob(ctx *gin.Context) {
	jobID := uuid.NewString()
	newJob := &repository.Job{
		JobID:              jobID,
		StatusInPercentage: 0,
		Status:             repository.StatusPending,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	// This is thread-safe because sync.Map handles concurrency internally
	j.manager.Jobs.Store(jobID, newJob)
	// Start processing in background
	go j.manager.ProcessJob(jobID)

	ctx.JSON(http.StatusOK, gin.H{
		"job_id": jobID,
		"status": newJob.Status,
	})
}
func (j *Job) CheckJobStatus(ctx *gin.Context) {
	jobID := ctx.Param("job_id")

	// Check if job exists
	jobValue, exists := j.manager.Jobs.Load(jobID)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	job := jobValue.(*repository.Job)

	// If job is already completed, return immediately
	if job.Status == repository.StatusCompleted {
		ctx.JSON(http.StatusOK, job)
		return
	}

	// Subscribe to updates
	updates := j.manager.Subscribe(jobID)
	defer j.manager.Unsubscribe(jobID, updates)

	select {
	case updatedJob := <-updates:
		ctx.JSON(http.StatusOK, updatedJob)
	case <-ctx.Done():
		// Client disconnected
		return
	case <-time.After(15 * time.Second):
		ctx.JSON(http.StatusRequestTimeout, gin.H{
			"job_id": jobID,
			"status": "timeout",
		})
	}
}
