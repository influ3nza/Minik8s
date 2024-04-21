package scheduler

type SchedulingPolicy string

const (
	LeastUsed SchedulingPolicy = "LeastUsed"
	Random    SchedulingPolicy = "Random"
	Poll      SchedulingPolicy = "Poll"
)

type SchedulerConfig struct {
	policy     SchedulingPolicy
	apiAddress string
	apiPort    string
}

func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		policy:     Poll,
		apiAddress: "127.0.0.1",
		apiPort:    "50000",
	}
}
