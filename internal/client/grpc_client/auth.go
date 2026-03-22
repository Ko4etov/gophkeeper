package grpcclient

import (
	"context"
	"time"

	"github.com/Ko4etov/gophkeeper/internal/proto/auth"
	"google.golang.org/grpc/metadata"
)

// Login выполняет вход пользователя на сервере.
func (c *GrpcClient) Login(email string, password string) (*auth.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	resp, err := c.authClient.Login(ctx, &auth.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Logout выполняет выход пользователя на сервере.
func (c *GrpcClient) Logout(token string) error {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	authCtx := c.GetAuthContext(ctx, token)

	_, err := c.authClient.Logout(authCtx, &auth.LogoutRequest{
		AccessToken: token,
	})
	if err != nil {
		return err
	}

	return nil
}

// Register создает нового пользователя на сервере.
func (c *GrpcClient) Register(email string, password string) error {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := c.authClient.Register(ctx, &auth.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return err
	}

	return nil
}

// RefreshToken обновляет access-токен с помощью refresh-токена.
func (c *GrpcClient) RefreshToken(accessToken string, refreshToken string) (*auth.RefreshTokenResponse, error) {
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	authCtx := c.GetAuthContext(ctx, accessToken)

	resp, err := c.authClient.RefreshToken(authCtx, &auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetAuthContext добавляет токен авторизации в контекст gRPC.
func (c *GrpcClient) GetAuthContext(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
