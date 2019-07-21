package logging

import "fmt"

type Log struct {
	Name string
}

func NewLog(name string) *Log {
	if name == "" {
		name = "default"
	}
	return &Log{
		Name: name,
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
