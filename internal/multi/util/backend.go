package util

import (
	"net"
	"time"
)

const ErrStringBackendNotResponding = "backend not responding"

func IsBackendResponding(backend string) bool {
	conn, err := net.DialTimeout("tcp", backend, time.Second*5)
	if err == nil {
		conn.Close()
	}
	return err == nil
}
