package utils

import (
	"fmt"
	"log"
	"os"
)

var UseLogger bool

func LogMessage(format string, v ...interface{}) {
	if UseLogger {
		log.Printf(format, v...)
	} else {
		fmt.Printf(format, v...)
	}
}

func QuitSuccess() {
	os.Exit(0)
}

func QuitFatal() {
	os.Exit(-1)
}
