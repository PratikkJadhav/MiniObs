package main

import (
	"log"
	"net"

	"github.com/PratikkJadhav/MiniObs/receiver"
	"github.com/PratikkJadhav/MiniObs/storage"
	collectorv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":4317")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	store, err := storage.NewStore("./data")
	if err != nil {
		log.Fatalf("failed to create store: %v", err)
	}
	rec := &receiver.Receiver{Store: store}

	collectorv1.RegisterTraceServiceServer(s, rec)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
