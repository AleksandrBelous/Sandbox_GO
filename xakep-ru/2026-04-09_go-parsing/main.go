package main

import (
	"fmt"
)

func main() {
	arpResult := retrieveArpTable()
	fmt.Print(arpResult)
}
