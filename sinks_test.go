package healthchecker

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestConsoleSinkEmit(t *testing.T) {
	consoleSink, _ := NewConsoleSink(map[string]string{"useStdout": "true"})
	cs := consoleSink.(*ConsoleSink)
	r, w, _ := os.Pipe()
	cs.TargetStream = w
	c := CheckResult{time.Now(), Failure, time.Duration(1)}
	cs.Emit("testCheck", "TestCheck", &c)
	w.Close()

	msg, _ := ioutil.ReadAll(r)
	matcher := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d+ \[testCheck\]: Failure [0-9.s]+`
	if match, _ := regexp.Match(matcher, msg); !match {
		fmt.Println(string(msg))
		t.FailNow()
	}
}

func TestNewUDPInfluxSinkStartsCollector(t *testing.T) {
	sink, err := NewUDPInfluxSink(map[string]string{
		"addr":          "localhost:9999",
		"flushInterval": "10",
		"flushCount":    "10",
	})
	if err != nil {
		t.Errorf("Couldnt create new UDPInfluxSink: %s", err)
	}
	influxSink := sink.(*UDPInfluxSink)

	if !influxSink.IsCollectorRunning() {
		t.Errorf("Collector goro isn't running after creating UDPInfluxSink")
	}
}

//TODO test emit using count and client interface:
// https://github.com/influxdata/influxdb/blob/1.7/client/v2/client.go#L69

//TODO test emit using flushInterval and client interface
