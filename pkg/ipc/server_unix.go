//go:build !windows

package ipc

import (
	"net"
	"os"

	"codeberg.org/veya/ermokie/pkg/globals"
)

func getListener() (net.Listener, error) {
	_ = os.Remove(globals.UnixIPCSocketPath)
	return net.Listen("unix", globals.UnixIPCSocketPath)
}
