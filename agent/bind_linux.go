//go:build linux

package main

import (
	"syscall"
)

func bindToDevice(name string) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		var sockErr error
		err := c.Control(func(fd uintptr) {
			sockErr = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, name)
		})
		if err != nil {
			return err
		}
		return sockErr
	}
}
