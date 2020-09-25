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

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc, _ := checker.NewSimpleHTTPCheck(map[string]string{"url": ts.URL})
	res := <-checkerFunc()
	if res.Result != Success {
		t.Errorf("Failed with result: %v", res)
	}
}

func TestSimpleHTTPCheckFail(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - internal error"))
	}))
	defer ts.Close()

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc, _ := checker.NewSimpleHTTPCheck(map[string]string{"url": ts.URL})
	res := <-checkerFunc()
	if res.Result == Success {
		t.FailNow()
	}
}

func TestSimpleHTTPCheckTimeout(t *testing.T) {
	timeout := time.Millisecond * 100
	timeoutSleep := timeout + 50
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeoutSleep)
		w.Write([]byte("Exceeded timeout"))
	}))
	defer ts.Close()

	checker := NewHTTPChecker(100 * time.Millisecond)
	checkerFunc, _ := checker.NewSimpleHTTPCheck(map[string]string{"url": ts.URL})
	res := <-checkerFunc()
	if res.Result == Success || res.Duration < timeoutSleep {
		t.Fail()
	}
}

func TestRegexpHTTPCheckPass(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc, err := checker.NewRegexpHTTPCheck(map[string]string{
		"url":         ts.URL,
		"checkRegexp": "He[a-z]l(o)?",
	})
	if err != nil {
		t.Error(err)
	}

	res := <-checkerFunc()
	if res.Result != Success {
		t.Errorf("Failed with result: %v", res)
	}
}

func TestRegexpHTTPCheckFailStatus(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - internal error"))
	}))
	defer ts.Close()

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc, err := checker.NewRegexpHTTPCheck(map[string]string{
		"url":         ts.URL,
		"checkRegexp": "He[a-z]l(o)?",
	})
	if err != nil {
		t.Error(err)
	}

	res := <-checkerFunc()
	if res.Result == Success {
		t.Fail()
	}
}

func TestRegexpHTTPCheckFailMatch(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bye world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc, err := checker.NewRegexpHTTPCheck(map[string]string{
		"url":         ts.URL,
		"checkRegexp": "He[a-z]l(o)?",
	})
	if err != nil {
		t.Error(err)
	}

	res := <-checkerFunc()
	if res.Result == Success {
		t.Fail()
	}
}

func TestSimpleHTTPCheckNoURLArg(t *testing.T) {
	checker := NewHTTPChecker(1 * time.Second)
	_, err := checker.NewSimpleHTTPCheck(map[string]string{})
	if err == nil {
		t.Fail()
	}
}
