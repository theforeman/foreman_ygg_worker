package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/subpop/go-log"
)

func setupLogCapture(t *testing.T, w io.Writer, level log.Level) {
	origOut := log.Writer()
	origFlags := log.Flags()
	origLevel := log.CurrentLevel()
	t.Cleanup(func() {
		log.SetOutput(origOut)
		log.SetFlags(origFlags)
		log.SetLevel(origLevel)
	})
	log.SetOutput(w)
	log.SetFlags(0)
	log.SetLevel(level)
}

func TestDispatch(t *testing.T) {
	tests := []struct {
		description  string
		input        Message
		want         string
		setupContext func(*testing.T) serverContext
	}{
		{
			description: "Unknown event",
			want:        "Received unknown event 'test'\n",
			input: Message{
				Metadata: map[string]string{
					"event": "test",
				},
			},
			setupContext: func(t *testing.T) serverContext {
				js := newJobStorage()
				return serverContext{jobStorage: &js}
			},
		},
		{
			description: "Missing event key defaults to start",
			want:        "Message metadata does not contain event field, assuming 'start'",
			input: Message{
				Metadata: map[string]string{
					"return_url": "http://test",
				},
				Content: json.RawMessage("#!/bin/sh\necho test"),
			},
			setupContext: func(t *testing.T) serverContext {
				js := newJobStorage()
				return serverContext{
					jobStorage:           &js,
					workingDirectory:     t.TempDir(),
					externalCommunicator: &Capture{},
				}
			},
		},
		{
			description: "Cancel event with no job_uuid",
			want:        "No job uuid found in job's metadata, aborting.\n",
			input: Message{
				Metadata: map[string]string{
					"event": "cancel",
				},
			},
			setupContext: func(t *testing.T) serverContext {
				js := newJobStorage()
				return serverContext{jobStorage: &js}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var writer strings.Builder
			setupLogCapture(t, &writer, log.LevelDebug)

			sc := test.setupContext(t)

			dispatch(context.Background(), test.input, &sc)

			got := writer.String()
			if !strings.Contains(got, test.want) {
				t.Errorf("Expected log to contain |%+v|, got |%+v|", test.want, got)
			}
		})
	}
}

func TestCancel(t *testing.T) {
	jobStorage := newJobStorage()
	jobStorage.Set("123-abc", 12345)

	tests := []struct {
		description string
		input       Message
		killError   error
		want        string
	}{
		{
			want: "No job uuid found in job's metadata, aborting.\n",
		},
		{
			want: "Cannot cancel unknown job 456-qwe\n",
			input: Message{
				Metadata: map[string]string{
					"job_uuid": "456-qwe",
				},
			},
		},
		{
			want: "Cancelling job 123-abc, sending SIGTERM to process 12345\n",
			input: Message{
				Metadata: map[string]string{
					"job_uuid": "123-abc",
				},
			},
		},
		{
			want: "Cancelling job 123-abc, sending SIGTERM to process 12345\n" +
				"Failed to send SIGTERM to process 12345: kill: (12345) - No such process\n",
			input: Message{
				Metadata: map[string]string{
					"job_uuid": "123-abc",
				},
			},
			killError: errors.New("kill: (12345) - No such process"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var writer strings.Builder
			setupLogCapture(t, &writer, log.LevelInfo)

			syscallKill = func(pid int, sig syscall.Signal) error { return test.killError }
			defer func() { syscallKill = syscall.Kill }()

			sc := serverContext{jobStorage: &jobStorage}

			cancel(context.Background(), test.input, &sc)

			got := writer.String()
			if got != test.want {
				t.Errorf("|%+v| != |%+v|", got, test.want)
			}
		})
	}
}

func TestReportStartError(t *testing.T) {
	var writer strings.Builder
	setupLogCapture(t, &writer, log.LevelInfo)

	updates := make(chan V1Update, 2)
	message := "test error message"

	reportStartError(message, updates)

	outputUpdate := <-updates
	if outputUpdate.Type != "output" {
		t.Errorf("First update should be output, got %v", outputUpdate.Type)
	}
	if outputUpdate.Stream == nil || *outputUpdate.Stream != "stderr" {
		t.Errorf("Output stream should be stderr, got %v", outputUpdate.Stream)
	}
	if outputUpdate.Content == nil || *outputUpdate.Content != message {
		t.Errorf("Output content should be %v, got %v", message, outputUpdate.Content)
	}

	exitUpdate := <-updates
	if exitUpdate.Type != "exit" {
		t.Errorf("Second update should be exit, got %v", exitUpdate.Type)
	}
	if exitUpdate.ExitCode == nil || *exitUpdate.ExitCode != preStartErrorExitCode {
		t.Errorf("Exit code should be %v, got %v", preStartErrorExitCode, exitUpdate.ExitCode)
	}

	_, ok := <-updates
	if ok {
		t.Errorf("Channel should be closed")
	}

	logOutput := writer.String()
	if !strings.Contains(logOutput, message) {
		t.Errorf("Log should contain error message, got %v", logOutput)
	}
}

