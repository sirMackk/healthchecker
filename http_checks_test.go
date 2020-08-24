package healthchecker

import (
	"testing"
	"fmt"
	"net/http"
	ht "net/http/httptest"
)

func TestSimpleHTTPCheck(t *testing.T) {
	ts := ht.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	defer ts.Close()

	checker := NewHTTPChecker()
	res := checker.SimpleHTTPCheck(ts.URL)
	if res != true {
		t.FailNow()
	}
}
