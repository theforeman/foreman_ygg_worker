package main

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

type V1Updates struct {
	Version string     `json:"version"`
	Updates []V1Update `json:"updates"`
}
