package utils

import (
	"fmt"
	"net"
)

const startPort = 6001

func IsPortBusy(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}

func ResolvePort(port int) (int, error) {
	if port > 0 {
		if IsPortBusy(port) {
			return 0, fmt.Errorf("port %d already in use", port)
		}
		return port, nil
	}

	for p := startPort; p < 65535; p++ {
		if !IsPortBusy(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no free port found")
}
