package main

import (
	"log"
	"net"

	"github.com/PratikkJadhav/MiniObs/receiver"
	collectorv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":4317")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	collectorv1.RegisterTraceServiceServer(s, &receiver.Receiver{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
