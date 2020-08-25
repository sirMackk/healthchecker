package healthchecker

import (
	"time"
	"fmt"
)

type CheckResult struct {
	Timestamp time.Time
	Name      string
	Result    bool
	Duration  time.Duration
}

func (c *CheckResult) TimestampString() string {
	return fmt.Sprintf(c.Timestamp.Format("2006-01-02 15:04:05.999999"))
}
