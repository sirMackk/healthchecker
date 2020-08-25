package healthchecker

import (
	"fmt"
	"net/http"
	ht "net/http/httptest"
	"testing"
	"time"
	"regexp"
)

func TestSimpleHTTPCheckPass(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(3 * time.Second))
	res := checker.SimpleHTTPCheck(ts.URL)
	if res.Result != true {
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
	res := checker.SimpleHTTPCheck(ts.URL)
	if res.Result == true {
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
	res := checker.SimpleHTTPCheck(ts.URL)
	if res.Result == true || res.Duration < timeoutSleep {
		t.Fail()
	}
}


func TestRegexpHTTPCheckPass(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(1 * time.Second))
	res := checker.RegexpHTTPCheck(ts.URL, regexp.MustCompile(`He[a-z]l(o)?`))
	if res.Result != true {
		t.Fail()
	}
}

func TestRegexpHTTPCheckFailStatus(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - internal error"))
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(1 * time.Second))
	res := checker.RegexpHTTPCheck(ts.URL, regexp.MustCompile(`He[a-z]l(o)?`))
	if res.Result == true {
		t.Fail()
	}
}

func TestRegexpHTTPCheckFailMatch(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bye world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker(time.Duration(1 * time.Second))
	res  := checker.RegexpHTTPCheck(ts.URL, regexp.MustCompile(`He[a-z]l(o)?`))
	if res.Result == true {
		t.Fail()
	}
}
