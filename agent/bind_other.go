//go:build !linux

package main

import (
	"fmt"
	"syscall"
)

func bindToDevice(name string) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		return fmt.Errorf("binding to network interface %s is not supported on this platform", name)
	}
}
