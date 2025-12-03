package utils

import (
	"net"
	"testing"
)

func TestIsPortBusy(t *testing.T) {
	// Get a free port by listening and then closing
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	// Port should be busy while listening
	if !IsPortBusy(port) {
		t.Error("expected port to be busy while listening")
	}

	ln.Close()

	// Port should be free after closing
	if IsPortBusy(port) {
		t.Error("expected port to be free after closing")
	}
}

func TestResolvePort_Specific(t *testing.T) {
	// Get a free port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	freePort := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	// Should return the same port if free
	got, err := ResolvePort(freePort)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got != freePort {
		t.Errorf("got %d, want %d", got, freePort)
	}
}

func TestResolvePort_Busy(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	busyPort := ln.Addr().(*net.TCPAddr).Port

	_, err = ResolvePort(busyPort)
	if err == nil {
		t.Error("expected error for busy port")
	}
}

func TestResolvePort_Auto(t *testing.T) {
	port, err := ResolvePort(0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if port < startPort {
		t.Errorf("got %d, want >= %d", port, startPort)
	}
}
