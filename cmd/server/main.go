// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	authproto "github.com/Ko4etov/gophkeeper/internal/proto/auth"
	syncproto "github.com/Ko4etov/gophkeeper/internal/proto/sync"
	"github.com/Ko4etov/gophkeeper/internal/server/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/config"
	"github.com/Ko4etov/gophkeeper/internal/server/interceptors"
	authservice "github.com/Ko4etov/gophkeeper/internal/server/service/auth"
	syncservice "github.com/Ko4etov/gophkeeper/internal/server/service/sync"
	"github.com/Ko4etov/gophkeeper/internal/server/storage"
	"google.golang.org/grpc"
)

func main() {
	mainCtx := context.Background()

    config, err := config.New()
    if err != nil {
        log.Fatal(err)
    }
    
    storage, err := storage.NewStorage(config.ConnectionPool, mainCtx)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer storage.Close()

    jwtManager := auth.NewJWTManager(&auth.JWTConfig{
		SecretKey: config.JWTSecret,
		AccessTTL: config.AccessTokenTTL,
		RefreshTTL: config.RefreshTokenTTL,
	})

    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(interceptors.NewAuthInterceptor(jwtManager)),
        grpc.StreamInterceptor(interceptors.NewAuthStreamInterceptor(jwtManager)),
    )

    authService := authservice.NewService(storage, jwtManager)
    syncService := syncservice.NewSyncService(storage)
    authproto.RegisterAuthServiceServer(grpcServer, authService)
    syncproto.RegisterSyncServiceServer(grpcServer, syncService)

    lis, err := net.Listen("tcp", config.GRPCAddress)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

	notifyCtx, stop := signal.NotifyContext(mainCtx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
    defer stop()

    errChan := make(chan error, 1)

    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            errChan <- fmt.Errorf("gRPC server error: %w", err)
        }
    }()

    select {
	case err := <-errChan:
		log.Fatalf("%v", err)
	case <-notifyCtx.Done():
        
	}

	grpcServer.GracefulStop()
}