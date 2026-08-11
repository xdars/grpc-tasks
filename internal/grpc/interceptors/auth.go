package interceptors

import (
	"context"
	"strings"

	jwt "github.com/xdars/grpc-tasks/internal/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/taskservice.AuthService/Login":                                 true,
	"/taskservice.AuthService/Register":                              true,
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,
}

type contextKey string

const UserClaimsKey contextKey = "user_claims"

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }

func StreamAuthInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if publicMethods[info.FullMethod] {
		return handler(srv, ss)
	}
	ctx, err := authenticate(ss.Context())
	if err != nil {
		return err
	}
	return handler(srv, &wrappedStream{ss, ctx})
}

func authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	claims, err := jwt.Validate(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return context.WithValue(ctx, UserClaimsKey, claims), nil
}

func UnaryAuthInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if publicMethods[info.FullMethod] {
		return handler(ctx, req)
	}
	ctx, err := authenticate(ctx)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}
