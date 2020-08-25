package healthchecker

//TODO rsyslog sink
//TODO influxdb sink

import (
	"fmt"
	"os"
	"time"
)

type Sink interface {
	Emit(c *CheckResult)
}

type ConsoleSink struct {
	targetStream *os.File
}

func NewConsoleSink(stdout bool) Sink {
	if stdout {
		return &ConsoleSink{os.Stdout}
	}
	return &ConsoleSink{os.Stderr}
}

func (l *ConsoleSink) Emit(c *CheckResult) {
	fmt.Fprintf(l.targetStream, "%s [%s]: %t %s", c.TimestampString(), c.Name, c.Result, c.Duration.Round(time.Millisecond))
}
