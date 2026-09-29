package interceptor

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type RequestIDKey string

const requestIDKey RequestIDKey = "request_id_key"

func Log() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		now := time.Now()

		requestID := uuid.New().String()

		ctx = context.WithValue(ctx, requestIDKey, requestID)

		peerAddr := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			peerAddr = p.Addr.String()
		}

		slog.InfoContext(ctx, "grpc call started",
			"request_id", requestID,
			"method", info.FullMethod,
			"peer", peerAddr,
		)

		res, err := handler(ctx, req)

		duration := time.Since(now)

		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		userID := "unauthenticated"
		if claims, ok := GetUserClaims(ctx); ok {
			userID = claims.UserID
		}

		logAttr := []any{
			"request_id", requestID,
			"method", info.FullMethod,
			"code", code,
			"duration_ms", duration.Milliseconds(),
			"user_id", userID,
			"peer", peerAddr,
		}
		if err == nil {
			slog.InfoContext(ctx, "grpc call completed", logAttr...)
		} else {
			slog.ErrorContext(ctx, "grpc call failed", logAttr...)
		}

		return res, err

	}
}

func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}
