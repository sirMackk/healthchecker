package main

import (
	"net/http"
	"fmt"
	"time"
	// get https://github.com/go-gcfg/gcfg/tree/v1.2.3 - dep or modules?
)

// break out into non-main package
// introduce interfaces

type HTTPChecker struct {
	HTTPClient http.Client
}

type NetworkChecker struct {
}

func NewHTTPChecker() *HTTPChecker {
	return &HTTPChecker{HTTPClient: &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Duration(3 * time.Second)
	}}
}

func (h *HTTPChecker) SimpleHTTPCheck(url string) {
	resp, err := h.HTTPClient.Get(url)
	if err != nil {
		panic(fmt.Sprintf("Error while checking %s", url))
	}
	if resp.StatusCode <= 200 || resp.StatusCode >= 300 {
		fmt.Println("check failed")
	} else {
		fmt.Println("check succeeded")
	}
}

func (h *HTTPChecker) AdvancdedHTTPCheck(url string, callback func(*http.Response) bool) {
	// get site, then call custom callback on response
}


func (n *NetworkChecker) ICMPPingCheck() {
}
