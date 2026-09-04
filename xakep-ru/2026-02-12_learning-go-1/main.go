package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	var (
		envKey   string
		userName string
	)

	if runtime.GOOS == "windows" {
		envKey = "USERNAME"
	} else {
		envKey = "LOGNAME"
	}

	userName, isOk := os.LookupEnv(envKey)
	if !isOk {
		userName = "Xakep"
	}

	fmt.Println("Hello,", userName)
}
