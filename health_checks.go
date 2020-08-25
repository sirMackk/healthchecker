package healthchecker

import (
	"time"
)

type CheckResult struct {
	Timestamp time.Time
	Name      string
	Result    bool
	Duration  time.Duration
}
