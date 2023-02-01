package slf4go_sentry

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	slog "github.com/go-eden/slf4go"
	"reflect"
	"strings"
	"time"
)

type SentryDriver struct {
	Level slog.Level
}

func NewSentryDriver(level slog.Level) *SentryDriver {
	return &SentryDriver{
		Level: level,
	}
}

func (d *SentryDriver) Name() string {
	return "slf4go-sentry"
}

func (d *SentryDriver) Print(l *slog.Log) {
	if l.Level < d.Level {
		return
	}

	event := sentry.NewEvent()
	event.Timestamp = time.Unix(l.Time/1000000, (l.Time%1000000)*1000)
	event.Logger = l.Logger
	event.Extra = l.Fields

	switch l.Level {
	case slog.TraceLevel:
		event.Level = sentry.LevelDebug
		break
	case slog.DebugLevel:
		event.Level = sentry.LevelDebug
		break
	case slog.InfoLevel:
		event.Level = sentry.LevelInfo
		break
	case slog.WarnLevel:
		event.Level = sentry.LevelWarning
		break
	case slog.ErrorLevel:
		event.Level = sentry.LevelError
		break
	case slog.PanicLevel:
		event.Level = sentry.LevelFatal
		break
	case slog.FatalLevel:
		event.Level = sentry.LevelFatal
		break
	}

	if l.Args != nil {
		for _, arg := range l.Args {
			err, ok := arg.(error)
			if ok {
				event.Exception = append(event.Exception, sentry.Exception{
					Value:      err.Error(),
					Type:       reflect.TypeOf(err).String(),
					Stacktrace: newStacktrace(),
				})
			}
		}
	}

	if l.Format == nil {
		parts := make([]string, 0)
		for _, arg := range l.Args {
			parts = append(parts, fmt.Sprintf("%v", arg))
		}
		event.Message = strings.Join(parts, " ")
	} else {
		event.Message = fmt.Sprintf(*l.Format, l.Args...)
	}

	sentry.CaptureEvent(event)
}

func (d *SentryDriver) GetLevel(logger string) (sl slog.Level) {
	return d.Level
}

func newStacktrace() *sentry.Stacktrace {
	const (
		currentModule = "github.com/ngyewch/slf4go-sentry"
		slf4goModule  = "github.com/go-eden/slf4go"
	)

	st := sentry.NewStacktrace()

	threshold := len(st.Frames) - 1
	// drop current module frames
	for ; threshold > 0 && st.Frames[threshold].Module == currentModule; threshold-- {
	}

outer:
	// try to drop slf4go module frames after logger call point
	for i := threshold; i > 0; i-- {
		if st.Frames[i].Module == slf4goModule {
			for j := i - 1; j >= 0; j-- {
				if st.Frames[j].Module != slf4goModule {
					threshold = j
					break outer
				}
			}
			break
		}
	}

	st.Frames = st.Frames[:threshold+1]

	return st
}
