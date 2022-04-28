package main

import "time"

type OutputEvent struct {
	Content *string `json:"content,omitempty"`
	Stream  *string `json:"stream,omitempty"`
}

type ExitEvent struct {
	ExitCode *int `json:"exit_code,omitempty"`
}

type V1Update struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	// When type == "exit"
	ExitEvent
	// When type == "output"
	OutputEvent
}

func NewExitUpdate(code int) V1Update {
	up := V1Update{Type: "exit"}
	up.Timestamp = time.Now().Format(time.RFC3339)
	up.ExitCode = &code
	return up
}

func NewOutputUpdate(stream string, content string) V1Update {
	up := V1Update{Type: "output"}
	up.Timestamp = time.Now().Format(time.RFC3339)
	up.Stream = &stream
	up.Content = &content
	return up
}

type V1Updates struct {
	Version string     `json:"version"`
	Updates []V1Update `json:"updates"`
}
