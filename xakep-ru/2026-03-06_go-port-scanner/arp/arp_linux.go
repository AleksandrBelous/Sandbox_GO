//go:build linux

package arp

import (
	"bufio"
	"net"
	"os"
	"strings"
)

func retrieveArpTable() map[string]string {
	result := make(map[string]string)

	// Платформенно-зависимый код (Linux)
	arpFile, err := os.Open("/proc/net/arp")
	if err != nil {
		return result
	}
	defer arpFile.Close()

	scanner := bufio.NewScanner(arpFile)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
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
