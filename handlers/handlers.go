package handlers

import (
	"fmt"
	"log"
	"long-pool-demo/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ResponseInsertJob struct {
	JobID string `json:"job_id"`
}

func InsertJob(ctx *gin.Context) {
	generateUniqueID := uuid.NewString()
	responseData := &ResponseInsertJob{
		JobID: generateUniqueID,
	}
	newJob := repository.Job{JobID: generateUniqueID, StatusInPercentage: 0}
	repository.Jobs = append(repository.Jobs, newJob)

	log.Println("new job : ", generateUniqueID)
	log.Println("list of job :", repository.Jobs)

	ctx.JSON(http.StatusOK, responseData) // Return the struct directly
}

func CheckJobStatus(ctx *gin.Context) {
	jobID := ctx.Param("job_id")

	// Create channels for progress and completion
	progressChan := make(chan int)
	doneChan := make(chan bool)

	// Simulate long running task
	go func() {
		progress := 0
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for progress < 100 {
			t := <-ticker.C
			if !t.IsZero() {
				progress += 10 // Increment progress by 10%
				progressChan <- progress
			}
		}

		// Task completed
		doneChan <- true
	}()

	for {
		select {
		case <-doneChan:
			// Only return when task is complete
			ctx.JSON(http.StatusOK, gin.H{
				"job_id":   jobID,
				"status":   "completed",
				"progress": 100,
			})
			return

		case progress := <-progressChan:
			// Update global Jobs status
			for i := range repository.Jobs {
				if repository.Jobs[i].JobID == jobID {
					repository.Jobs[i].StatusInPercentage = progress
					break
				}
			}

			// Don't return here - continue processing
			fmt.Printf("Job %s progress: %d%%\n", jobID, progress)

		case <-time.After(15 * time.Second):
			ctx.JSON(http.StatusRequestTimeout, gin.H{
				"job_id": jobID,
				"status": "timeout",
			})
			return
		}
	}
}
