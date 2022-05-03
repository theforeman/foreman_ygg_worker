package main

import (
	"context"
	"errors"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
	"time"
)

type ExternalCommunication interface {
	Connect() (err error)
	Send(d *pb.Data) (err error)
	Disconnect()
}

type YggdrasilGrpc struct {
	conn *grpc.ClientConn
	c    pb.DispatcherClient
}

func (c *YggdrasilGrpc) Connect() (err error) {
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	c.conn = conn
	c.c = pb.NewDispatcherClient(conn)
	return nil
}

func (c *YggdrasilGrpc) Disconnect() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *YggdrasilGrpc) Send(d *pb.Data) (err error) {
	if c.c == nil {
		return errors.New("Trying to send without established connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.c.Send(ctx, d)
	return err
}
