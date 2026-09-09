//go:build windows

package arp

import (
	"net"
	"os/exec"
	"strings"
)

func retrieveArpTable() map[string]string {
	result := make(map[string]string)

	// Платформенно-зависимый код (Windows)
	arpDataRaw, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return result
	}

	arpDataStr := string(arpDataRaw)

	for line := range strings.Lines(arpDataStr) {
		line = strings.TrimSpace(line)
		if len(line) < 24 {
			continue
		}

		tokens := strings.Fields(line)
		if len(tokens) != 3 {
			continue
		}

		ipStr := tokens[0]
		macStr := tokens[1]

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
