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
