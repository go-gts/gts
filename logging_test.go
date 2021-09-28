package gts

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/go-gts/gts/internal/testutils"
)

var loggingTests = []struct {
	level  LogLevel
	handle []LogLevel
	ignore []LogLevel
}{
	{
		SILENT,
		[]LogLevel{},
		[]LogLevel{ERROR, WARN, INFO, DEBUG},
	},
	{
		ERROR,
		[]LogLevel{ERROR},
		[]LogLevel{WARN, INFO, DEBUG},
	},
	{
		WARN,
		[]LogLevel{ERROR, WARN},
		[]LogLevel{INFO, DEBUG},
	},
	{
		INFO,
		[]LogLevel{ERROR, WARN, INFO},
		[]LogLevel{DEBUG},
	},
	{
		DEBUG,
		[]LogLevel{ERROR, WARN, INFO, DEBUG},
		[]LogLevel{},
	},
}

func TestLogging(t *testing.T) {
	b := &bytes.Buffer{}
	SetLogWriter(b)

	for i, tt := range loggingTests {
		testutils.RunCase(t, i, func(t *testing.T) {
			SetLogLevel(tt.level)

			for _, level := range tt.handle {
				b.Reset()
				exp := fmt.Sprintf("[%s] %s\n", level, level)
				switch level {
				case DEBUG:
					Debugf("%s\n", level)
				case INFO:
					Infof("%s\n", level)
				case WARN:
					Warnf("%s\n", level)
				case ERROR:
					Errorf("%s\n", level)
				}
				testutils.Diff(t, b.String(), exp)
			}

			for _, level := range tt.handle {
				b.Reset()
				exp := fmt.Sprintf("[%s] %s\n", level, level)
				switch level {
				case DEBUG:
					Debugln(level)
				case INFO:
					Infoln(level)
				case WARN:
					Warnln(level)
				case ERROR:
					Errorln(level)
				}
				testutils.Diff(t, b.String(), exp)
			}

			for _, level := range tt.ignore {
				b.Reset()
				switch level {
				case DEBUG:
					Debugf("%s\n", level)
				case INFO:
					Infof("%s\n", level)
				case WARN:
					Warnf("%s\n", level)
				case ERROR:
					Errorf("%s\n", level)
				}
				if s := b.String(); s != "" {
					t.Errorf("expected log for %s to be empty, got %q", level, s)
				}
			}

			for _, level := range tt.ignore {
				b.Reset()
				switch level {
				case DEBUG:
					Debugln(level)
				case INFO:
					Infoln(level)
				case WARN:
					Warnln(level)
				case ERROR:
					Errorln(level)
				}
				if s := b.String(); s != "" {
					t.Errorf("expected log for %s to be empty, got %q", level, s)
				}
			}
		})
	}

	testutils.Panics(t, func() {
		t.Error(SILENT.String())
	})
}
