package scheduler

import "github.com/labstack/gommon/log"

// SchedulerReporter report metrics data to metrics service
type SchedulerReporter struct {
	schedulers []Scheduler
}

func NewSchedulerReporter(schedulers []Scheduler) *SchedulerReporter {
	return &SchedulerReporter{
		schedulers: schedulers,
	}
}

type SchedulerRunTimeReport struct {
	JobName string
	Times   int64
}

func (r *SchedulerReporter) Report() []SchedulerRunTimeReport {
	reports := make([]SchedulerRunTimeReport, 0, len(r.schedulers))
	for _, scheduler := range r.schedulers {
		name := scheduler.Name()
		runTimes := scheduler.RunTimes()
		log.Infof("[Report] name:%s, runTimes:%d", name, runTimes)
		reports = append(reports, SchedulerRunTimeReport{JobName: name, Times: runTimes})
	}
	return reports
}
