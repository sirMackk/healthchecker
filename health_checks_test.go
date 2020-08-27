package healthchecker

import (
	"errors"
	"testing"
	"time"
)

func testingCheckConstructor(_ map[string]string) (func() *CheckResult, error) {
	return func() *CheckResult {
		return &CheckResult{}
	}, nil
}

func TestNewCheckCorrect(t *testing.T) {
	registry := NewCheckRegistry()
	registry.CheckConstructors["testing"] = testingCheckConstructor

	registry.NewCheck("some check", "testing", nil, 0, nil)
}

func TestNewCheckFail(t *testing.T) {
	registry := NewCheckRegistry()
	registry.CheckConstructors["testing"] = func(_ map[string]string) (func() *CheckResult, error) {
		return nil, errors.New("Wat")
	}

	registry.NewCheck("some check", "testing", nil, 0, nil)
}

func TestStartRunning(t *testing.T) {
	registry := NewCheckRegistry()
	var ran = false
	registry.CheckConstructors["testing"] = func(_ map[string]string) (func() *CheckResult, error) {
		ran = true
		return func() *CheckResult { return nil }, nil
	}

	sinks := make([]Emitter, 0)
	registry.NewCheck("some check", "testing", nil, 1, sinks)
	registry.NewCheck("some check", "testing", nil, 1, sinks)

	go registry.StartRunning()
	time.Sleep(1100 * time.Millisecond)
	registry.StopRunning()
	if !ran {
		t.Errorf("Failed to execute check in loop")
	}
}
