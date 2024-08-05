package main

import (
	"context"

	"github.com/redhatinsights/yggdrasil/worker"
	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
)

// foremanServer implements the Worker gRPC service as defined by the yggdrasil
// gRPC protocol. It accepts Assignment messages, unmarshals the data into a
// string, and echoes the content back to the Dispatch service by calling the
// "Finish" method.
type foremanServer struct {
	pb.UnimplementedWorkerServer
	serverContext
}

type serverContext struct {
	jobStorage           *jobStorage
	workingDirectory     string
	externalCommunicator ExternalCommunication
}

// Send implements the "Send" method of the Worker gRPC service.
func (s *foremanServer) Send(ctx context.Context, d *pb.Data) (*pb.Receipt, error) {
	msg := Message{
		MessageID:  d.GetMessageId(),
		ResponseTo: d.GetResponseTo(),
		Metadata:   d.GetMetadata(),
		Directive:  d.GetDirective(),
		Content:    d.GetContent(),
	}
	go dispatch(ctx, msg, &s.serverContext)

	// Respond to the start request that the work was accepted.
	return &pb.Receipt{}, nil
}

func (s *foremanServer) handleRx(w *worker.Worker, addr string, id string, responseTo string, metadata map[string]string, data []byte) error {
	msg := Message{
		Directive:  addr,
		MessageID:  id,
		ResponseTo: responseTo,
		Metadata:   metadata,
		Content:    data,
	}
	go dispatch(context.Background(), msg, &s.serverContext)

	return nil
}
