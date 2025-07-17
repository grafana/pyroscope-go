package testutil

import (
	"fmt"
	"sync"
)

type TestLogger struct {
	sync.Mutex

	lines []string
}

func NewTestLogger() *TestLogger {
	return &TestLogger{lines: make([]string, 0)}
}

func (t *TestLogger) Debugf(format string, args ...interface{}) { t.putf(format, args...) }
func (t *TestLogger) Infof(format string, args ...interface{})  { t.putf(format, args...) }
func (t *TestLogger) Errorf(format string, args ...interface{}) { t.putf(format, args...) }

func (t *TestLogger) putf(format string, args ...interface{}) {
	t.Lock()
	t.lines = append(t.lines, fmt.Sprintf(format, args...))
	t.Unlock()
}

// Lines returns a copy of the logged lines
func (t *TestLogger) Lines() []string {
	t.Lock()
	defer t.Unlock()

	return append([]string(nil), t.lines...)
}
