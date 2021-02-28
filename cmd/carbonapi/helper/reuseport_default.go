// +build !linux

package helper

import (
	"syscall"
)

func ReusePort(network, address string, conn syscall.RawConn) error {
	return nil
}
