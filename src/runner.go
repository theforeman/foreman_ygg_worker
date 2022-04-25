package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"git.sr.ht/~spc/go-log"
	"github.com/google/uuid"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
)

func dispatch(ctx context.Context, d *pb.Data, s *jobStorage) {
	event, prs := d.GetMetadata()["event"]
	if !prs {
		log.Warnln("Message metadata does not contain event field, assuming 'start'")
		event = "start"
	}

	switch event {
	case "start":
		startScript(ctx, d, s)
	case "cancel":
		cancel(ctx, d, s)
	default:
		log.Errorf("Received unknown event '%v'", event)
	}
}

func startScript(ctx context.Context, d *pb.Data, s *jobStorage) {
	jobUUID, jobUUIDP := d.GetMetadata()["job_uuid"]
	if !jobUUIDP {
		log.Warnln("No job uuid found in job's metadata, will not be able to cancel this job")
	}

	script := string(d.GetContent())
	log.Tracef("running script : %#v", script)

	scriptfile, err := ioutil.TempFile("/tmp", "ygg_rex")
	if err != nil {
		log.Errorf("failed to create script tmp file: %v", err)
	}
	defer os.Remove(scriptfile.Name())

	n2, err := scriptfile.Write(d.GetContent())
	if err != nil {
		log.Errorf("failed to write script to tmp file: %v", err)
	}
	log.Debugf("script of %d bytes written in : %#v", n2, scriptfile.Name())

	err = scriptfile.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chmod(scriptfile.Name(), 0700)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("/bin/sh", "-c", scriptfile.Name())
	// cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("cannot connect to stdout: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("cannot connect to stderr: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Errorf("cannot run script: %v", err)
		return
	}
	log.Infof("started script process: %v", cmd.Process.Pid)
	if jobUUIDP {
		s.Set(jobUUID, cmd.Process.Pid)
		defer s.Remove(jobUUID)
	}

	// Dial the Dispatcher
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewDispatcherClient(conn)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { outputCollector(c, d, "stdout", stdout); wg.Done() }()
	go func() { outputCollector(c, d, "stderr", stderr); wg.Done() }()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				sendExitCode(c, d.GetMessageId(), d.GetMetadata()["return_url"], status.ExitStatus())
			}
		} else {
			log.Errorf("script run failed: %v", err)
		}
	} else {
		sendExitCode(c, d.GetMessageId(), d.GetMetadata()["return_url"], 0)
	}
}

func outputCollector(c pb.DispatcherClient, d *pb.Data, stdtype string, pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		msg := scanner.Text()
		log.Tracef("%v message: %v", stdtype, msg)
		sendUpdate(c, d.GetMessageId(), d.GetMetadata()["return_url"], msg, stdtype)
	}
	if err := scanner.Err(); err != nil {
		log.Errorf("cannot read from %v: %v", stdtype, err)
	}
}

var syscallKill = syscall.Kill

func cancel(ctx context.Context, d *pb.Data, s *jobStorage) {
	jobUUID, jobUUIDP := d.GetMetadata()["job_uuid"]
	if !jobUUIDP {
		log.Errorln("No job uuid found in job's metadata, aborting.")
		return
	}

	pid, prs := s.Get(jobUUID)
	if !prs {
		log.Errorf("Cannot cancel unknown job %v", jobUUID)
		return
	}

	log.Infof("Cancelling job %v, sending SIGTERM to process %v", jobUUID, pid)
	if err := syscallKill(pid, syscall.SIGTERM); err != nil {
		log.Errorf("Failed to send SIGTERM to process %v: %v", pid, err)
	}
}

type Update struct {
	Output string `json:"output"`
	Type   string `json:"type"`
}

func sendUpdate(c pb.DispatcherClient, origmsgid string, url string, message string, stdtype string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	content, err := json.Marshal(Update{Output: message, Type: stdtype})
	if err != nil {
		log.Error(err)
		return
	}

	data := &pb.Data{
		MessageId:  uuid.New().String(),
		ResponseTo: origmsgid,
		Content:    content,
		Directive:  url,
	}

	if _, err := c.Send(ctx, data); err != nil {
		log.Error(err)
	}
}

func sendExitCode(c pb.DispatcherClient, origmsgid string, url string, code int) {
	// wait for the other updates
	time.Sleep(time.Duration(2) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data := &pb.Data{
		MessageId:  uuid.New().String(),
		ResponseTo: origmsgid,
		Content:    []byte(fmt.Sprintf("{\"exit_code\": %d}", code)),
		Directive:  url,
	}

	if _, err := c.Send(ctx, data); err != nil {
		log.Error(err)
	}
}
