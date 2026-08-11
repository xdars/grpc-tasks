package grpcserver

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/grpc/auth"
	"github.com/xdars/grpc-tasks/internal/grpc/interceptors"
	"github.com/xdars/grpc-tasks/internal/grpc/payments"
	"github.com/xdars/grpc-tasks/internal/grpc/tasks"
)

type Server struct {
	srv *grpc.Server
}

func New(authSvc *auth.AuthService, taskSvc *tasks.TaskService, paymentSvc *payments.PaymentService) *Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor,
			interceptors.UnaryAuthInterceptor,
			interceptors.RecoveryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptors.StreamLoggingInterceptor,
			interceptors.StreamAuthInterceptor,
		),
	)
	pb.RegisterAuthServiceServer(srv, authSvc)
	pb.RegisterTaskServiceServer(srv, taskSvc)
	pb.RegisterPaymentServiceServer(srv, paymentSvc)
	reflection.Register(srv)
	return &Server{srv: srv}
}

func (s *Server) Listen(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf(":%d", port))
}

func (s *Server) Serve(lis net.Listener) error {
	return s.srv.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.srv.GracefulStop()
}
