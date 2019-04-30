package main

import "fmt"

func panicWithMessage(msgFormat string, vars ...interface{}) {
	panic(fmt.Sprintf(msgFormat, vars...))
}

func logInfo(msgFormat string, vars ...interface{}) {
	fmt.Println(fmt.Sprintf(msgFormat, vars...))
}

func logDebug(msgFormat string, vars ...interface{}) {
	if debug {
		fmt.Println(fmt.Sprintf(msgFormat, vars...))
	}
}
