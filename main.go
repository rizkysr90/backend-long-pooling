package main

import (
	"long-pool-demo/handlers"
	"long-pool-demo/repository"

	"github.com/gin-gonic/gin"
)

var Jobs = []repository.Job{} // Initialize empty slice

func main() {
	r := gin.Default()
	jobManager := repository.NewJobManager()
	jobHandlers := handlers.NewJob(jobManager)
	/* Insert is endpoint for sent a job to server and immediately response with job id */
	r.POST("/insert", func(ctx *gin.Context) {
		jobHandlers.CreateJob(ctx)
	})
	// r.GET("/checkjobstatus/:job_id", func(ctx *gin.Context) {
	// 	handlers.CheckJobStatus(ctx)
	// })

	r.Run() // listen and serve on 0.0.0.0:8080
}
