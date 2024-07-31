package scheduler

type ScheduleTask struct {
	instanceID string
	version    string
	action     string
	retry      int
}

type Scheduler struct {
	queue chan ScheduleTask
}
