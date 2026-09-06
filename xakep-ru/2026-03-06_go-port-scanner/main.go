package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Target address not specified\n")
		os.Exit(1)
	}

	target := os.Args[1]

	if net.ParseIP(target) == nil {
		fmt.Fprintf(os.Stderr, "%s is not a valid address\n", target)
		os.Exit(1)
	}

	ports := []string{"21", "22", "25", "53", "80", "8080", "443", "110", "143", "3306", "3389"}
	timeout := 250 * time.Millisecond

	for _, port := range ports {
		addr := net.JoinHostPort(target, port)
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			continue
		}
		conn.Close()
		fmt.Printf("Port %s is open\n", port)
	}

	os.Exit(0)
}
