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

type CheckRegistry struct {
	CheckFuncs map[string]HealthCheckConstructor
	Checks     []*HealthCheck
	// TODO: generalize to Sink interface
	Sinks map[string]func(string) Sink
}

func NewCheckRegistry() *CheckRegistry {
	registry := CheckRegistry{}
	registry.CheckFuncs = make(map[string]HealthCheckConstructor)
	registry.Checks = make([]*HealthCheck, 0)
	registry.Sinks = make(map[string]func(string) Sink)
	return &registry
}

func (c *CheckRegistry) NewCheck(name string, args map[string]string, sinks []Sink) *HealthCheck {
	hc := HealthCheck{
		check: c.CheckFuncs[name](args),
		sinks: sinks,
	}
	c.Checks = append(c.Checks, &hc)
	return &hc
}

func (c *CheckResult) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}
