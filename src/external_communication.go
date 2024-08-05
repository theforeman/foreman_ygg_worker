package main

import (
	"context"
	"errors"
	"time"

	"github.com/redhatinsights/yggdrasil/worker"
	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ExternalCommunication interface {
	Connect() (err error)
	Send(msg Message) (err error)
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

func (c *YggdrasilGrpc) Send(msg Message) (err error) {
	if c.c == nil {
		return errors.New("Trying to send without established connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	d := &pb.Data{
		MessageId:  msg.MessageID,
		Metadata:   msg.Metadata,
		Content:    msg.Content,
		ResponseTo: msg.ResponseTo,
		Directive:  msg.Directive,
	}

	_, err = c.c.Send(ctx, d)
	return err
}

type YggdrasilDBus struct {
	w *worker.Worker
}

func (c *YggdrasilDBus) Connect() (err error) {
	return nil
}

func (c *YggdrasilDBus) Disconnect() {}

func (c *YggdrasilDBus) Send(msg Message) (err error) {
	_, _, _, err = c.w.Transmit(msg.Directive, msg.MessageID, msg.ResponseTo, msg.Metadata, msg.Content)

	return err
}
