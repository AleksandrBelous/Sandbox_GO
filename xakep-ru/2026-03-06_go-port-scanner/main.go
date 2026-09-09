package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"portscan/arp"
	"strconv"
	"time"
)

func loadPortsList(fileName string) ([]string, error) {
	var ports []string

	prtFile, err := os.Open(fileName)
	if err != nil {
		return ports, err
	}
	defer prtFile.Close()

	scanner := bufio.NewScanner(prtFile)

	for scanner.Scan() {
		s := scanner.Text()
		_, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			continue
		}
		ports = append(ports, s)
	}

	err = scanner.Err()
	return ports, err
}

func grabBanner(conn net.Conn, buf []byte) (string, bool) {
	request := fmt.Sprintf("HEAD / HTTP/1.1\r\nHost: %s\r\n\r\n", conn.RemoteAddr())
	conn.Write([]byte(request))

	n, err := conn.Read(buf)
	if err == nil && n > 0 {
		return string(buf[:n]), true
	}

	return "", false
}

func run() (err error) {
	const (
		srcFileName = "targets.txt"
		prtFileName = "ports.txt"
		dstFileName = "results.txt"
	)

	ports, err := loadPortsList(prtFileName)
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		e := fmt.Sprintf("errpr reading %s: empty file\n", prtFileName)
		return errors.New(e)
	}

	srcFile, err := os.Open(srcFileName)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	scanner := bufio.NewScanner(srcFile)

	dstFile, err := os.Create(dstFileName)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	writer := bufio.NewWriter(dstFile)

	defer func() {
		// Планируем сброс буфера и объединяем ошибки
		flushErr := writer.Flush()
		err = errors.Join(err, flushErr)
	}()

	timeout := 250 * time.Millisecond
	buffer := make([]byte, 256)

	arpReader := arp.NewArpReader()

	for scanner.Scan() {
		targetHost := scanner.Text()

		if net.ParseIP(targetHost) == nil {
			ips, err := net.LookupHost(targetHost)
			if err != nil || len(ips) == 0 {
				continue
			}
			targetHost = ips[0]
		}

		mac, isMacFound := arpReader.GetMac(targetHost)
		if isMacFound {
			fmt.Println(mac)
		}

		for _, port := range ports {
			addr := net.JoinHostPort(targetHost, port)

			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err != nil {
				continue
			}

			s := fmt.Sprintf("%s is open\n", addr)
			fmt.Print(s)
			writer.WriteString(s)

			deadline := time.Now().Add(timeout)
			conn.SetDeadline(deadline)

			banner, _ := grabBanner(conn, buffer)
			fmt.Println(banner)
			writer.WriteString(banner)

			conn.Close()
		}
	}

	return scanner.Err()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
