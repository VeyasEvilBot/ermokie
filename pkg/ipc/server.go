package ipc

import (
	"context"
	"net"
	"os"

	"codeberg.org/veya/ermokie/pkg/globals"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type UnixIpcServer struct {
	socketPath string
	EventChan  chan any
}

type SplitAdvancedEvent struct {
	RunId    int64
	SplitIdx int64
}

func NewUnixIpcServer(eventChan chan any) *UnixIpcServer {
	return &UnixIpcServer{
		socketPath: globals.UnixIPCSocketPath,
		EventChan:  eventChan,
	}
}

func (s *UnixIpcServer) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	resp := &Response{Id: req.Id}

	switch payload := req.Payload.(type) {
	case *Request_AdvanceSplit:
		result := s.handleAdvanceSplit(payload.AdvanceSplit)
		s.EventChan <- result
		resp.Payload = &Response_AdvanceSplitResult{AdvanceSplitResult: result}
	}

	return resp, nil
}

func (s *UnixIpcServer) handleAdvanceSplit(req *AdvanceSplit) *AdvanceSplitResult {
	return &AdvanceSplitResult{RunId: req.RunId, ActiveSplitIdx: req.ActiveSplitIdx + 1}
}

func (s *UnixIpcServer) Serve(ctx context.Context) error {
	_ = os.Remove(s.socketPath)
	lis, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}
	defer lis.Close()

	m := drpcmux.New()
	err = DRPCRegisterIPC(m, s)
	if err != nil {
		return err
	}

	srv := drpcserver.New(m)
	return srv.Serve(ctx, lis)
}
