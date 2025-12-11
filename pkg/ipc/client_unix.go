//go:build !windows

package ipc

import (
	"net"

	"codeberg.org/veya/ermokie/pkg/globals"
)

func getConnection() (net.Conn, error) {
	return net.Dial("unix", globals.UnixIPCSocketPath)
}
