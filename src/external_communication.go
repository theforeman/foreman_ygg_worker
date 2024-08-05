package main

import (
	"context"
	"errors"
	"time"

	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
