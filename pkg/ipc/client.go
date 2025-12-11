package ipc

import (
	"storj.io/drpc/drpcconn"
)

func GetIPCClient() (DRPCIPCClient, error) {
	rawConn, err := getConnection()
	if err != nil {
		return nil, err
	}
	drpccon := drpcconn.New(rawConn)
	return NewDRPCIPCClient(drpccon), nil
}
