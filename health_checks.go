package healthchecker

import (
	"time"
	"fmt"
)

type CheckRegistry struct {
	CheckFuncs map[string]func(map[string]string) (func() *CheckResult)
	Checks []*HealthCheck
}

func NewCheckRegistry() *CheckRegistry {
	registry := CheckRegistry{}
	// TODO: set timeout from config
	httpChecker := NewHTTPChecker(10 * time.Second)
	registry.CheckFuncs["SimpleHTTPCheck"] = httpChecker.NewSimpleHTTPCheck
	registry.CheckFuncs["RegexpHTTPCheck"] = httpChecker.NewRegexpHTTPCheck
	return &registry
}

func (c *CheckRegistry) NewCheck(name string, args map[string]string) *HealthCheck {
	hc := HealthCheck{
		check: c.CheckFuncs[name](args),
		sink: nil,
	}
	c.Checks = append(c.Checks, &hc)
	return &hc
}


type CheckResult struct {
	Timestamp time.Time
	Name      string
	Result    bool
	Duration  time.Duration
}

func (c *CheckResult) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}


type HealthCheck struct {
	check func() *CheckResult
	sink []*Sink
}
