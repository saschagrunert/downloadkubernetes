package logging

import "fmt"

type Log struct {
	Name  string
	Debug bool
}

func NewLog(name string, debug bool) *Log {
	if name == "" {
		name = "default"
	}
	return &Log{
		Name:  name,
		Debug: debug,
	}
}

func (l *Log) Info(msg string) {
	fmt.Println(l.Name, msg)
}

func (l *Log) Infof(msg string, args ...interface{}) {
	l.Info(fmt.Sprintf(msg, args...))
}

func (l *Log) Error(err error) {
	l.Info(fmt.Sprintf("ERROR %+v", err))
}

func (l *Log) Debugf(format string, args ...interface{}) {
	l.Info(fmt.Sprintf("DEBUG %s", fmt.Sprintf(format, args...)))
}
