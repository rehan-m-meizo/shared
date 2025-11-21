package sagakit

import (
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill"
)

type StdLogger struct {
	*log.Logger
}

func NewStdLogger() StdLogger {
	return StdLogger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
	}
}

func (l StdLogger) Error(msg string, err error, fields watermill.LogFields) {
	l.Printf("[ERROR] %s: %v (%v)", msg, err, fields)
}

func (l StdLogger) Info(msg string, fields watermill.LogFields) {
	l.Printf("[INFO] %s: %v", msg, fields)
}
func (l StdLogger) Debug(msg string, fields watermill.LogFields) {
	l.Printf("[DEBUG] %s: %v", msg, fields)
}
func (l StdLogger) Trace(msg string, fields watermill.LogFields) {
	l.Printf("[TRACE] %s: %v", msg, fields)
}

func (l StdLogger) With(fields watermill.LogFields) watermill.LoggerAdapter { return l }
