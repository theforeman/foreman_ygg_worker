package main

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
)

type Capture struct {
	Captures  []*pb.Data
	NextError error
}

func (c *Capture) Connect() (err error) {
	return c.NextError
}

func (c *Capture) Send(d *pb.Data) (err error) {
	c.Captures = append(c.Captures, d)
	return c.NextError
}

func (c *Capture) Disconnect() {
}

func TestAdd(t *testing.T) {
	ua := NewUpdateAggregator("http://something.somewhere", "12345")
	update := NewExitUpdate(0)
	ua.Add(update)

	if ua.Count != 1 {
		t.Errorf("|%+v| != |%+v|", ua.Count, 1)
	}

	ua.Add(update)
	if ua.Count != 2 {
		t.Errorf("|%+v| != |%+v|", ua.Count, 2)
	}

	if ua.Updates[0] != update {
		t.Errorf("|%+v| != |%+v|", ua.Updates[0], update)
	}
}

func TestSendUpdates(t *testing.T) {
	c := Capture{}
	ua := NewUpdateAggregator("http://something.somewhere", "12345")
	ua.SendUpdates(&c)

	capture := c.Captures[0]
	if capture.ResponseTo != "12345" {
		t.Errorf("|%+v| != |%+v|", capture.ResponseTo, "12345")
	}
	if capture.Directive != "http://something.somewhere" {
		t.Errorf("|%+v| != |%+v|", capture.Directive, "http://something.somewhere")
	}
	if capture.Metadata["Content-Type"] != "application/json" {
		t.Errorf("|%+v| != |%+v|", capture.Metadata["Content-Type"], "application/json")
	}
}

func TestDispatchEvent(t *testing.T) {
	var timeThreshold time.Duration = TimeThreshold * 1000 * 1000
	tests := []struct {
		aggregator  UpdateAggregator
		description string
		update      V1Update
		count       int
		kept        int
	}{
		{
			description: "Without count threshold being reached",
			aggregator:  NewUpdateAggregator("1", "2"),
			update:      NewOutputUpdate("stdout", "hello"),
			kept:        1,
		},
		{
			description: "On exit update being sent",
			aggregator:  NewUpdateAggregator("1", "2"),
			update:      NewExitUpdate(0),
			count:       1,
		},
		{
			description: "When count threshold is reached",
			aggregator: UpdateAggregator{
				Updates: make([]V1Update, CountThreshold),
				Count:   CountThreshold - 1,
			},
			update: NewOutputUpdate("stdout", "hello"),
			count:  1,
		},
		{
			description: "When time threshold is reached",
			aggregator: UpdateAggregator{
				Updates:  make([]V1Update, CountThreshold),
				Count:    CountThreshold - 1,
				LastSend: time.Now().Add(-timeThreshold),
			},
			update: NewOutputUpdate("stdout", "hello"),
			count:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := Capture{}
			test.aggregator.DispatchEvent(test.update, &c)

			got := len(c.Captures)
			if got != test.count {
				t.Errorf("Captured unexpected number of outgoing updates: |%+v| != |%+v|", got, test.count)
			}
			got = test.aggregator.Count
			if got != test.kept {
				t.Errorf("UpdateAggregator retianed unexpected number of events: |%+v| != |%+v|", got, test.kept)
			}
		})
	}
}

func TestAggregate(t *testing.T) {
	tests := []struct {
		aggregator  UpdateAggregator
		description string
		updates     []V1Update
		count       int
		kept        int
	}{
		{
			description: "Reads from channel until it is closed",
			aggregator:  NewUpdateAggregator("1", "2"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := Capture{}
			updates := make(chan V1Update)
			var wg sync.WaitGroup
			wg.Add(1)

			go func() { test.aggregator.Aggregate(updates, &c); wg.Done() }()

			updates <- NewOutputUpdate("stdout", "hello")
			updates <- NewExitUpdate(0)
			close(updates)
			wg.Wait()

			got := len(c.Captures)
			if got != 1 {
				t.Errorf("Captured unexpected number of outgoing updates: |%+v| != |%+v|", got, 1)
			}
			got = test.aggregator.Count
			if got != test.kept {
				t.Errorf("UpdateAggregator retianed unexpected number of events: |%+v| != |%+v|", got, 0)
			}
			data := V1Updates{}
			if err := json.Unmarshal(c.Captures[0].Content, &data); err != nil {
				t.Errorf("Failed to decode update JSON: %v", err)
			}

			got = len(data.Updates)
			if got != 2 {
				t.Errorf("Bulk update contained unexpected number of sub-updates: |%+v| != |%+v|", got, 0)
			}
		})
	}
}
