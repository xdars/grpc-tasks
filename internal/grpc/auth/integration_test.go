package auth_test

import (
	"context"
	"testing"

	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/db"
	"github.com/xdars/grpc-tasks/internal/grpc/auth"
	"github.com/xdars/grpc-tasks/internal/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIntegration_Register(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := db.NewUserRepository(pool)
	svc := auth.NewAuthService(userRepo)

	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	})

	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if resp.UserId == "" {
		t.Fatal("expected user id")
	}
}

func TestIntegration_Login_Success(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := db.NewUserRepository(pool)
	svc := auth.NewAuthService(userRepo)

	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	})

	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if resp.UserId == "" {
		t.Fatal("expected user id")
	}

	resp2, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: "alice",
		Password: "secret123",
	})

	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if resp2.Token == "" {
		t.Fatal("expected Token")
	}
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := db.NewUserRepository(pool)
	svc := auth.NewAuthService(userRepo)

	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	})

	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if resp.UserId == "" {
		t.Fatal("expected user id")
	}

	_, err = svc.Login(context.Background(), &pb.LoginRequest{
		Username: "alice",
		Password: "secret124",
	})

	code := status.Code(err)
	if code != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", code)
	}
}
