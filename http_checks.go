package healthchecker

import (
	"fmt"
	"net"
	"net/http"
	"time"
	// get https://github.com/go-gcfg/gcfg/tree/v1.2.3 - dep or modules?
)

//func (h *HTTPChecker) AdvancdedHTTPCheck(url string, callback func(*http.Response) bool) {
//// get site, then call custom callback on response
//}

// add timing decorator

//func (n *NetworkChecker) ICMPPingCheck() {
//}

type HTTPChecker struct {
	HTTPClient *http.Client
}

type NetworkChecker struct {
}

func NewHTTPChecker(timeout time.Duration) *HTTPChecker {
	return &HTTPChecker{HTTPClient: &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}}
}

func (h *HTTPChecker) SimpleHTTPCheck(url string) (bool, time.Duration) {
	timeStart := time.Now()
	resp, err := h.HTTPClient.Get(url)
	timeElapsed := time.Since(timeStart)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return false, timeElapsed
		}
		panic(fmt.Sprintf("Error while checking %s", url))
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, timeElapsed
	} else {
		return true, timeElapsed
	}
}
