package scheduler

type Scheduler interface {
	Start() error
	Stop() error
}
