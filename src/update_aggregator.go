package main

import (
	"context"
	"encoding/json"
	"git.sr.ht/~spc/go-log"
	"github.com/google/uuid"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
	"time"
)

const Threshold = 32

type UpdateAggregator struct {
	Updates   []V1Update
	Count     int
	ReturnURL string
	MessageID string
}

func NewUpdateAggregator(returnUrl string, messageID string) UpdateAggregator {
	return UpdateAggregator{
		Updates:   make([]V1Update, Threshold),
		ReturnURL: returnUrl,
		MessageID: messageID,
	}
}

func (a *UpdateAggregator) Aggregate(events <-chan V1Update) {
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewDispatcherClient(conn)
	for event := range events {
		a.DispatchEvent(event, c)
	}
}

func (a *UpdateAggregator) DispatchEvent(event V1Update, c pb.DispatcherClient) {
	a.Updates[a.Count] = event
	a.Count++

	if a.Count == Threshold || event.Type == "exit" {
		// TODO: Or time since last update
		a.SendUpdates(c)
	}
}

func (a *UpdateAggregator) SendUpdates(c pb.DispatcherClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	updates := V1Updates{
		Version: "1",
		Updates: a.Updates[:a.Count],
	}

	payload, err := json.Marshal(updates)
	if err != nil {
		log.Errorf("Failed to marshal json: %v", err)
		return
	}

	data := &pb.Data{
		MessageId:  uuid.New().String(),
		ResponseTo: a.MessageID,
		Content:    payload,
		Metadata: map[string]string{
			"Content-Type": "application/json",
		},
		Directive: a.ReturnURL,
	}

	if _, err := c.Send(ctx, data); err != nil {
		log.Error(err)
	}

	a.Count = 0
}
