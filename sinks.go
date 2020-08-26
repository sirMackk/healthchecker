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

func NewConsoleSink(args map[string]string) Sink {
	useStdout, ok := args["useStdout"]
	if !ok {
		fmt.Println("Error creating NewConsoleSink - useStdout missing!")
		panic("wot")
	}
	choice, err := strconv.ParseBool(useStdout)
	if err != nil || !choice {
		fmt.Println("Error parsing useStdout - using stdout")
		return &ConsoleSink{TargetStream: os.Stderr}
	}
	return &ConsoleSink{TargetStream: os.Stdout}
}

func (s *ConsoleSink) Emit(c *CheckResult) {
	fmt.Fprintf(s.TargetStream, "%s [%s]: %t %s\n", c.TimestampString(), c.Name, c.Result, c.Duration.Round(time.Millisecond))
}
