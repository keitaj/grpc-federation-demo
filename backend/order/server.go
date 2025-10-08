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

type orderServer struct {
	pb.UnimplementedOrderServiceServer
	orders          map[string]*pb.Order
	userOrders      map[string][]*pb.Order
	orderCounter    int
	maintenanceMode bool
}

func NewOrderServer() *orderServer {
	s := &orderServer{
		orders:          make(map[string]*pb.Order),
		userOrders:      make(map[string][]*pb.Order),
		orderCounter:    1,
		maintenanceMode: false, // Set to true to simulate maintenance
	}

	s.seedOrders()
	return s
}

func (s *orderServer) seedOrders() {
	orders := []*pb.Order{
		{
			Id:     "order-001",
			UserId: "user-001",
			Items: []*pb.OrderItem{
				{ProductId: "prod-001", Quantity: 1, Price: 1299.99},
				{ProductId: "prod-002", Quantity: 2, Price: 29.99},
			},
			TotalAmount: 1359.97,
			Status:      pb.OrderStatus_ORDER_STATUS_DELIVERED,
			CreatedAt:   time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:     "order-002",
			UserId: "user-001",
			Items: []*pb.OrderItem{
				{ProductId: "prod-003", Quantity: 1, Price: 49.99},
				{ProductId: "prod-004", Quantity: 1, Price: 89.99},
			},
			TotalAmount: 139.98,
			Status:      pb.OrderStatus_ORDER_STATUS_PROCESSING,
			CreatedAt:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:     "order-003",
			UserId: "user-002",
			Items: []*pb.OrderItem{
				{ProductId: "prod-005", Quantity: 2, Price: 399.99},
			},
			TotalAmount: 799.98,
			Status:      pb.OrderStatus_ORDER_STATUS_SHIPPED,
			CreatedAt:   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:     "order-004",
			UserId: "user-002",
			Items: []*pb.OrderItem{
				{ProductId: "prod-002", Quantity: 3, Price: 29.99},
			},
			TotalAmount: 89.97,
			Status:      pb.OrderStatus_ORDER_STATUS_PENDING,
			CreatedAt:   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:     "order-005",
			UserId: "user-003",
			Items: []*pb.OrderItem{
				{ProductId: "prod-001", Quantity: 1, Price: 1299.99},
				{ProductId: "prod-005", Quantity: 1, Price: 399.99},
			},
			TotalAmount: 1699.98,
			Status:      pb.OrderStatus_ORDER_STATUS_DELIVERED,
			CreatedAt:   time.Now().Add(-120 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-96 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:          "order-999",
			UserId:      "user-001",
			Items:       []*pb.OrderItem{}, // Empty items to trigger ORDER_NO_ITEMS
			TotalAmount: 0,
			Status:      pb.OrderStatus_ORDER_STATUS_CANCELLED,
			CreatedAt:   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:     "order-998",
			UserId: "user-001",
			Items: []*pb.OrderItem{
				{ProductId: "prod-notfound", Quantity: 1, Price: 99.99}, // Non-existent product
			},
			TotalAmount: 99.99,
			Status:      pb.OrderStatus_ORDER_STATUS_PENDING,
			CreatedAt:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			Id:     "order-997",
			UserId: "user-001",
			Items: []*pb.OrderItem{
				{ProductId: "prod-999", Quantity: 1, Price: 99.99}, // Out of stock product
			},
			TotalAmount: 99.99,
			Status:      pb.OrderStatus_ORDER_STATUS_PENDING,
			CreatedAt:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			UpdatedAt:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}

	for _, order := range orders {
		s.orders[order.Id] = order
		s.userOrders[order.UserId] = append(s.userOrders[order.UserId], order)
	}
	s.orderCounter = 8
}

func (s *orderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	log.Printf("GetOrder called with order_id: %s", req.OrderId)

	// Check if service is in maintenance mode
	if s.maintenanceMode {
		return nil, errorutil.UnavailableError(
			fmt.Errorf("order service is under maintenance"),
			pb.OrderErrorReason_ORDER_ERROR_REASON_MAINTENANCE,
		)
	}

	order, exists := s.orders[req.OrderId]
	if !exists {
		return nil, errorutil.FailedPreconditionError(
			pb.OrderFailureCode_ORDER_FAILURE_CODE_ORDER_NOT_FOUND,
			"OrderService/GetOrder",
			fmt.Sprintf("Order not found: %s", req.OrderId),
		)
	}

	// Check if order is cancelled
	if order.Status == pb.OrderStatus_ORDER_STATUS_CANCELLED {
		return nil, errorutil.FailedPreconditionError(
			pb.OrderFailureCode_ORDER_FAILURE_CODE_ORDER_CANCELLED,
			"OrderService/GetOrder",
			"This order is cancelled and cannot be viewed.",
		)
	}

	// Check if order has no items
	if len(order.Items) == 0 {
		return nil, errorutil.FailedPreconditionError(
			pb.OrderFailureCode_ORDER_FAILURE_CODE_ORDER_NO_ITEMS,
			"OrderService/GetOrder",
			"This order has no items.",
		)
	}

	return &pb.GetOrderResponse{
		Order: order,
	}, nil
}

func (s *orderServer) ListOrdersByUser(ctx context.Context, req *pb.ListOrdersByUserRequest) (*pb.ListOrdersByUserResponse, error) {
	log.Printf("ListOrdersByUser called with user_id: %s, limit: %d, offset: %d", req.UserId, req.Limit, req.Offset)

	orders := s.userOrders[req.UserId]
	if orders == nil {
		orders = []*pb.Order{}
	}

	start := int(req.Offset)
	if start > len(orders) {
		start = len(orders)
	}

	end := start + int(req.Limit)
	if end > len(orders) {
		end = len(orders)
	}

	return &pb.ListOrdersByUserResponse{
		Orders: orders[start:end],
		Total:  int32(len(orders)),
	}, nil
}

func (s *orderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log.Printf("CreateOrder called with user_id: %s, items: %v", req.UserId, req.Items)

	orderId := fmt.Sprintf("order-%03d", s.orderCounter)
	s.orderCounter++

	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	order := &pb.Order{
		Id:          orderId,
		UserId:      req.UserId,
		Items:       req.Items,
		TotalAmount: totalAmount,
		Status:      pb.OrderStatus_ORDER_STATUS_PENDING,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	s.orders[orderId] = order
	s.userOrders[req.UserId] = append(s.userOrders[req.UserId], order)

	return &pb.CreateOrderResponse{
		Order: order,
	}, nil
}

func main() {
	port := ":50053"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, NewOrderServer())
	reflection.Register(s)

	log.Printf("Order service starting on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}