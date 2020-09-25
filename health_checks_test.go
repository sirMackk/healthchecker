package healthchecker

import (
	"errors"
	"testing"
	//"time"
)

func testingCheckConstructor(_ map[string]string) (func() chan *Result, error) {
	return func() chan *Result {
		resultStream := make(chan *Result, 1)
		defer close(resultStream)
		resultStream <- &Result{}
		return resultStream
	}, nil
}

func TestNewCheckCorrect(t *testing.T) {
	registry := NewRegistry()
	registry.CheckConstructors["testing"] = testingCheckConstructor

	registry.AddCheck("some check", "testing", nil, 0, nil)
}

func TestNewCheckFail(t *testing.T) {
	registry := NewRegistry()
	registry.CheckConstructors["testing"] = func(_ map[string]string) (func() chan *Result, error) {
		return nil, errors.New("Wat")
	}

	registry.AddCheck("some check", "testing", nil, 0, nil)
}

func TestStartRunning(t *testing.T) {
	registry := NewRegistry()
	var ran = false
	registry.CheckConstructors["testing"] = func(_ map[string]string) (func() chan *Result, error) {
		ran = true
		return func() chan *Result {
			resStream := make(chan *Result)
			defer close(resStream)
			return resStream
		}, nil
	}

	sinks := make([]Emitter, 0)
	registry.AddCheck("some check", "testing", nil, 1, sinks)
	registry.AddCheck("some check", "testing", nil, 1, sinks)

	go registry.StartRunning()
	//time.Sleep(1100 * time.Millisecond)
	registry.StopRunning()
	if !ran {
		t.Errorf("Failed to execute check in loop")
	}
}

func TestRegisterHealthChecks(t *testing.T) {
	config := &Config{
		Core: map[string]string{},
		HealthChecks: []HealthChecksConfig{
			{
				Name: "SomeCheck",
				Type: "testing",
				Args: map[string]string{"url": "http://example.com"},
				Sinks: []map[string]map[string]string{
					{
						"FileSink": map[string]string{
							"id":   "sink1",
							"path": "/tmp/testfile"},
					},
				},
				Interval: 5,
			},
			{
				Name: "SomeOtherCheck",
				Type: "testing",
				Args: map[string]string{"url": "http://example.com"},
				Sinks: []map[string]map[string]string{
					{
						"FileSink": map[string]string{
							"id":   "sink2",
							"path": "/tmp/testfile2"},
					},
				},
				Interval: 5,
			},
		},
	}

	registry := NewRegistry()
	var ran = false
	registry.CheckConstructors["testing"] = func(_ map[string]string) (func() chan *Result, error) {
		ran = true
		return func() chan *Result {
				resStream := make(chan *Result)
				defer close(resStream)
				return resStream
			},
			nil
	}
	registry.SinkConstructors["FileSink"] = NewFileSink
	registry.RegisterHealthChecks(config)

	if !ran {
		t.Errorf("Failed to register check")
	}

	if len(registry.Checks) != 2 || len(registry.Sinks) != 2 {
		t.Errorf("Wrong number of checks or sinks: c:%d, s:%d", len(registry.Checks), len(registry.Sinks))
	}

	if cName := registry.Checks[0].Name; cName != "SomeCheck" {
		t.Errorf("First check should be 'SomeCheck', got: %s", cName)
	}

	if cName := registry.Checks[1].Name; cName != "SomeOtherCheck" {
		t.Errorf("Second check should be 'SomeOtherCheck', got: %s", cName)
	}
}

func TestRegisterHealthChecksWithSameSink(t *testing.T) {
	config := &Config{
		Core: map[string]string{},
		HealthChecks: []HealthChecksConfig{
			{
				Name: "SomeCheck",
				Type: "testing",
				Args: map[string]string{"url": "http://example.com"},
				Sinks: []map[string]map[string]string{
					{
						"FileSink": map[string]string{
							"id":   "sink1",
							"path": "/tmp/testfile2"},
					},
				},
				Interval: 5,
			},
			{
				Name: "SomeOtherCheck",
				Type: "testing",
				Args: map[string]string{"url": "http://example.com"},
				Sinks: []map[string]map[string]string{
					{
						"FileSink": map[string]string{
							"id":   "sink1",
							"path": "/tmp/testfile1"},
					},
				},
				Interval: 5,
			},
		},
	}

	registry := NewRegistry()
	registry.CheckConstructors["testing"] = func(_ map[string]string) (func() chan *Result, error) {
		return func() chan *Result {
			resStream := make(chan *Result)
			defer close(resStream)
			return resStream
		}, nil
	}
	registry.SinkConstructors["FileSink"] = NewFileSink
	registry.RegisterHealthChecks(config)

	if nChecks := len(registry.Checks); nChecks != 2 {
		t.Errorf("Expected 2 checks, got %d", nChecks)
	}

	if nSinks := len(registry.Sinks); nSinks != 1 {
		t.Errorf("Expected 1 sink, got %d", nSinks)
	}
}
