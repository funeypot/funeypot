package logs

import "github.com/dgraph-io/badger/v4"

type BadgerLogger struct {
	Logger
}

func NewBadgerLogger(logger Logger) *BadgerLogger {
	return &BadgerLogger{Logger: logger}
}

var _ badger.Logger = (*BadgerLogger)(nil)

func (l *BadgerLogger) Warningf(s string, i ...interface{}) {
	l.Warnf(s, i...)
}
