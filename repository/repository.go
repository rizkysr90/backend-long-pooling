package repository

var Jobs = []Job{} // Capitalized, so visible to other packages

type Job struct {
	JobID              string
	StatusInPercentage int
}
