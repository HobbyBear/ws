package ws

import "fmt"

type Log interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

var defaultLogger = &DefaultLogger{}

type DefaultLogger struct {
}

func (d DefaultLogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[Info] "+format+"\n", args...)
}

func (d DefaultLogger) Debugf(format string, args ...interface{}) {
	fmt.Printf("[Debug] "+format+"\n", args...)
}

func (d DefaultLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[Error] "+format+"\n", args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}
