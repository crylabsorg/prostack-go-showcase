package project

import (
	"net/http"

	"github.com/crylabsorg/prostack-go/ps_error"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidName = ps_error.PSError{
		ErrorCode: ps_error.ErrorCodeValidationFailed,
		Message:   "project name is required",
		HStatus:   http.StatusBadRequest,
		GStatus:   codes.InvalidArgument,
	}
	ErrNotFound = ps_error.PSError{
		ErrorCode: ps_error.ErrorCodeGeneral,
		Message:   "project not found",
		HStatus:   http.StatusNotFound,
		GStatus:   codes.NotFound,
	}
)
