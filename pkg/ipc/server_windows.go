//go:build windows

package ipc

import (
	"net"

	"codeberg.org/veya/ermokie/pkg/globals"
	"github.com/Microsoft/go-winio"
)

func getListener() (net.Listener, error) {
	return winio.ListenPipe(globals.WindowsIPCPipe, nil)
}
