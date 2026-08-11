package interceptors

import (
	"context"
	"testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecoveryInterceptor(t *testing.T) {
    panicHandler := func(ctx context.Context, req any) (any, error) {
        panic("something went wrong")
    }

    _, err := RecoveryInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{
        FullMethod: "/test",
    }, panicHandler)

    if err == nil {
        t.Fatal("expected error, got nil")
    }

    if status.Code(err) != codes.Internal {
        t.Fatalf("expected codes.Internal, got %v", status.Code(err))
    }
}