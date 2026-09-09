//go:build !linux && !windows

package main

func retrieveArpTable() map[string]string {
	result := make(map[string]string)
	// Заглушка для платформенно-неопределённой компиляции
	return result
}
