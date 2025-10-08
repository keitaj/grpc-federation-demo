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

type userServer struct {
	pb.UnimplementedUserServiceServer
	users           map[string]*pb.User
	maintenanceMode bool
}


func NewUserServer() *userServer {
	return &userServer{
		maintenanceMode: false, // Set to true to simulate maintenance
		users: map[string]*pb.User{
			"user-001": {
				Id:        "user-001",
				Name:      "Alice Johnson",
				Email:     "alice@example.com",
				Age:       28,
				CreatedAt: time.Now().Format(time.RFC3339),
			},
			"user-002": {
				Id:        "user-002",
				Name:      "Bob Smith",
				Email:     "bob@example.com",
				Age:       35,
				CreatedAt: time.Now().Format(time.RFC3339),
			},
			"user-003": {
				Id:        "user-003",
				Name:      "Charlie Brown",
				Email:     "charlie@example.com",
				Age:       42,
				CreatedAt: time.Now().Format(time.RFC3339),
			},
			"user-999": {
				Id:        "user-999",
				Name:      "Suspended User",
				Email:     "suspended@suspended.example.com",
				Age:       30,
				CreatedAt: time.Now().Format(time.RFC3339),
			},
		},
	}
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log.Printf("GetUser called with user_id: %s", req.UserId)

	// Check if service is in maintenance mode
	if s.maintenanceMode {
		return nil, errorutil.UnavailableError(
			fmt.Errorf("user service is under maintenance"),
			pb.UserErrorReason_USER_ERROR_REASON_MAINTENANCE,
		)
	}

	user, exists := s.users[req.UserId]
	if !exists {
		return nil, errorutil.FailedPreconditionError(
			pb.UserFailureCode_USER_FAILURE_CODE_USER_NOT_FOUND,
			"UserService/GetUser",
			fmt.Sprintf("User not found: %s", req.UserId),
		)
	}

	// Check if account is suspended
	if user.Email != "" && len(user.Email) > len("@suspended.example.com") {
		if user.Email[len(user.Email)-len("@suspended.example.com"):] == "@suspended.example.com" {
			return nil, errorutil.FailedPreconditionError(
				pb.UserFailureCode_USER_FAILURE_CODE_USER_ACCOUNT_SUSPENDED,
				"UserService/GetUser",
				"This account has been suspended.",
			)
		}
	}

	return &pb.GetUserResponse{
		User: user,
	}, nil
}

func (s *userServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log.Printf("ListUsers called with limit: %d, offset: %d", req.Limit, req.Offset)

	var users []*pb.User
	for _, user := range s.users {
		users = append(users, user)
	}

	start := int(req.Offset)
	if start > len(users) {
		start = len(users)
	}

	end := start + int(req.Limit)
	if end > len(users) {
		end = len(users)
	}

	return &pb.ListUsersResponse{
		Users: users[start:end],
		Total: int32(len(users)),
	}, nil
}

func main() {
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, NewUserServer())
	reflection.Register(s)

	log.Printf("User service starting on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}