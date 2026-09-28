package handler

import (
	"errors"

	"github.com/kidrury/grpc-blog/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrPostNotFound):
		return status.Error(codes.NotFound, "post not found")
	case errors.Is(err, domain.ErrIDRequired):
		return status.Error(codes.InvalidArgument, "id required")
	case errors.Is(err, domain.ErrTitleRequired):
		return status.Error(codes.InvalidArgument, "title required")
	case errors.Is(err, domain.ErrContentRequired):
		return status.Error(codes.InvalidArgument, "content required")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
