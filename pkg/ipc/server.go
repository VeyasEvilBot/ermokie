// server.go
package ipc

import (
	"context"
	"net"

	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

type IPCServer struct {
	Listener  net.Listener
	EventChan chan any
}

type SplitAdvancedEvent struct {
	RunId    int64
	SplitIdx int64
}

func NewIPCServer(eventChan chan any) (*IPCServer, error) {
	lis, err := getListener()
	if err != nil {
		return nil, err
	}
	return &IPCServer{
		Listener:  lis,
		EventChan: eventChan,
	}, nil
}

func (s *IPCServer) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	resp := &Response{Id: req.Id}

	switch payload := req.Payload.(type) {
	case *Request_AdvanceSplit:
		result := s.handleAdvanceSplit(payload.AdvanceSplit)
		s.EventChan <- result
		resp.Payload = &Response_AdvanceSplitResult{AdvanceSplitResult: result}
	case *Request_MoveActiveSplitBack:
		result := s.handleMoveActiveSplitBack(payload.MoveActiveSplitBack)
		s.EventChan <- result
		resp.Payload = &Response_MoveActiveSplitBackResult{MoveActiveSplitBackResult: result}
	}

	return resp, nil
}

func (s *IPCServer) handleAdvanceSplit(req *AdvanceSplit) *AdvanceSplitResult {
	return &AdvanceSplitResult{RunId: req.RunId, ActiveSplitIdx: req.ActiveSplitIdx}
}

func (s *IPCServer) handleMoveActiveSplitBack(req *MoveActiveSplitBack) *MoveActiveSplitBackResult {
	return &MoveActiveSplitBackResult{RunId: req.RunId, ActiveSplitIdx: req.ActiveSplitIdx}
}

func (s *IPCServer) Serve(ctx context.Context) error {
	m := drpcmux.New()
	err := DRPCRegisterIPC(m, s)
	if err != nil {
		return err
	}

	srv := drpcserver.New(m)
	return srv.Serve(ctx, s.Listener)
}
