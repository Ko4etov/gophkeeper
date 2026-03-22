// internal/server/interceptors/auth.go
package interceptors

import (
	"context"
	"strings"

	"github.com/Ko4etov/gophkeeper/internal/server/auth"
	"github.com/Ko4etov/gophkeeper/internal/server/service/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Публичные методы, не требующие аутентификации
var publicMethods = map[string]bool{
	"/auth.AuthService/Register": true,
	"/auth.AuthService/Login":    true,
}

// NewAuthInterceptor создает интерцептор для unary методов
func NewAuthInterceptor(jwtManager *auth.JWTManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		logger.Logger.Infof("Unary interceptor: %s", info.FullMethod)
		
		if !isProtectedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Извлекаем и валидируем токен
		ctx, err := validateAndAddUserContext(ctx, jwtManager)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// NewAuthStreamInterceptor создает интерцептор для stream методов
func NewAuthStreamInterceptor(jwtManager *auth.JWTManager) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		logger.Logger.Infof("Stream interceptor: %s", info.FullMethod)
		
		if !isProtectedMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		// Извлекаем и валидируем токен из контекста стрима
		ctx, err := validateAndAddUserContext(ss.Context(), jwtManager)
		if err != nil {
			return err
		}

		// Создаем wrapper для стрима с новым контекстом
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// validateAndAddUserContext валидирует токен и добавляет user context
func validateAndAddUserContext(ctx context.Context, jwtManager *auth.JWTManager) (context.Context, error) {
	// Извлекаем токен
	token, err := extractToken(ctx)
	if err != nil {
		return nil, err
	}

	// Валидируем токен
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Добавляем claims в контекст
	ctx = auth.WithUserContext(ctx, claims)
	logger.Logger.Infof("Authenticated user: %s", claims.UserID)

	return ctx, nil
}

// isProtectedMethod проверяет, требует ли метод аутентификации
func isProtectedMethod(method string) bool {
	return !publicMethods[method]
}

// extractToken извлекает токен из metadata
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	return parts[1], nil
}

// wrappedServerStream оборачивает ServerStream, позволяя заменить контекст
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context возвращает новый контекст с пользовательскими данными
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}