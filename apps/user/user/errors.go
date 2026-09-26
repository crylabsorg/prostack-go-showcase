package user

import (
	"net/http"

	"github.com/crylabsorg/prostack-go/ps_error"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidInput = ps_error.PSError{
		ErrorCode: ps_error.ErrorCodeValidationFailed,
		Message:   "invalid user input",
		HStatus:   http.StatusBadRequest,
		GStatus:   codes.InvalidArgument,
	}
	ErrNotFound = ps_error.PSError{
		ErrorCode: ps_error.ErrorCodeGeneral,
		Message:   "user not found",
		HStatus:   http.StatusNotFound,
		GStatus:   codes.NotFound,
	}
	ErrEmailExists = ps_error.PSError{
		ErrorCode: ps_error.ErrorCodeGeneral,
		Message:   "email already exists",
		HStatus:   http.StatusConflict,
		GStatus:   codes.AlreadyExists,
	}
)
