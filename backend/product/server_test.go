package main

import (
	"context"
	"testing"

	pb "github.com/keitaj/grpc-federation-demo/proto/backend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetProduct_Success(t *testing.T) {
	s := NewProductServer()
	resp, err := s.GetProduct(context.Background(), &pb.GetProductRequest{ProductId: "prod-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Product.Id != "prod-001" {
		t.Errorf("expected product ID prod-001, got %s", resp.Product.Id)
	}
	if resp.Product.Name != "Laptop Pro 15" {
		t.Errorf("expected name Laptop Pro 15, got %s", resp.Product.Name)
	}
	if resp.Product.Price != 1299.99 {
		t.Errorf("expected price 1299.99, got %f", resp.Product.Price)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	s := NewProductServer()
	_, err := s.GetProduct(context.Background(), &pb.GetProductRequest{ProductId: "prod-notfound"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", st.Code())
	}
}

func TestGetProduct_OutOfStock(t *testing.T) {
	s := NewProductServer()
	_, err := s.GetProduct(context.Background(), &pb.GetProductRequest{ProductId: "prod-999"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", st.Code())
	}
}

func TestGetProduct_MaintenanceMode(t *testing.T) {
	s := NewProductServer()
	s.maintenanceMode = true
	_, err := s.GetProduct(context.Background(), &pb.GetProductRequest{ProductId: "prod-001"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC status error")
	}
	if st.Code() != codes.Unavailable {
		t.Errorf("expected Unavailable, got %v", st.Code())
	}
}

func TestListProducts(t *testing.T) {
	s := NewProductServer()

	t.Run("all products", func(t *testing.T) {
		resp, err := s.ListProducts(context.Background(), &pb.ListProductsRequest{Limit: 20, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 6 {
			t.Errorf("expected total 6, got %d", resp.Total)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		resp, err := s.ListProducts(context.Background(), &pb.ListProductsRequest{Category: "Electronics", Limit: 20, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, p := range resp.Products {
			if p.Category != "Electronics" {
				t.Errorf("expected category Electronics, got %s", p.Category)
			}
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		resp, err := s.ListProducts(context.Background(), &pb.ListProductsRequest{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Products) != 2 {
			t.Errorf("expected 2 products, got %d", len(resp.Products))
		}
	})
}

func TestGetProductsByIDs(t *testing.T) {
	s := NewProductServer()

	t.Run("existing products", func(t *testing.T) {
		resp, err := s.GetProductsByIDs(context.Background(), &pb.GetProductsByIDsRequest{
			ProductIds: []string{"prod-001", "prod-002"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Products) != 2 {
			t.Errorf("expected 2 products, got %d", len(resp.Products))
		}
	})

	t.Run("mixed existing and non-existing", func(t *testing.T) {
		resp, err := s.GetProductsByIDs(context.Background(), &pb.GetProductsByIDsRequest{
			ProductIds: []string{"prod-001", "prod-notfound"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Products) != 1 {
			t.Errorf("expected 1 product, got %d", len(resp.Products))
		}
	})

	t.Run("empty IDs", func(t *testing.T) {
		resp, err := s.GetProductsByIDs(context.Background(), &pb.GetProductsByIDsRequest{
			ProductIds: []string{},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Products) != 0 {
			t.Errorf("expected 0 products, got %d", len(resp.Products))
		}
	})
}
