package builtin

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stream wraps gRPC's stream with reconnection.
// In gRPC, connections will auto-reconnect, but streams don't.
//
// This doesn't currently retry Send/Recv calls. It's expected
// that the calling code (i.e. the Node) will be calling Send/Recv
// on an interval anyway, so better not to do complex retries here.
type stream struct {
	client SchedulerServiceClient
	mtx    sync.Mutex
	stream SchedulerService_NodeChatClient
}

func (s *stream) get() (SchedulerService_NodeChatClient, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.stream == nil {
		// There is no active stream, so reconnect.

		// TODO probably want to pass in context.
		chat, err := s.client.NodeChat(context.Background())
		if err != nil {
			return nil, err
		}
		s.stream = chat
	}
	return s.stream, nil
}

func (s *stream) Send(d *Node) error {
	stream, err := s.get()
	if err != nil {
		return err
	}

	err = stream.Send(d)
	// Avoid noisy "context canceled" logs.
	if status.Code(err) == codes.Canceled {
		return nil
	}
	// If the connection died, delete the dead stream.
	if status.Code(err) == codes.Unavailable {
		s.mtx.Lock()
		defer s.mtx.Unlock()
		s.stream = nil
	}
	return err
}

func (s stream) Recv() (*Control, error) {
	stream, err := s.get()
	if err != nil {
		return nil, err
	}

	c, err := stream.Recv()
	// If the connection died, delete the dead stream.
	if status.Code(err) == codes.Unavailable {
		s.mtx.Lock()
		defer s.mtx.Unlock()
		s.stream = nil
	}
	return c, err
}

func (s *stream) Close() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.stream != nil {
		s.stream.CloseSend()
	}
}
