package interceptor

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const claimsKey contextKey = "user_claims"

type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func Auth(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			err = status.Error(codes.Unauthenticated, "missing request metadata")
			return
		}

		authValues := md.Get("authorization")

		if len(authValues) == 0 {
			err = status.Error(codes.Unauthenticated, "missing authorization token")
		}

		tokenStr := authValues[0]

		const bearerPrefix = "Bearer "

		if len(tokenStr) <= len(bearerPrefix) || tokenStr[:len(bearerPrefix)] != bearerPrefix {
			err = status.Error(codes.Unauthenticated, "invalid authorization format: expected 'Bearer <token>'")
			return
		}

		tokenRaw := tokenStr[len(bearerPrefix):]

		claims := &UserClaims{}

		token, err := jwt.ParseWithClaims(tokenRaw, claims, func(t *jwt.Token) (any, error) {
			_, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, status.Error(
					codes.PermissionDenied,
					fmt.Sprintf("expected signing algorithm %v", t.Header["alg"]),
				)
			}

			return []byte(secret), nil
		})

		if !token.Valid || err != nil {
			err = status.Error(codes.PermissionDenied, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, claimsKey, claims)

		return handler(ctx, req)
	}
}

func GetUserClaims(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(claimsKey).(*UserClaims)
	return claims, ok
}
