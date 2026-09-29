package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/kidrury/grpc-blog/gen/blog"
	"github.com/kidrury/grpc-blog/internal/database"
	"github.com/kidrury/grpc-blog/internal/repository/postgres"
	"github.com/kidrury/grpc-blog/internal/service"
	"github.com/kidrury/grpc-blog/server/handler"
	"github.com/kidrury/grpc-blog/server/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

func main() {
	var logHandler slog.Handler

	env := getEnv("ENV", "development")

	if env == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	slog.SetDefault(slog.New(logHandler))

	addr := getEnv("GRPC_ADDR", ":50051")
	jwtSecret := getEnv("JWT_SECRET", "fallback_jwt_secret")
	tlsEnabled := getEnv("TLS_ENABLED", "false") == "true"
	certFile := getEnv("TLS_CERT", "server.crt")
	keyFile := getEnv("TLS_KEY", "server.key")

	pool, err := database.NewPostgresPool(context.Background(), getEnv("DATABASE_URL", "error"))
	if err != nil {
		slog.Error("failed creating database pool", "error", err)
		os.Exit(1)
	}

	postRepo := postgres.NewPostRepository(pool)

	postService := service.NewPostService(postRepo)

	postHandler := handler.NewPostHandler(postService)

	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.Recovery(),
			interceptor.Log(),
			interceptor.Auth(jwtSecret),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:              15 * time.Second,
			Timeout:           5 * time.Second,
			MaxConnectionIdle: 30 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.MaxRecvMsgSize(4 * 1024 * 1024),
		grpc.MaxSendMsgSize(4 * 1024 * 1024),
	}

	if tlsEnabled {
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			slog.Error("failed to load TLS credentials", "error", err)
			os.Exit(1)
		}

		serverOptions = append(serverOptions, grpc.Creds(creds))
		slog.Info("TLS enabled", "cert", certFile)
	} else {
		slog.Warn("TLS disabled - use TLS_ENABLED=true in production", "cert", certFile)
	}

	grpcServer := grpc.NewServer(serverOptions...)

	pb.RegisterBlogServiceServer(grpcServer, postHandler)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("blog.BlogService", healthpb.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to create lsitener", "error", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan

		healthServer.SetServingStatus("blog.BlogService", healthpb.HealthCheckResponse_NOT_SERVING)

		grpcServer.GracefulStop()
		slog.Info("grpc server stopped")
	}()

	slog.Info("gRPC server listening", "addr", addr, "tls", tlsEnabled)
	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
