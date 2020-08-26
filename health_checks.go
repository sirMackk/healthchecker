package healthchecker

import (
	"fmt"
	"time"
)

type Outcome int

const (
	Success Outcome = 0
	Failure Outcome = 1
	Error   Outcome = 2
)

func (o Outcome) String() string {
	switch o {
	case 0:
		return "Success"
	case 1:
		return "Failure"
	default:
		return "Error"
	}
}

type CheckResult struct {
	Timestamp time.Time
	Name      string
	Result    Outcome
	Duration  time.Duration
}

type HealthCheck struct {
	check func() *CheckResult
	sinks []Sink
}

func (h *HealthCheck) Run() {
	res := h.check()
	for _, s := range h.sinks {
		s.Emit(res)
	}
}

type HealthCheckConstructor func(map[string]string) (func() *CheckResult, error)
type SinkConstructor func(map[string]string) (Sink, error)

type CheckRegistry struct {
	CheckConstructors map[string]HealthCheckConstructor
	SinkConstructors  map[string]SinkConstructor
	Checks            []*HealthCheck
}

func NewCheckRegistry() *CheckRegistry {
	registry := CheckRegistry{}
	registry.CheckConstructors = make(map[string]HealthCheckConstructor)
	registry.SinkConstructors = make(map[string]SinkConstructor)
	registry.Checks = make([]*HealthCheck, 0)
	return &registry
}

func (c *CheckRegistry) NewCheck(checkType string, checkArgs map[string]string, sinks []Sink) (*HealthCheck, error) {
	newCheck, err := c.CheckConstructors[checkType](checkArgs)
	if err != nil {
		return &HealthCheck{}, fmt.Errorf("Unable to create check '%s: %v' because: %v", checkType, checkArgs, err)
	}
	hc := HealthCheck{
		check: newCheck,
		sinks: sinks,
	}
	c.Checks = append(c.Checks, &hc)
	return &hc, nil
}

func (c *CheckResult) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}
