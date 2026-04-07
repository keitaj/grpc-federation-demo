package errorutil

import (
	"fmt"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testFailureCode string

func (c testFailureCode) String() string { return string(c) }

type testErrorReason string

func (r testErrorReason) String() string { return string(r) }

func TestFailedPreconditionError(t *testing.T) {
	tests := []struct {
		name        string
		failureCode testFailureCode
		subject     string
		description string
	}{
		{
			name:        "user not found",
			failureCode: "USER_NOT_FOUND",
			subject:     "UserService/GetUser",
			description: "User not found: user-999",
		},
		{
			name:        "order cancelled",
			failureCode: "ORDER_CANCELLED",
			subject:     "OrderService/GetOrder",
			description: "This order is cancelled.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FailedPreconditionError(tt.failureCode, tt.subject, tt.description)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("expected gRPC status error")
			}

			if st.Code() != codes.FailedPrecondition {
				t.Errorf("expected code FailedPrecondition, got %v", st.Code())
			}

			details := st.Details()
			if len(details) != 1 {
				t.Fatalf("expected 1 detail, got %d", len(details))
			}

			pf, ok := details[0].(*errdetails.PreconditionFailure)
			if !ok {
				t.Fatal("expected PreconditionFailure detail")
			}

			if len(pf.Violations) != 1 {
				t.Fatalf("expected 1 violation, got %d", len(pf.Violations))
			}

			v := pf.Violations[0]
			if v.Type != tt.failureCode.String() {
				t.Errorf("expected violation type %q, got %q", tt.failureCode, v.Type)
			}
			if v.Subject != tt.subject {
				t.Errorf("expected subject %q, got %q", tt.subject, v.Subject)
			}
			if v.Description != tt.description {
				t.Errorf("expected description %q, got %q", tt.description, v.Description)
			}
		})
	}
}

func TestUnavailableError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		reason testErrorReason
	}{
		{
			name:   "maintenance mode",
			err:    fmt.Errorf("service is under maintenance"),
			reason: "MAINTENANCE",
		},
		{
			name:   "custom reason",
			err:    fmt.Errorf("temporarily unavailable"),
			reason: "TEMPORARILY_UNAVAILABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UnavailableError(tt.err, tt.reason)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("expected gRPC status error")
			}

			if st.Code() != codes.Unavailable {
				t.Errorf("expected code Unavailable, got %v", st.Code())
			}

			if st.Message() != tt.err.Error() {
				t.Errorf("expected message %q, got %q", tt.err.Error(), st.Message())
			}

			details := st.Details()
			if len(details) != 1 {
				t.Fatalf("expected 1 detail, got %d", len(details))
			}

			ei, ok := details[0].(*errdetails.ErrorInfo)
			if !ok {
				t.Fatal("expected ErrorInfo detail")
			}

			if ei.Reason != tt.reason.String() {
				t.Errorf("expected reason %q, got %q", tt.reason, ei.Reason)
			}
		})
	}
}
