package main

import (
  "context"
  "time"
  // "net/http"

  "git.sr.ht/~spc/go-log"
  "github.com/google/uuid"
  pb "github.com/redhatinsights/yggdrasil/protocol"
  "google.golang.org/grpc"
)

// foremanServer implements the Worker gRPC service as defined by the yggdrasil
// gRPC protocol. It accepts Assignment messages, unmarshals the data into a
// string, and echoes the content back to the Dispatch service by calling the
// "Finish" method.
type foremanServer struct {
  pb.UnimplementedWorkerServer
}

// Send implements the "Send" method of the Worker gRPC service.
func (s *foremanServer) Send(ctx context.Context, d *pb.Data) (*pb.Receipt, error) {
  go func() {
    log.Tracef("received data: %#v", d)
    message := string(d.GetContent())
    log.Infof("message is: %#v", message)

    // Dial the Dispatcher and call "Finish"
    conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
    if err != nil {
      log.Fatal(err)
    }
    defer conn.Close()

    // Create a client of the Dispatch service
    c := pb.NewDispatcherClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // Create a data message to send back to the dispatcher.
    data1 := &pb.Data{
      MessageId:  uuid.New().String(),
      ResponseTo: d.GetMessageId(),
      Metadata:   d.GetMetadata(),
      Content:    []byte("{\"json\":\"I loved what youve send me\"}"),
      Directive:  d.GetDirective(),
    }

    // Call "Send"
    if _, err := c.Send(ctx, data1); err != nil {
      log.Error(err)
    }

    data2 := &pb.Data{
      MessageId:  uuid.New().String(),
      ResponseTo: d.GetMessageId(),
      Metadata:   d.GetMetadata(),
      Content:    []byte("{\"json\":\"I loved what youve send me second time\"}"),
      Directive:  d.GetDirective(),
    }

    if _, err := c.Send(ctx, data2); err != nil {
      log.Error(err)
    }
  }()

  // Respond to the start request that the work was accepted.
  return &pb.Receipt{}, nil
}
