package main

import (
	"context"
	"net"
	"os"
	"time"

	"git.sr.ht/~spc/go-log"

	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var yggdDispatchSocketAddr string

func main() {
	// Get initialization values from the environment.
	var ok bool
	yggdDispatchSocketAddr, ok = os.LookupEnv("YGG_SOCKET_ADDR")
	if !ok {
		log.Fatal("Missing YGG_SOCKET_ADDR environment variable")
	}

	yggdLogLevel, ok := os.LookupEnv("YGG_LOG_LEVEL")
	if ok {
		level, ok := log.ParseLevel(yggdLogLevel)
		if ok != nil {
			log.Errorf("Could not parse log level '%v'", yggdLogLevel)
		} else {
			log.SetLevel(level)
		}
	} else {
		// Yggdrasil < 3.0 does not share its configured log level with the
		// workers in any way
		log.SetLevel(log.LevelInfo)
	}

	// Dial the dispatcher on its well-known address.
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create a dispatcher client
	c := pb.NewDispatcherClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Register as a handler of the "foreman" type.
	r, err := c.Register(ctx, &pb.RegistrationRequest{Handler: "foreman", Pid: int64(os.Getpid()), DetachedContent: true})
	if err != nil {
		log.Fatal(err)
	}
	if !r.GetRegistered() {
		log.Fatalf("handler registration failed: %v", err)
	}

	// Listen on the provided socket address.
	l, err := net.Listen("unix", r.GetAddress())
	if err != nil {
		log.Fatal(err)
	}

	// Register as a Worker service with gRPC and start accepting connections.
	s := grpc.NewServer()

	js := newJobStorage()
	fs := foremanServer{
		serverContext: serverContext{jobStorage: &js, workingDirectory: determineWorkdir()},
	}
	pb.RegisterWorkerServer(s, &fs)
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}

func determineWorkdir() string {
	workdir, workdirP := os.LookupEnv("FOREMAN_YGG_WORKER_WORKDIR")
	if workdirP {
		return workdir
	}

	workdir, workdirP = os.LookupEnv("XDG_RUNTIME_DIR")
	if workdirP {
		return workdir
	}

	return "/run"
}
