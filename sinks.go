package healthchecker

//TODO rsyslog sink
//TODO influxdb sink

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Sink interface {
	Emit(c *CheckResult)
}

type ConsoleSink struct {
	TargetStream *os.File
}

func NewConsoleSink(args map[string]string) (Sink, error) {
	useStdout, ok := args["useStdout"]
	if !ok {
		return nil, fmt.Errorf("Error creating ConsoleSink - useStdout option missing")
	}
	choice, err := strconv.ParseBool(useStdout)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ConsoleSink 'useStdout' option")
	}

	if !choice {
		return &ConsoleSink{TargetStream: os.Stderr}, nil
	}
	return &ConsoleSink{TargetStream: os.Stdout}, nil
}

func (s *ConsoleSink) Emit(c *CheckResult) {
	fmt.Fprintf(s.TargetStream, "%s [%s]: %s %s\n", c.TimestampString(), c.Name, c.Result, c.Duration.Round(time.Millisecond))
}
