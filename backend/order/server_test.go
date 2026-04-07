package main

import (
	"context"
	"testing"

	pb "github.com/keitaj/grpc-federation-demo/proto/backend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetOrder_Success(t *testing.T) {
	s := NewOrderServer()
	resp, err := s.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: "order-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Order.Id != "order-001" {
		t.Errorf("expected order ID order-001, got %s", resp.Order.Id)
	}
	if resp.Order.UserId != "user-001" {
		t.Errorf("expected user ID user-001, got %s", resp.Order.UserId)
	}
	if len(resp.Order.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Order.Items))
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	s := NewOrderServer()
	_, err := s.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: "order-notfound"})
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

func TestGetOrder_Cancelled(t *testing.T) {
	s := NewOrderServer()
	_, err := s.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: "order-999"})
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

func TestGetOrder_MaintenanceMode(t *testing.T) {
	s := NewOrderServer()
	s.maintenanceMode = true
	_, err := s.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: "order-001"})
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

func TestListOrdersByUser(t *testing.T) {
	s := NewOrderServer()

	t.Run("user with orders", func(t *testing.T) {
		resp, err := s.ListOrdersByUser(context.Background(), &pb.ListOrdersByUserRequest{
			UserId: "user-001",
			Limit:  10,
			Offset: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total == 0 {
			t.Error("expected non-zero total for user-001")
		}
	})

	t.Run("user with no orders", func(t *testing.T) {
		resp, err := s.ListOrdersByUser(context.Background(), &pb.ListOrdersByUserRequest{
			UserId: "user-noorders",
			Limit:  10,
			Offset: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 0 {
			t.Errorf("expected total 0, got %d", resp.Total)
		}
		if len(resp.Orders) != 0 {
			t.Errorf("expected 0 orders, got %d", len(resp.Orders))
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		resp, err := s.ListOrdersByUser(context.Background(), &pb.ListOrdersByUserRequest{
			UserId: "user-001",
			Limit:  1,
			Offset: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Orders) != 1 {
			t.Errorf("expected 1 order, got %d", len(resp.Orders))
		}
	})
}

func TestCreateOrder(t *testing.T) {
	s := NewOrderServer()

	items := []*pb.OrderItem{
		{ProductId: "prod-001", Quantity: 2, Price: 1299.99},
		{ProductId: "prod-002", Quantity: 1, Price: 29.99},
	}

	resp, err := s.CreateOrder(context.Background(), &pb.CreateOrderRequest{
		UserId: "user-001",
		Items:  items,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Order.UserId != "user-001" {
		t.Errorf("expected user ID user-001, got %s", resp.Order.UserId)
	}
	if len(resp.Order.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Order.Items))
	}
	expectedTotal := 1299.99*2 + 29.99
	if resp.Order.TotalAmount != expectedTotal {
		t.Errorf("expected total %f, got %f", expectedTotal, resp.Order.TotalAmount)
	}
	if resp.Order.Status != pb.OrderStatus_ORDER_STATUS_PENDING {
		t.Errorf("expected PENDING status, got %v", resp.Order.Status)
	}

	// Verify order is retrievable
	getResp, err := s.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: resp.Order.Id})
	if err != nil {
		t.Fatalf("unexpected error getting created order: %v", err)
	}
	if getResp.Order.Id != resp.Order.Id {
		t.Errorf("expected order ID %s, got %s", resp.Order.Id, getResp.Order.Id)
	}
}
