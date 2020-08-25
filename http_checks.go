package healthchecker

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"
)

type HTTPChecker struct {
	Client *http.Client
}

func NewHTTPChecker(timeout time.Duration) *HTTPChecker {
	checker := HTTPChecker{
		Client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: timeout,
		},
	}
	return &checker
}

// TODO better func name
func (h *HTTPChecker) checkRequest(url string, checkFn func(*http.Response) bool) (bool, time.Duration) {
	timeStart := time.Now()
	resp, err := h.Client.Get(url)
	timeElapsed := time.Since(timeStart)
	if err != nil {
		// TODO what other errors might this return?
		// default timeout should be default/10s and allow checks to define custom timeout
		// that measures timeElapsed and compares to custom value.
		if err, ok := err.(net.Error); ok && err.Timeout() {
			//TODO make into a debugging log
			//fmt.Printf("get error: %v", err)
			return false, timeElapsed
		}
		panic(fmt.Sprintf("Error while checking %s", url))
	}
	defer resp.Body.Close()
	return checkFn(resp), timeElapsed
}

func (h *HTTPChecker) statusCheck(rsp *http.Response) bool {
	return rsp.StatusCode >= 200 && rsp.StatusCode < 300
}

func (h *HTTPChecker) contentsCheck(rsp *http.Response, reg *regexp.Regexp) bool {
	if !h.statusCheck(rsp) {
		return false
	}
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	return reg.MatchString(string(rspBody))
}

func (h *HTTPChecker) SimpleHTTPCheck(url string) *CheckResult {
	checkName := fmt.Sprintf("SimpleHTTPCheck: %s", url)
	checkTime := time.Now()
	success, duration := h.checkRequest(url, h.statusCheck)
	return &CheckResult{checkTime, checkName, success, duration}
}

func (h *HTTPChecker) RegexpHTTPCheck(url string, rex *regexp.Regexp) *CheckResult {
	checkName := fmt.Sprintf("RegexpHTTPCheck: %s", url)
	checkTime := time.Now()
	contentsCheckWrapper := func(rsp *http.Response) bool {
		return h.contentsCheck(rsp, rex)
	}
	success, duration := h.checkRequest(url, contentsCheckWrapper)
	return &CheckResult{checkTime, checkName, success, duration}
}

//func (h *HTTPChecker) NewSimpleHTTPCheck(args map[string]string) func() *CheckResult {
func (h *HTTPChecker) NewSimpleHTTPCheck(args map[string]string) func() *CheckResult {
	var url string
	var ok bool
	if url, ok = args["url"]; !ok {
		panic("SimpleHTTPCheck missing 'url' parameter")
	}
	return func() *CheckResult {
		return h.SimpleHTTPCheck(url)
	}
}

func (h *HTTPChecker) NewRegexpHTTPCheck(args map[string]string) func() *CheckResult {
	// TODO abstract argument checking
	var url, regexpStr string
	var ok bool
	if url, ok = args["url"]; !ok {
		panic("RegexpHTTPCheck missing 'url' parameter")
	}
	//TODO bad variable name
	if regexpStr, ok = args["regexpStr"]; !ok {
		panic("RegexpHTTPCheck missing 'regexpStr' parameter")
	}
	regexpArg := regexp.MustCompile(regexpStr)
	return func() *CheckResult {
		return h.RegexpHTTPCheck(url, regexpArg)
	}
}
