package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func retrieveArpTable() map[string]string {
	result := make(map[string]string)

	// Плдатформенно-зависимый код (Linux)
	arpFile, err := os.Open("/proc/net/arp")
	if err != nil {
		return result
	}
	defer arpFile.Close()

	scanner := bufio.NewScanner(arpFile)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Print(line)
		line = strings.TrimSpace(line)
		fmt.Print(line)
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		fmt.Print(tokens)
		if len(tokens) < 4 {
			continue
		}

		ipStr := tokens[0]
		macStr := tokens[3]

		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}

		_, err := net.ParseMAC(macStr)
		if err != nil {
			continue
		}

		result[ipStr] = macStr
	}

	return result
}
