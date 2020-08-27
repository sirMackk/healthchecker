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

func (h *HTTPChecker) checkResponse(url string, checkFn func(*http.Response) Outcome) (Outcome, time.Duration) {
	timeStart := time.Now()
	resp, err := h.Client.Get(url)
	timeElapsed := time.Since(timeStart)
	if err != nil {
		log.Debugf("checkResponse to %s failed: %v", url, err)
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return Failure, timeElapsed
		}
		return Error, timeElapsed
	}
	defer resp.Body.Close()
	return checkFn(resp), timeElapsed
}

func (h *HTTPChecker) statusCheck(rsp *http.Response) Outcome {
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return Failure
	}
	return Success
}

func (h *HTTPChecker) contentsCheck(rsp *http.Response, reg *regexp.Regexp) Outcome {
	if res := h.statusCheck(rsp); res != Success {
		return res
	}
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	if !reg.MatchString(string(rspBody)) {
		return Failure
	}
	return Success
}

func (h *HTTPChecker) SimpleHTTPCheck(url string) *CheckResult {
	checkTime := time.Now()
	outcome, duration := h.checkResponse(url, h.statusCheck)
	return &CheckResult{
		Timestamp: checkTime,
		Result:    outcome,
		Duration:  duration,
	}
}

func (h *HTTPChecker) RegexpHTTPCheck(url string, rex *regexp.Regexp) *CheckResult {
	checkTime := time.Now()
	contentsCheckWrapper := func(rsp *http.Response) Outcome {
		return h.contentsCheck(rsp, rex)
	}
	outcome, duration := h.checkResponse(url, contentsCheckWrapper)
	return &CheckResult{
		Timestamp: checkTime,
		Result:    outcome,
		Duration:  duration,
	}
}

func (h *HTTPChecker) NewSimpleHTTPCheck(args map[string]string) (func() *CheckResult, error) {
	var url string
	var ok bool
	if url, ok = args["url"]; !ok {
		return nil, fmt.Errorf("SimpleHTTPCheck missing 'url' parameter")
	}
	return func() *CheckResult {
		return h.SimpleHTTPCheck(url)
	}, nil
}

func (h *HTTPChecker) NewRegexpHTTPCheck(args map[string]string) (func() *CheckResult, error) {
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
	return func() *CheckResult {
		return h.RegexpHTTPCheck(url, regexpArg)
	}, nil
}
