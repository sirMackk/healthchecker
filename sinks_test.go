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
	consoleSink := NewConsoleSink(true)
	r, w, _ := os.Pipe()
	consoleSink.targetStream = w
	c := CheckResult{time.Now(), "testCheck", false, time.Duration(1)}
	consoleSink.Emit(&c)
	w.Close()

	msg, _ := ioutil.ReadAll(r)
	matcher := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d+ \[testCheck\]: false [0-9.s]+`
	if match, _ := regexp.Match(matcher, msg); !match {
		fmt.Println(string(msg))
		t.FailNow()
	}
}
