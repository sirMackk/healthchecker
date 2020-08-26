package healthchecker

import (
	"fmt"
	"time"
)

type CheckResult struct {
	Timestamp time.Time
	Name      string
	Result    bool
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

type HealthCheckConstructor func(map[string]string) func() *CheckResult
type SinkConstructor func(map[string]string) Sink

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

func (c *CheckRegistry) NewCheck(checkType string, checkArgs map[string]string, sinks []Sink) *HealthCheck {
	hc := HealthCheck{
		check: c.CheckConstructors[checkType](checkArgs),
		sinks: sinks,
	}
	c.Checks = append(c.Checks, &hc)
	return &hc
}

func (c *CheckResult) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}
