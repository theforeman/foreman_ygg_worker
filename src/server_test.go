package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/subpop/go-log"

	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
)

func waitForLog(writer *strings.Builder, substr string) bool {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(writer.String(), substr) {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

func TestForemanServerSend(t *testing.T) {
	var writer strings.Builder
	setupLogCapture(t, &writer, log.LevelInfo)

	js := newJobStorage()
	fs := foremanServer{
		serverContext: serverContext{
			jobStorage:           &js,
			workingDirectory:     t.TempDir(),
			externalCommunicator: &Capture{},
		},
	}

	data := &pb.Data{
		MessageId:  "test-message-id",
		ResponseTo: "test-response-to",
		Metadata: map[string]string{
			"event": "unknown",
		},
		Directive: "test-directive",
		Content:   []byte("test content"),
	}

	receipt, err := fs.Send(context.Background(), data)

	if err != nil {
		t.Errorf("Send should not return error, got %v", err)
	}

	if receipt == nil {
		t.Errorf("Send should return non-nil Receipt")
	}

	expected := "Received unknown event 'unknown'"
	if !waitForLog(&writer, expected) {
		t.Errorf("Expected dispatch to be called and log unknown event, got %v", writer.String())
	}
}

func TestForemanServerHandleRx(t *testing.T) {
	var writer strings.Builder
	setupLogCapture(t, &writer, log.LevelInfo)

	js := newJobStorage()
	fs := foremanServer{
		serverContext: serverContext{
			jobStorage:           &js,
			workingDirectory:     t.TempDir(),
			externalCommunicator: &Capture{},
		},
	}

	metadata := map[string]string{
		"event": "unknown",
	}

	// nil worker: handleRx does not use the worker argument currently
	err := fs.handleRx(nil, "test-addr", "test-id", "test-response-to", metadata, []byte("test data"))

	if err != nil {
		t.Errorf("handleRx should not return error, got %v", err)
	}

	expected := "Received unknown event 'unknown'"
	if !waitForLog(&writer, expected) {
		t.Errorf("Expected dispatch to be called and log unknown event, got %v", writer.String())
	}
}
