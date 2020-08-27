package healthchecker

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
	Result    Outcome
	Duration  time.Duration
}

type HealthCheck struct {
	check    func() *CheckResult
	sinks    []Emitter
	Interval time.Duration
	Name     string
	Type     string
}

func (h *HealthCheck) Run() {
	res := h.check()
	for _, s := range h.sinks {
		log.Debugf("Emitting %s to: %v", h.Name, s)
		s.Emit(h.Name, h.Type, res)
	}
}

func (h *HealthCheck) String() string {
	return fmt.Sprintf("(%s: %s)", h.Type, h.Name)
}

type HealthCheckConstructor func(map[string]string) (func() *CheckResult, error)
type SinkConstructor func(map[string]string) (Emitter, error)

type CheckRegistry struct {
	CheckConstructors map[string]HealthCheckConstructor
	SinkConstructors  map[string]SinkConstructor
	Checks            []*HealthCheck
	Sinks             map[string]Emitter
	running           bool
}

func NewCheckRegistry() *CheckRegistry {
	log.Debug("Creating new CheckRegistry")
	registry := CheckRegistry{}
	registry.CheckConstructors = make(map[string]HealthCheckConstructor)
	registry.SinkConstructors = make(map[string]SinkConstructor)
	registry.Checks = make([]*HealthCheck, 0)
	registry.Sinks = make(map[string]Emitter)
	registry.running = false
	return &registry
}

func (c *CheckRegistry) NewCheck(checkName string, checkType string, checkArgs map[string]string, interval int, sinks []Emitter) (*HealthCheck, error) {
	log.Infof("Creating new health check: %s (%s) (%v)", checkType, checkName, checkArgs)
	checkConstructor, ok := c.CheckConstructors[checkType]
	if !ok {
		return nil, fmt.Errorf("Unable to create check: '%s:%s' because it's not registered", checkType, checkName)
	}
	newCheck, err := checkConstructor(checkArgs)
	if err != nil {
		return nil, fmt.Errorf("Unable to create check '%s: %v' because: %v", checkType, checkArgs, err)
	}

	hc := HealthCheck{
		check:    newCheck,
		sinks:    sinks,
		Interval: time.Duration(interval) * time.Second,
		Name:     checkName,
		Type:     checkType,
	}
	c.Checks = append(c.Checks, &hc)
	return &hc, nil
}

func (c *CheckRegistry) RegisterHealthChecks(conf *Config) {
	for _, hc := range conf.HealthChecks {
		log.Debugf("Creating sinks for %s", hc.Name)
		sinks, err := c.setupSinks(hc.Sinks)
		if err != nil {
			log.Errorf("Could not register %s due to: %s", hc.Name, err)
			continue
		}
		_, err = c.NewCheck(hc.Name, hc.Type, hc.Args, hc.Interval, sinks)
		if err != nil {
			log.Errorf("Could not register %s due to: %s", hc.Name, err)
		}
	}
}

func (c *CheckRegistry) getOrCreateSink(sinkName string, sinkArgs map[string]string) (Emitter, error) {
	sinkId, sinkIdExists := sinkArgs["id"]
	if sinkIdExists {
		delete(sinkArgs, "id")
		if sink, ok := c.Sinks[sinkId]; ok {
			return sink, nil
		}
	}

	sinkConstructor, ok := c.SinkConstructors[sinkName]
	if !ok {
		return nil, fmt.Errorf("Unable to create sink '%s': unknown sink type", sinkName)
	}

	newSink, err := sinkConstructor(sinkArgs)
	if err != nil {
		return nil, fmt.Errorf("Unable to create sink '%s' with args: %v", sinkName, sinkArgs)
	}
	if sinkIdExists {
		c.Sinks[sinkId] = newSink
	}
	return newSink, nil
}

func (c *CheckRegistry) setupSinks(sinkConfigs []map[string]map[string]string) ([]Emitter, error) {
	sinks := make([]Emitter, 0)
	for _, sinkConfig := range sinkConfigs {
		for sinkName, sinkArgs := range sinkConfig {
			sink, err := c.getOrCreateSink(sinkName, sinkArgs)
			if err != nil {
				return nil, err
			}
			sinks = append(sinks, sink)
		}
	}
	return sinks, nil
}

func (c *CheckRegistry) StartRunning() {
	//TODO: Refactor: Running the checks shouldn't belong in the registry.
	log.Infof("Will start %d health checks", len(c.Checks))
	log.Debugf("Health checks: %s", c.Checks)
	c.running = true
	var wg sync.WaitGroup
	for i, _ := range c.Checks {
		wg.Add(1)
		// TODO something to catch errors and restart goroutines with checks
		go func(chk *HealthCheck) {
			defer wg.Done()
			log.Infof("Running check: %s", chk.Name)
			chk.Run()
			for range time.Tick(chk.Interval) {
				if !c.running {
					log.Infof("Stopping check: %s", chk.Name)
					return
				}
				log.Infof("Running check: %s", chk.Name)
				chk.Run()
			}
		}(c.Checks[i])
	}
	wg.Wait()
}

func (c *CheckRegistry) StopRunning() {
	log.Info("Stopping health checks")
	c.running = false
}

func (c *CheckResult) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}
