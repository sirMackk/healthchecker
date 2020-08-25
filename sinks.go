package healthchecker

//TODO rsyslog sink
//TODO influxdb sink

import (
	"fmt"
	"io"
)

type Sink interface {
	emit(c CheckResult)
}

type LogSink struct {
	targetStream io.Writer
}


func (l *LogSink) emit(c CheckResult) {
	fmt.Fprintf(l.targetStream, "%s [%s]: %t %s", c.Timestamp, c.Name, c.Result, c.Duration)
}
