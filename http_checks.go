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

func (h *HTTPChecker) SimpleHTTPCheck(url string) bool {
	timer_start := time.Now()
	resp, err := h.HTTPClient.Get(url)
	time_elapsed := time.Now().Sub(timer_start)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return false
		}
		panic(fmt.Sprintf("Error while checking %s", url))
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Println("check failed")
		//fmt.Println(resp)
		fmt.Println(time_elapsed)
		return false
	} else {
		fmt.Println("check succeeded")
		fmt.Println(resp)
		fmt.Println(time_elapsed)
		return true
	}
}
