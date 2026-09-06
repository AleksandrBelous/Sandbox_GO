package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func grabBanner(conn net.Conn, buf []byte) (string, bool) {
	request := fmt.Sprintf("HEAD / HTTP/1.1\r\nHost: %s\r\n\r\n", conn.RemoteAddr())
	conn.Write([]byte(request))

	n, err := conn.Read(buf)
	if err == nil && n > 0 {
		return string(buf[:n]), true
	}

	return "", false
}

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
	buffer := make([]byte, 256)

	for _, port := range ports {
		addr := net.JoinHostPort(target, port)

		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			continue
		}

		fmt.Printf("Port %s is open\n", port)

		deadline := time.Now().Add(timeout)
		conn.SetDeadline(deadline)

		s, isOk := grabBanner(conn, buffer)
		if isOk {
			fmt.Println(s)
		}
	}

	os.Exit(0)
}
