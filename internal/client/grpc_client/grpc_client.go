// Package grpcclient предоставляет клиент для взаимодействия с gRPC сервером GophKeeper.
package grpcclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/Ko4etov/gophkeeper/internal/client/config"
	"github.com/Ko4etov/gophkeeper/internal/proto/auth"
	"github.com/Ko4etov/gophkeeper/internal/proto/sync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// BuildInfo содержит информацию о сборке приложения.
type BuildInfo struct {
	Version   string // версия приложения
	BuildDate string // дата сборки
}

// GrpcClient основной клиент для взаимодействия с gRPC сервером.
type GrpcClient struct {
	conn       *grpc.ClientConn
	config     *config.ClientConfig
	buildInfo  *BuildInfo
	ctx        context.Context
	authClient auth.AuthServiceClient
	syncClient sync.SyncServiceClient
}

// New создает нового gRPC клиента.
func New(ctx context.Context, cfg *config.ClientConfig, buildInfo *BuildInfo) (*GrpcClient, error) {
	conn, err := createGRPCConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	if buildInfo == nil {
		buildInfo = &BuildInfo{
			Version:   "dev",
			BuildDate: "unknown",
		}
	}

	return &GrpcClient{
		conn:       conn,
		authClient: auth.NewAuthServiceClient(conn),
		syncClient: sync.NewSyncServiceClient(conn),
		config:     cfg,
		buildInfo:  buildInfo,
		ctx:        ctx,
	}, nil
}

// Close закрывает gRPC соединение.
func (c *GrpcClient) Close() error {
	return c.conn.Close()
}

// GetBuildInfo возвращает информацию о сборке.
func (c *GrpcClient) GetBuildInfo() *BuildInfo {
	return c.buildInfo
}

// Version возвращает строку с версией.
func (c *GrpcClient) Version() string {
	return c.buildInfo.Version
}

// createGRPCConnection создает gRPC соединение с поддержкой TLS.
func createGRPCConnection(cfg *config.ClientConfig) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if cfg.InsecureTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig, err := createTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	return grpc.NewClient(cfg.ServerAddress, opts...)
}

// createTLSConfig создает конфигурацию TLS из сертификатов.
func createTLSConfig(cfg *config.ClientConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.CACertPath != "" {
		caCert, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = certPool
	}

	if cfg.ClientCertPath != "" && cfg.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}