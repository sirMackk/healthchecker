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
	checkerFunc := checker.NewSimpleHTTPCheck(map[string]string{"url": ts.URL})
	res := checkerFunc()
	if res.Result != true {
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
	checkerFunc := checker.NewSimpleHTTPCheck(map[string]string{"url": ts.URL})
	res := checkerFunc()
	if res.Result == true {
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
	checkerFunc := checker.NewSimpleHTTPCheck(map[string]string{"url": ts.URL})
	res := checkerFunc()
	if res.Result == true || res.Duration < timeoutSleep {
		t.Fail()
	}
}


func TestRegexpHTTPCheckPass(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc := checker.NewRegexpHTTPCheck(map[string]string{
		"url": ts.URL,
		"regexpStr": "He[a-z]l(o)?",
	})

	res := checkerFunc()
	if res.Result != true {
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
	checkerFunc := checker.NewRegexpHTTPCheck(map[string]string{
		"url": ts.URL,
		"regexpStr": "He[a-z]l(o)?",
	})

	res := checkerFunc()
	if res.Result == true {
		t.Fail()
	}
}

func TestRegexpHTTPCheckFailMatch(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bye world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(1 * time.Second)
	checkerFunc := checker.NewRegexpHTTPCheck(map[string]string{
		"url": ts.URL,
		"regexpStr": "He[a-z]l(o)?",
	})

	res := checkerFunc()
	if res.Result == true {
		t.Fail()
	}
}
