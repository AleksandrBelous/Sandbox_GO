//go:build !linux && !windows

package arp

func retrieveArpTable() map[string]string {
	result := make(map[string]string)
	// Заглушка для платформенно-неопределённой компиляции
	return result
}
