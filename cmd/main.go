package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/PratikkJadhav/MiniObs/api"
	"github.com/PratikkJadhav/MiniObs/receiver"
	"github.com/PratikkJadhav/MiniObs/storage"
	collectorv1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func main() {
	store, err := storage.NewStore("./data")
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	listener, err := net.Listen("tcp", ":4317")
	if err != nil {
		log.Fatalf("failed to listen on 4317: %v", err)
	}

	grpcServer := grpc.NewServer()
	collectorv1.RegisterTraceServiceServer(grpcServer, &receiver.Receiver{Store: store})

	go func() {
		log.Println("Starting gRPC collector")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server crashed %v", err)
		}
	}()

	r := api.NewRouter(store)
	go func() {
		log.Println("Starting HTTP API on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("HTTP server crashed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	store.SaveHint()
}
