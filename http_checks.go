package healthchecker

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
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

func (h *HTTPChecker) checkAndTimeResponse(url string, checkFn func(*http.Response) ResultCode) (ResultCode, time.Duration) {
	timeStart := time.Now()
	resp, err := h.Client.Get(url)
	timeElapsed := time.Since(timeStart)
	if err != nil {
		log.Debugf("checkAndTimeResponse to %s failed: %v", url, err)
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return Failure, timeElapsed
		}
		return Error, timeElapsed
	}
	defer resp.Body.Close()
	return checkFn(resp), timeElapsed
}

func (h *HTTPChecker) checkStatusCode(rsp *http.Response) ResultCode {
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return Failure
	}
	return Success
}

func (h *HTTPChecker) checkBodyForRegexp(rsp *http.Response, reg *regexp.Regexp) ResultCode {
	if res := h.checkStatusCode(rsp); res != Success {
		return res
	}
	// TODO: stream the Body?
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	if !reg.MatchString(string(rspBody)) {
		return Failure
	}
	return Success
}

func (h *HTTPChecker) SimpleHTTPCheck(url string) *Result {
	checkTime := time.Now()
	outcome, duration := h.checkAndTimeResponse(url, h.checkStatusCode)
	return &Result{
		Timestamp: checkTime,
		Result:    outcome,
		Duration:  duration,
	}
}

func (h *HTTPChecker) RegexpHTTPCheck(url string, rex *regexp.Regexp) *Result {
	checkTime := time.Now()
	bodyCheckWrapper := func(rsp *http.Response) ResultCode {
		return h.checkBodyForRegexp(rsp, rex)
	}
	outcome, duration := h.checkAndTimeResponse(url, bodyCheckWrapper)
	return &Result{
		Timestamp: checkTime,
		Result:    outcome,
		Duration:  duration,
	}
}

func (h *HTTPChecker) NewSimpleHTTPCheck(args map[string]string) (func() *Result, error) {
	var url string
	var ok bool
	if url, ok = args["url"]; !ok {
		return nil, fmt.Errorf("SimpleHTTPCheck missing 'url' parameter")
	}
	return func() *Result {
		return h.SimpleHTTPCheck(url)
	}, nil
}

func (h *HTTPChecker) NewRegexpHTTPCheck(args map[string]string) (func() *Result, error) {
	// TODO abstract argument checking
	var url, checkRegexp string
	var ok bool
	if url, ok = args["url"]; !ok {
		return nil, fmt.Errorf("RegexpHTTPCheck missing 'url' parameter")
	}
	if checkRegexp, ok = args["checkRegexp"]; !ok {
		return nil, fmt.Errorf("RegexpHTTPCheck missing 'checkRegexp' parameter")
	}
	regexpArg := regexp.MustCompile(checkRegexp)
	return func() *Result {
		return h.RegexpHTTPCheck(url, regexpArg)
	}, nil
}
