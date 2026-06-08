package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/subpop/go-log"

	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
)

func TestForemanServerSend(t *testing.T) {
	var writer strings.Builder
	log.SetOutput(&writer)
	log.SetFlags(0)

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

	time.Sleep(50 * time.Millisecond)

	logOutput := writer.String()
	if !strings.Contains(logOutput, "Received unknown event 'unknown'") {
		t.Errorf("Expected dispatch to be called and log unknown event, got %v", logOutput)
	}
}

func TestForemanServerHandleRx(t *testing.T) {
	var writer strings.Builder
	log.SetOutput(&writer)
	log.SetFlags(0)

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

	err := fs.handleRx(nil, "test-addr", "test-id", "test-response-to", metadata, []byte("test data"))

	if err != nil {
		t.Errorf("handleRx should not return error, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	logOutput := writer.String()
	if !strings.Contains(logOutput, "Received unknown event 'unknown'") {
		t.Errorf("Expected dispatch to be called and log unknown event, got %v", logOutput)
	}
}
