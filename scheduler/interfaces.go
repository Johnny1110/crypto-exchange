package scheduler

type Scheduler interface {
	Start() error
	Stop() error
	Name() string
	RunTimes() int64
}
