package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/keitaj/grpc-federation-demo/backend/errorutil"
	pb "github.com/keitaj/grpc-federation-demo/proto/backend"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type productServer struct {
	pb.UnimplementedProductServiceServer
	products        map[string]*pb.Product
	maintenanceMode bool
}

func NewProductServer() *productServer {
	return &productServer{
		maintenanceMode: false, // Set to true to simulate maintenance
		products: map[string]*pb.Product{
			"prod-001": {
				Id:          "prod-001",
				Name:        "Laptop Pro 15",
				Description: "High-performance laptop with 15-inch display",
				Price:       1299.99,
				Category:    "Electronics",
				Stock:       50,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
			"prod-002": {
				Id:          "prod-002",
				Name:        "Wireless Mouse",
				Description: "Ergonomic wireless mouse with long battery life",
				Price:       29.99,
				Category:    "Accessories",
				Stock:       200,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
			"prod-003": {
				Id:          "prod-003",
				Name:        "USB-C Hub",
				Description: "7-in-1 USB-C hub with HDMI and card reader",
				Price:       49.99,
				Category:    "Accessories",
				Stock:       150,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
			"prod-004": {
				Id:          "prod-004",
				Name:        "Mechanical Keyboard",
				Description: "RGB mechanical keyboard with blue switches",
				Price:       89.99,
				Category:    "Accessories",
				Stock:       75,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
			"prod-005": {
				Id:          "prod-005",
				Name:        "4K Monitor",
				Description: "27-inch 4K IPS monitor with HDR support",
				Price:       399.99,
				Category:    "Electronics",
				Stock:       30,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
			"prod-999": {
				Id:          "prod-999",
				Name:        "Out of Stock Item",
				Description: "This product is out of stock",
				Price:       99.99,
				Category:    "Test",
				Stock:       0, // Out of stock
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
		},
	}
}

func (s *productServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	log.Printf("GetProduct called with product_id: %s", req.ProductId)

	// Check if service is in maintenance mode
	if s.maintenanceMode {
		return nil, errorutil.UnavailableError(
			fmt.Errorf("product service is under maintenance"),
			pb.ProductErrorReason_PRODUCT_ERROR_REASON_MAINTENANCE,
		)
	}

	product, exists := s.products[req.ProductId]
	if !exists {
		return nil, errorutil.FailedPreconditionError(
			pb.ProductFailureCode_PRODUCT_FAILURE_CODE_PRODUCT_NOT_FOUND,
			"ProductService/GetProduct",
			fmt.Sprintf("Product not found: %s", req.ProductId),
		)
	}

	// Check if product is out of stock
	if product.Stock == 0 {
		return nil, errorutil.FailedPreconditionError(
			pb.ProductFailureCode_PRODUCT_FAILURE_CODE_PRODUCT_OUT_OF_STOCK,
			"ProductService/GetProduct",
			fmt.Sprintf("Product is out of stock: %s", req.ProductId),
		)
	}

	return &pb.GetProductResponse{
		Product: product,
	}, nil
}

func (s *productServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	log.Printf("ListProducts called with category: %s, limit: %d, offset: %d", req.Category, req.Limit, req.Offset)

	var products []*pb.Product
	for _, product := range s.products {
		if req.Category == "" || product.Category == req.Category {
			products = append(products, product)
		}
	}

	start := int(req.Offset)
	if start > len(products) {
		start = len(products)
	}

	end := start + int(req.Limit)
	if end > len(products) {
		end = len(products)
	}

	return &pb.ListProductsResponse{
		Products: products[start:end],
		Total:    int32(len(products)),
	}, nil
}

func (s *productServer) GetProductsByIDs(ctx context.Context, req *pb.GetProductsByIDsRequest) (*pb.GetProductsByIDsResponse, error) {
	log.Printf("GetProductsByIDs called with product_ids: %v", req.ProductIds)

	var products []*pb.Product
	for _, id := range req.ProductIds {
		if product, exists := s.products[id]; exists {
			products = append(products, product)
		}
	}

	return &pb.GetProductsByIDsResponse{
		Products: products,
	}, nil
}

func main() {
	port := ":50052"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterProductServiceServer(s, NewProductServer())
	reflection.Register(s)

	log.Printf("Product service starting on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}