package healthchecker

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ResultCode int

const (
	Success ResultCode = 0
	Failure ResultCode = 1
	Error   ResultCode = 2
)

func (o ResultCode) String() string {
	switch o {
	case 0:
		return "Success"
	case 1:
		return "Failure"
	default:
		return "Error"
	}
}

type Result struct {
	Timestamp time.Time
	Result    ResultCode
	Duration  time.Duration
}

func (c *Result) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}

type HealthCheck struct {
	fn       func() chan *Result
	sinks    []Emitter
	Interval time.Duration
	Name     string
	Type     string
}

// TODO get rid of duplicate names eg. CheckResult -> Result
func (h *HealthCheck) Run(ctx context.Context) {

	select {
	case res := <-h.fn():
		for _, s := range h.sinks {
			log.Debugf("Emitting %s result (%s) to %v", h.Name, res.Result, s.Name())
			go func(s Emitter) {
				select {
				case <-s.Emit(h.Name, h.Type, res):
				case <-ctx.Done():
					if err := ctx.Err(); err != nil {
						log.Errorf("Error emitting result from %s to sink %s: %s", h.Name, s.Name(), err)
					}
				}
			}(s)

		}
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			log.Errorf("Error running %s: %s", h.Name, err)
		}
	}
}

func (h *HealthCheck) String() string {
	return fmt.Sprintf("(%s: %s)", h.Type, h.Name)
}

type HealthCheckConstructor func(map[string]string) (func() chan *Result, error)
type SinkConstructor func(map[string]string) (Emitter, error)

type Registry struct {
	CheckConstructors map[string]HealthCheckConstructor
	SinkConstructors  map[string]SinkConstructor
	Checks            []*HealthCheck
	Sinks             map[string]Emitter
	running           bool
}

func NewRegistry() *Registry {
	log.Debug("Creating new CheckRegistry")
	registry := Registry{}
	registry.CheckConstructors = make(map[string]HealthCheckConstructor)
	registry.SinkConstructors = make(map[string]SinkConstructor)
	registry.Checks = make([]*HealthCheck, 0)
	registry.Sinks = make(map[string]Emitter)
	registry.running = false
	return &registry
}

func (c *Registry) AddCheck(checkName string, checkType string, checkArgs map[string]string, interval int, sinks []Emitter) (*HealthCheck, error) {
	log.Infof("Creating new health check: %s (%s) (%v)", checkType, checkName, checkArgs)
	checkConstructor, ok := c.CheckConstructors[checkType]
	if !ok {
		return nil, fmt.Errorf("Unable to create check: '%s:%s' because it's not registered", checkType, checkName)
	}
	checkFn, err := checkConstructor(checkArgs)
	if err != nil {
		return nil, fmt.Errorf("Unable to create check '%s: %v' because: %v", checkType, checkArgs, err)
	}

	hc := HealthCheck{
		fn:       checkFn,
		sinks:    sinks,
		Interval: time.Duration(interval) * time.Second,
		Name:     checkName,
		Type:     checkType,
	}
	c.Checks = append(c.Checks, &hc)
	return &hc, nil
}

func (c *Registry) RegisterHealthChecks(conf *Config) {
	for _, hc := range conf.HealthChecks {
		log.Debugf("Creating sinks for %s", hc.Name)
		sinks, err := c.setupSinks(hc.Name, hc.Sinks)
		if err != nil {
			log.Errorf("Could not register %s due to: %s", hc.Name, err)
			continue
		}
		_, err = c.AddCheck(hc.Name, hc.Type, hc.Args, hc.Interval, sinks)
		if err != nil {
			log.Errorf("Could not register %s due to: %s", hc.Name, err)
		}
	}
}

func (c *Registry) getOrCreateSink(checkName, sinkName string, sinkArgs map[string]string) (Emitter, error) {
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
		return nil, fmt.Errorf("Unable to create sink '%s' with args: %v because: %s", sinkName, sinkArgs, err)
	}
	if sinkIdExists {
		c.Sinks[sinkId] = newSink
	}
	return newSink, nil
}

func (c *Registry) setupSinks(checkName string, sinkConfigs []map[string]map[string]string) ([]Emitter, error) {
	sinks := make([]Emitter, 0)
	for _, sinkConfig := range sinkConfigs {
		for sinkName, sinkArgs := range sinkConfig {
			sink, err := c.getOrCreateSink(checkName, sinkName, sinkArgs)
			if err != nil {
				return nil, err
			}
			sinks = append(sinks, sink)
		}
	}
	return sinks, nil
}

func (c *Registry) StartRunning() {
	log.Infof("Starting: %d health checks", len(c.Checks))
	log.Debugf("Health checks: %s", c.Checks)
	c.running = true
	var wg sync.WaitGroup
	for i, _ := range c.Checks {
		wg.Add(1)
		go func(chk *HealthCheck) {
			defer wg.Done()
			log.Infof("Running check: %s", chk.Name)
			ctx, cancel := context.WithTimeout(context.Background(), chk.Interval)
			defer cancel()
			chk.Run(ctx)
			for range time.Tick(chk.Interval) {
				if !c.running {
					log.Infof("Stopping check: %s", chk.Name)
					return
				}
				log.Infof("Running check: %s", chk.Name)
				ctx, cancel := context.WithTimeout(context.Background(), chk.Interval)
				chk.Run(ctx)
				cancel()
			}
			log.Errorf("Stopping check loop for: %s", chk.Name)
		}(c.Checks[i])
	}
	wg.Wait()
}

func (c *Registry) StopRunning() {
	log.Info("Stopping health checks")
	c.running = false
}
