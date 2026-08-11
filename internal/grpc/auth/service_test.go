package auth

import (
	"context"
	"log"
	"testing"

	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/db"
)

func setup() *AuthService {
	pool, err := db.New("postgres://localhost/grpc_tasks?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	userRepo := db.NewUserRepository(pool)
	return &AuthService{users: userRepo}
}

func TestRegister_Success(t *testing.T) {
	svc := setup()
	resp, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	})
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if resp.UserId == "" {
		t.Fatal("expected user id, got empty string")
	}
}

func TestRegister_DuplicateUser(t *testing.T) {
	svc := setup()
	req := &pb.RegisterRequest{Username: "alice", Password: "secret123"}
	_, err := svc.Register(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for duplicate user, got nil")
	}
}

func TestLogin_Success(t *testing.T) {
	svc := setup()
	_, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice", Password: "secret123",
	})
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	resp, err := svc.Login(context.Background(), &pb.LoginRequest{
		Username: "alice", Password: "secret123",
	})
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token, got empty string")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := setup()
	_, err := svc.Register(context.Background(), &pb.RegisterRequest{
		Username: "alice", Password: "secret123",
	})
	if err != nil {
		t.Fatal("expected no error, got:", err)
	}
	_, err = svc.Login(context.Background(), &pb.LoginRequest{
		Username: "alice", Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
