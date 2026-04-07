package main

import (
	"context"
	"testing"

	pb "github.com/keitaj/grpc-federation-demo/proto/backend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUser_Success(t *testing.T) {
	s := NewUserServer()
	resp, err := s.GetUser(context.Background(), &pb.GetUserRequest{UserId: "user-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Id != "user-001" {
		t.Errorf("expected user ID user-001, got %s", resp.User.Id)
	}
	if resp.User.Name != "Alice Johnson" {
		t.Errorf("expected name Alice Johnson, got %s", resp.User.Name)
	}
	if resp.User.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", resp.User.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	s := NewUserServer()
	_, err := s.GetUser(context.Background(), &pb.GetUserRequest{UserId: "user-notfound"})
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

func TestGetUser_Suspended(t *testing.T) {
	s := NewUserServer()
	_, err := s.GetUser(context.Background(), &pb.GetUserRequest{UserId: "user-999"})
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

func TestGetUser_MaintenanceMode(t *testing.T) {
	s := NewUserServer()
	s.maintenanceMode = true
	_, err := s.GetUser(context.Background(), &pb.GetUserRequest{UserId: "user-001"})
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

func TestListUsers(t *testing.T) {
	s := NewUserServer()

	t.Run("all users", func(t *testing.T) {
		resp, err := s.ListUsers(context.Background(), &pb.ListUsersRequest{Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Total != 4 {
			t.Errorf("expected total 4, got %d", resp.Total)
		}
		if len(resp.Users) != 4 {
			t.Errorf("expected 4 users, got %d", len(resp.Users))
		}
	})

	t.Run("with limit", func(t *testing.T) {
		resp, err := s.ListUsers(context.Background(), &pb.ListUsersRequest{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Users) != 2 {
			t.Errorf("expected 2 users, got %d", len(resp.Users))
		}
		if resp.Total != 4 {
			t.Errorf("expected total 4, got %d", resp.Total)
		}
	})

	t.Run("offset beyond total", func(t *testing.T) {
		resp, err := s.ListUsers(context.Background(), &pb.ListUsersRequest{Limit: 10, Offset: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Users) != 0 {
			t.Errorf("expected 0 users, got %d", len(resp.Users))
		}
	})
}
