package healthchecker

import (
	//"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"

	influx_client "github.com/influxdata/influxdb/client/v2"
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

type FakeClient struct {
	WriteCalled int
	CloseCalled int
}

func (fc *FakeClient) Close() error {
	fc.CloseCalled += 1
	return nil
}

func (fc *FakeClient) Write(bp influx_client.BatchPoints) error {
	fc.WriteCalled += 1
	return nil
}

func (fc *FakeClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	return 1 * time.Second, "", nil
}

func (fc *FakeClient) Query(q influx_client.Query) (*influx_client.Response, error) {
	return nil, nil
}

func (fc *FakeClient) QueryAsChunk(q influx_client.Query) (*influx_client.ChunkedResponse, error) {
	return nil, nil
}

func TestUDPInfluxSinkEmitOnCount(t *testing.T) {
	sink, _ := NewUDPInfluxSink(map[string]string{
		"addr":          "localhost:9999",
		"flushInterval": "2",
		"flushCount":    "1",
	})
	influxSink := sink.(*UDPInfluxSink)
	influxSink.Client = &FakeClient{}

	c := &CheckResult{time.Now(), Failure, time.Duration(1)}
	fmt.Println(c)
	sink.Emit("a testing check", "ExampleCheck", c)
	time.Sleep(2 * time.Second) // sleep for FlushInterval
	if influxSink.Client.(*FakeClient).WriteCalled < 1 {
		t.Errorf("Client didnt write or close: %v", influxSink.Client)
	}
}

//TODO test emit using flushInterval and client interface
