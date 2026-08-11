package auth

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/db"
	"github.com/xdars/grpc-tasks/internal/domain"
	jwt "github.com/xdars/grpc-tasks/internal/jwt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	users *db.UserRepository
}

func NewAuthService(users *db.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(req.Username) == 0 || len(req.Password) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty username or password")
	}
	exists, err := s.users.Exists(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "db error")
	}
	if exists {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	id := uuid.NewString()
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not hash password")
	}
	if err := s.users.Add(ctx, &domain.User{
		ID:           id,
		Username:     req.Username,
		PasswordHash: hash,
	}); err != nil {
		return nil, status.Error(codes.Internal, "could not create user")
	}

	return &pb.RegisterResponse{UserId: id}, nil
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, ok, err := s.users.Get(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "db error")
	}
	if !ok || !CheckPassword(req.Password, user.PasswordHash) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := jwt.Generate(req.Username, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}
