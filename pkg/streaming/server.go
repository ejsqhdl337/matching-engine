package streaming

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"matching_engine/pkg/streaming/proto"
)

// Consideration for gRPC server design:
// - We are using a simple gRPC server. This is a good starting point, but it has its limitations.
// - For a production system, we would want to use a more robust gRPC server, with features such as authentication, authorization, and monitoring.
// - We are using a simple in-memory event bus. This is a good choice for performance, but it has a fixed size.
// - For a more flexible solution, we could use a dynamic data structure, such as a slice or a linked list.
type Server struct {
	proto.UnimplementedEventServiceServer
	eventBus *EventBus
	grpcServer *grpc.Server
}

func NewServer(eventBus *EventBus) *Server {
	return &Server{
		eventBus: eventBus,
	}
}

func (s *Server) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
	if err := s.eventBus.AddBatch(req.Payloads); err != nil {
		return nil, err
	}
	return &proto.AddResponse{}, nil
}

func (s *Server) Poll(req *proto.PollRequest, stream proto.EventService_PollServer) error {
	for {
		events, err := s.eventBus.Poll(int(req.MaxEvents))
		if err != nil {
			return err
		}

		var protoEvents []*proto.Event
		for _, event := range events {
			protoEvents = append(protoEvents, &proto.Event{
				Timestamp: event.Timestamp,
				Payload:   event.Payload,
			})
		}

		if err := stream.Send(&proto.PollResponse{Events: protoEvents}); err != nil {
			return err
		}
	}
}

func (s *Server) ListenAndServe(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer()
	proto.RegisterEventServiceServer(s.grpcServer, s)

	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
