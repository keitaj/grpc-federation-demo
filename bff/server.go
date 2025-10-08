package main

import (
	"fmt"
	"log"
	"net"
	"os"

	backend "github.com/keitaj/grpc-federation-demo/proto/backend"
	pb "github.com/keitaj/grpc-federation-demo/proto/bff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type clientFactory struct {
	userClient    backend.UserServiceClient
	productClient backend.ProductServiceClient
	orderClient   backend.OrderServiceClient
}

func (f *clientFactory) Order_OrderServiceClient(cfg pb.BFFServiceClientConfig) (backend.OrderServiceClient, error) {
	return f.orderClient, nil
}

func (f *clientFactory) Product_ProductServiceClient(cfg pb.BFFServiceClientConfig) (backend.ProductServiceClient, error) {
	return f.productClient, nil
}

func (f *clientFactory) User_UserServiceClient(cfg pb.BFFServiceClientConfig) (backend.UserServiceClient, error) {
	return f.userClient, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	port := getEnv("PORT", "50054")
	userServiceAddr := getEnv("USER_SERVICE_ADDR", "localhost:50051")
	productServiceAddr := getEnv("PRODUCT_SERVICE_ADDR", "localhost:50052")
	orderServiceAddr := getEnv("ORDER_SERVICE_ADDR", "localhost:50053")

	log.Printf("Starting BFF server on port %s", port)
	log.Printf("Connecting to User Service at %s", userServiceAddr)
	log.Printf("Connecting to Product Service at %s", productServiceAddr)
	log.Printf("Connecting to Order Service at %s", orderServiceAddr)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC client connections
	userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userConn.Close()

	productConn, err := grpc.NewClient(productServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer productConn.Close()

	orderConn, err := grpc.NewClient(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Order Service: %v", err)
	}
	defer orderConn.Close()

	// Create client factory
	factory := &clientFactory{
		userClient:    backend.NewUserServiceClient(userConn),
		productClient: backend.NewProductServiceClient(productConn),
		orderClient:   backend.NewOrderServiceClient(orderConn),
	}

	// Create BFF service with grpc-federation
	bffService, err := pb.NewBFFService(pb.BFFServiceConfig{
		Client: factory,
	})
	if err != nil {
		log.Fatalf("Failed to create BFF service: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterBFFServiceServer(grpcServer, bffService)
	reflection.Register(grpcServer)

	log.Printf("BFF Server listening on %s", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
