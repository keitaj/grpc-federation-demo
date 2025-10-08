package errorutil

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FailureCodeStringer is an interface for enums that can be converted to string
type FailureCodeStringer interface {
	String() string
}

// FailedPreconditionError creates a FAILED_PRECONDITION error with PreconditionFailure details
func FailedPreconditionError(failureCode FailureCodeStringer, subject, description string) error {
	st := status.New(codes.FailedPrecondition, "")

	v := &errdetails.PreconditionFailure_Violation{
		Type:        failureCode.String(),
		Subject:     subject,
		Description: description,
	}

	pfailure := &errdetails.PreconditionFailure{
		Violations: []*errdetails.PreconditionFailure_Violation{v},
	}

	st, err := st.WithDetails(pfailure)
	if err != nil {
		// Fallback to simple error if WithDetails fails
		return status.Error(codes.FailedPrecondition, description)
	}

	return st.Err()
}

// ErrorReasonStringer is an interface for error reason enums that can be converted to string
type ErrorReasonStringer interface {
	String() string
}

// UnavailableError creates an UNAVAILABLE error with ErrorInfo details
func UnavailableError(err error, reason ErrorReasonStringer) error {
	in := &errdetails.ErrorInfo{
		Reason: reason.String(),
	}

	st := status.New(codes.Unavailable, err.Error())

	st, withErr := st.WithDetails(in)
	if withErr != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to set details: %s", withErr))
	}

	return st.Err()
}
