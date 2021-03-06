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

func TestFileSinkEmit(t *testing.T) {
	fileSink, _ := NewFileSink(map[string]string{"path": "/tmp/someFile"})
	fs := fileSink.(*FileSink)
	r, w, _ := os.Pipe()
	fs.TargetFile = w
	c := Result{time.Now(), Failure, time.Duration(1)}
	fs.Emit("testCheck", "TestCheck", &c)
	w.Close()

	msg, _ := ioutil.ReadAll(r)
	matcher := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d+ \[testCheck\]: Failure [0-9.s]+`

	if match, _ := regexp.Match(matcher, msg); !match {
		fmt.Println(string(msg))
		t.FailNow()
	}
}

func TestNewFileSink(t *testing.T) {
	newSinkTests := []struct {
		name    string
		args    map[string]string
		succeed bool
	}{
		{"no path", map[string]string{"otherParam": "arg"}, false},
		{"relative path", map[string]string{"path": "tmp/testfile"}, false},
		{"path dont exist", map[string]string{"path": "/some_improbablePath_whichCantExist/file1"}, false},
		{"inaccessible path", map[string]string{"path": "/root/afile"}, false},
		{"alright path", map[string]string{"path": "/tmp/testfile1"}, true},
	}

	for _, tt := range newSinkTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileSink(tt.args)
			if !tt.succeed && err == nil {
				t.Errorf("Got %#v, expected to fail but succeeded", tt.args)
			}
			if tt.succeed && err != nil {
				t.Errorf("Got %#v, expected to succeed but failed", tt.args)
			}
		})
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

	c := &Result{time.Now(), Failure, time.Duration(1)}
	fmt.Println(c)
	sink.Emit("a testing check", "ExampleCheck", c)
	// TODO: create test that doesn't need sleeping
	time.Sleep(2 * time.Second) // sleep for FlushInterval
	if influxSink.Client.(*FakeClient).WriteCalled < 1 {
		t.Errorf("Client didnt write or close: %v", influxSink.Client)
	}
}
