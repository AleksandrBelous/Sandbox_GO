package main

import (
	"fmt"
	"os/user"
)

func main() {
	var userName string

	u, err := user.Current()
	if err != nil {
		userName = "Xakep"
	} else {
		userName = u.Username
	}

	fmt.Println("Hello,", userName)
}
