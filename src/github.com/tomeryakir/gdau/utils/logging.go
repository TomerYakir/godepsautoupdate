package utils

import "fmt"

type Logger struct {
	debug bool
}

func NewLogger(debug bool) *Logger {
	return &Logger{debug}
}

func (l *Logger) PanicWithMessage(msgFormat string, vars ...interface{}) {
	panic(fmt.Sprintf(msgFormat, vars...))
}

func (l *Logger) LogInfo(msgFormat string, vars ...interface{}) {
	fmt.Println(fmt.Sprintf(msgFormat, vars...))
}

func (l *Logger) LogDebug(msgFormat string, vars ...interface{}) {
	if l.debug {
		fmt.Println(fmt.Sprintf(msgFormat, vars...))
	}
}
