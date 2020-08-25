package healthchecker

import (
	"fmt"
	"net/http"
	ht "net/http/httptest"
	"testing"
	"time"
)

func TestSimpleHTTPCheckPass(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(3 * time.Second))
	res, _ := checker.SimpleHTTPCheck(ts.URL)
	if res != true {
		t.FailNow()
	}
}

func TestSimpleHTTPCheckFail(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - internal error"))
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(3 * time.Second))
	res, _ := checker.SimpleHTTPCheck(ts.URL)
	if res == true {
		t.FailNow()
	}
}

func TestSimpleHTTPCheckTimeout(t *testing.T) {
	timeout := time.Duration(100)
	timeoutSleep := timeout + 50
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep((timeoutSleep) * time.Millisecond)
		w.Write([]byte("Exceeded timeout"))
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(timeout * time.Millisecond))
	res, timing := checker.SimpleHTTPCheck(ts.URL)
	if res == true || timing < timeoutSleep {
		t.Fail()
	}
}
