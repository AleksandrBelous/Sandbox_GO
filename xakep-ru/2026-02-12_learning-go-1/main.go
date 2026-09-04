package main

import (
	"fmt"
	"os"
)

func main() {
	var username string
	username = "Xakep"

	val, isOk := os.LookupEnv("USERNAME")
	if isOk {
		username = val
	}

	fmt.Println("Hello,", username)
}
