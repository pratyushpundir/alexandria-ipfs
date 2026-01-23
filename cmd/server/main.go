package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/pratyushpundir/alexandria-services/internal/config"
	"github.com/pratyushpundir/alexandria-services/internal/service"

	pb "github.com/pratyushpundir/alexandria-services/gen/ipfs/v1"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create IPFS client - use mock if no Blockfrost credentials provided
	var ipfsClient service.IPFSClient
	if cfg.BlockfrostProjectID == "" {
		log.Println("WARNING: No BLOCKFROST_IPFS_PROJECT_ID provided, running in mock mode")
		ipfsClient = service.NewMockClient(cfg.IPFSGatewayURL)
	} else {
		log.Println("Using Blockfrost IPFS backend")
		ipfsClient = service.NewBlockfrostClient(&service.BlockfrostConfig{
			ProjectID:  cfg.BlockfrostProjectID,
			BaseURL:    cfg.BlockfrostIPFSBaseURL,
			GatewayURL: cfg.IPFSGatewayURL,
		})
	}

	// Create gRPC server with increased message size for large media uploads (150MB)
	maxMsgSize := 150 * 1024 * 1024
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)

	// Register IPFS service
	ipfsService := service.NewGRPCServer(ipfsClient)
	pb.RegisterIPFSServiceServer(grpcServer, ipfsService)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("alexandria.ipfs.v1.IPFSService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Create listener
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("IPFS gRPC server listening on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down IPFS service...")
	grpcServer.GracefulStop()
	log.Println("IPFS service stopped")
}
