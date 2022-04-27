package main

type OutputEvent struct {
	Message string
	Stream  string
}

type ExitEvent struct {
	ExitCode int
}

type V1Update struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	// When type == "output"
	Content *string `json:"content,omitempty"`
	Stream  *string `json:"stream,omitempty"`
	// When type == "exit"
	ExitCode *int `json:"exit_code,omitempty"`
}

type V1Updates struct {
	Version string     `json:"version"`
	Updates []V1Update `json:"updates"`
}
