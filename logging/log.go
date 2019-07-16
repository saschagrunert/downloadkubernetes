package logging

import "fmt"

type Log struct{}

func (l *Log) Info(msg string) {
	fmt.Println(msg)
}

func (l *Log) Infof(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func (l *Log) Error(err error) {
	fmt.Printf("%+v", err)
}