type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("test read error")
}

func (e *errorReadCloser) Close() error {
	return nil
}

func TestOutputCollector(t *testing.T) {
	tests := []struct {
		description string
		setup       func() (io.ReadCloser, chan V1Update)
		validate    func(*testing.T, chan V1Update)
	}{
		{
			description: "Normal read",
			setup: func() (io.ReadCloser, chan V1Update) {
				r, w := io.Pipe()
				updates := make(chan V1Update, 10)
				go func() {
					_, _ = w.Write([]byte("hello world"))
					_ = w.Close()
				}()
				return r, updates
			},
			validate: func(t *testing.T, updates chan V1Update) {
				update := <-updates
				if update.Type != "output" {
					t.Errorf("Expected output update, got %v", update.Type)
				}
				if update.Content == nil || !strings.Contains(*update.Content, "hello world") {
					t.Errorf("Expected content to contain 'hello world', got %v", update.Content)
				}
			},
		},
		{
			description: "Empty pipe",
			setup: func() (io.ReadCloser, chan V1Update) {
				r, w := io.Pipe()
				updates := make(chan V1Update, 10)
				go func() {
					_ = w.Close()
				}()
				return r, updates
			},
			validate: func(t *testing.T, updates chan V1Update) {
				select {
				case update := <-updates:
					t.Errorf("Expected no updates, got %v", update)
				default:
				}
			},
		},
		{
			description: "Read error",
			setup: func() (io.ReadCloser, chan V1Update) {
				updates := make(chan V1Update, 10)
				return &errorReadCloser{}, updates
			},
			validate: func(t *testing.T, updates chan V1Update) {
				select {
				case update := <-updates:
					t.Errorf("Expected no updates on error, got %v", update)
				default:
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			pipe, updates := test.setup()
			outputCollector("stdout", pipe, updates)
			test.validate(t, updates)
		})
	}
}

func TestStartScript(t *testing.T) {
	tests := []struct {
		description      string
		script           string
		metadata         map[string]string
		validateCaptures func(*testing.T, []Message)
		expectLog        string
	}{
		{
			description: "Script with exit 0",
			script:      "#!/bin/sh\necho hello",
			metadata: map[string]string{
				"return_url": "http://test",
				"event":      "start",
				"job_uuid":   "test-uuid-1",
			},
			validateCaptures: func(t *testing.T, captures []Message) {
				if len(captures) == 0 {
					t.Fatal("Expected at least one message")
				}
				var updates V1Updates
				if err := json.Unmarshal(captures[0].Content, &updates); err != nil {
					t.Fatalf("Failed to unmarshal updates: %v", err)
				}
				foundOutput := false
				foundExit := false
				for _, u := range updates.Updates {
					if u.Type == "output" && u.Content != nil && strings.Contains(*u.Content, "hello") {
						foundOutput = true
					}
					if u.Type == "exit" && u.ExitCode != nil && *u.ExitCode == 0 {
						foundExit = true
					}
				}
				if !foundOutput {
					t.Error("Expected output update with 'hello'")
				}
				if !foundExit {
					t.Error("Expected exit update with code 0")
				}
			},
		},
		{
			description: "Script with non-zero exit",
			script:      "#!/bin/sh\nexit 42",
			metadata: map[string]string{
				"return_url": "http://test",
				"event":      "start",
				"job_uuid":   "test-uuid-2",
			},
			validateCaptures: func(t *testing.T, captures []Message) {
				if len(captures) == 0 {
					t.Fatal("Expected at least one message")
				}
				var updates V1Updates
				if err := json.Unmarshal(captures[0].Content, &updates); err != nil {
					t.Fatalf("Failed to unmarshal updates: %v", err)
				}
				foundExit := false
				for _, u := range updates.Updates {
					if u.Type == "exit" && u.ExitCode != nil && *u.ExitCode == 42 {
						foundExit = true
					}
				}
				if !foundExit {
					t.Error("Expected exit update with code 42")
				}
			},
		},
		{
			description: "Script with stderr output",
			script:      "#!/bin/sh\necho error >&2",
			metadata: map[string]string{
				"return_url": "http://test",
				"event":      "start",
				"job_uuid":   "test-uuid-3",
			},
			validateCaptures: func(t *testing.T, captures []Message) {
				if len(captures) == 0 {
					t.Fatal("Expected at least one message")
				}
				var updates V1Updates
				if err := json.Unmarshal(captures[0].Content, &updates); err != nil {
					t.Fatalf("Failed to unmarshal updates: %v", err)
				}
				foundStderr := false
				for _, u := range updates.Updates {
					if u.Type == "output" && u.Stream != nil && *u.Stream == "stderr" && u.Content != nil && strings.Contains(*u.Content, "error") {
						foundStderr = true
					}
				}
				if !foundStderr {
					t.Error("Expected stderr output update with 'error'")
				}
			},
		},
		{
			description: "Script with both stdout and stderr",
			script:      "#!/bin/sh\necho out; echo err >&2",
			metadata: map[string]string{
				"return_url": "http://test",
				"event":      "start",
				"job_uuid":   "test-uuid-4",
			},
			validateCaptures: func(t *testing.T, captures []Message) {
				if len(captures) == 0 {
					t.Fatal("Expected at least one message")
				}
				var updates V1Updates
				if err := json.Unmarshal(captures[0].Content, &updates); err != nil {
					t.Fatalf("Failed to unmarshal updates: %v", err)
				}
				foundStdout := false
				foundStderr := false
				for _, u := range updates.Updates {
					if u.Type == "output" && u.Stream != nil {
						if *u.Stream == "stdout" && u.Content != nil && strings.Contains(*u.Content, "out") {
							foundStdout = true
						}
						if *u.Stream == "stderr" && u.Content != nil && strings.Contains(*u.Content, "err") {
							foundStderr = true
						}
					}
				}
				if !foundStdout {
					t.Error("Expected stdout output")
				}
				if !foundStderr {
					t.Error("Expected stderr output")
				}
			},
		},
		{
			description: "No job_uuid in metadata",
			script:      "#!/bin/sh\necho test",
			metadata: map[string]string{
				"return_url": "http://test",
				"event":      "start",
			},
			validateCaptures: func(t *testing.T, captures []Message) {
				if len(captures) == 0 {
					t.Fatal("Expected at least one message")
				}
			},
			expectLog: "No job uuid found in job's metadata",
		},
		{
			description: "Unknown effective_user",
			script:      "#!/bin/sh\necho test",
			metadata: map[string]string{
				"return_url":     "http://test",
				"event":          "start",
				"job_uuid":       "test-uuid-5",
				"effective_user": "nonexistent_user_xyz",
			},
			validateCaptures: func(t *testing.T, captures []Message) {
				if len(captures) == 0 {
					t.Fatal("Expected at least one message")
				}
				var updates V1Updates
				if err := json.Unmarshal(captures[0].Content, &updates); err != nil {
					t.Fatalf("Failed to unmarshal updates: %v", err)
				}
				foundError := false
				foundExit := false
				for _, u := range updates.Updates {
					if u.Type == "output" && u.Stream != nil && *u.Stream == "stderr" && u.Content != nil && strings.Contains(*u.Content, "Unknown effective user") {
						foundError = true
					}
					if u.Type == "exit" && u.ExitCode != nil && *u.ExitCode == preStartErrorExitCode {
						foundExit = true
					}
				}
				if !foundError {
					t.Error("Expected error output about unknown effective user")
				}
				if !foundExit {
					t.Errorf("Expected exit code %v", preStartErrorExitCode)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var writer strings.Builder
			setupLogCapture(t, &writer, log.LevelInfo)

			capture := &Capture{}
			js := newJobStorage()
			sc := serverContext{
				jobStorage:           &js,
				workingDirectory:     t.TempDir(),
				externalCommunicator: capture,
			}

			msg := Message{
				Metadata: test.metadata,
				Content:  json.RawMessage(test.script),
			}

			startScript(context.Background(), msg, &sc)

			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) && len(capture.Captures) == 0 {
				time.Sleep(5 * time.Millisecond)
			}

			test.validateCaptures(t, capture.Captures)

			if test.expectLog != "" {
				logOutput := writer.String()
				if !strings.Contains(logOutput, test.expectLog) {
					t.Errorf("Expected log to contain '%v', got '%v'", test.expectLog, logOutput)
				}
			}
		})
	}
}
