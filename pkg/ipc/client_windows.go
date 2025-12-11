//go:build windows

package ipc

import (
	"net"

	"codeberg.org/veya/ermokie/pkg/globals"
	"github.com/Microsoft/go-winio"
)

func getConnection() (net.Conn, error) {
	return winio.DialPipe(globals.WindowsIPCPipe, nil)
}
