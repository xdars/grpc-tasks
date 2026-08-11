package interceptors

import (
	"context"
	"testing"

	jwtpkg "github.com/xdars/grpc-tasks/internal/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptor_NoToken(t *testing.T) {
	ctx := context.Background()

	_, err := UnaryAuthInterceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/taskservice.TaskService/CreateTask",
	}, fakeHandler)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthInterceptor_ValidToken(t *testing.T) {
	token, err := jwtpkg.Generate("alice", "123")
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	t.Log("token:", token)
	var claims *jwtpkg.Claims
	claims, err = jwtpkg.Validate(token)
	t.Log("claims:", claims, "err:", err)

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer "+token),
	)

	_, err = UnaryAuthInterceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/taskservice.TaskService/CreateTask",
	}, fakeHandler)

	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
}

func TestAuthInterceptor_PublicMethod(t *testing.T) {
	ctx := context.Background()

	_, err := UnaryAuthInterceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/taskservice.AuthService/Login",
	}, fakeHandler)

	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
}

func fakeHandler(ctx context.Context, req any) (any, error) {
	return nil, nil
}
