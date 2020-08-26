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
	targetStream *os.File
}

func NewConsoleSink(useStdout string) Sink {

	choice, err := strconv.ParseBool(useStdout)
	if err != nil {
		panic("NewConsoleSink arguments not true or false")
	}
	if choice {
		return &ConsoleSink{os.Stdout}
	}
	return &ConsoleSink{os.Stderr}
}

func (l *ConsoleSink) Emit(c *CheckResult) {
	fmt.Fprintf(l.targetStream, "%s [%s]: %t %s\n", c.TimestampString(), c.Name, c.Result, c.Duration.Round(time.Millisecond))
}
