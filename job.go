package jobs

type Job interface {
	Exec()
}

type JobFunc func()