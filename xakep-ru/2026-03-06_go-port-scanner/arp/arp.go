package arp

// Делаем структуру неэкспортируемой, записав ее имя с маленькой буквы
type arpReader struct {
	cache map[string]string
}

// NewArpReader is factory function
func NewArpReader() *arpReader {
	a := &arpReader{}
	a.cache = retrieveArpTable()
	return a
}

func (a *arpReader) GetMac(hostIP string) (string, bool) {
	if a.cache == nil {
		return "", false
	}

	val, isFound := a.cache[hostIP]
	return val, isFound
}
