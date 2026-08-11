package payments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/db"
	"github.com/xdars/grpc-tasks/internal/domain"
	"github.com/xdars/grpc-tasks/internal/grpc/interceptors"
	"github.com/xdars/grpc-tasks/internal/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
	payments *db.PaymentRepository
}

func NewPaymentService(payments *db.PaymentRepository) *PaymentService {
	return &PaymentService{payments: payments}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.Payment, error) {
	id := uuid.NewString()
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}
	payment := &domain.Payment{
		ID:       id,
		UserID:   claims.UserID,
		Type:     domain.PaymentType(req.Type),
		Amount:   req.Amount,
		Currency: req.Currency,
		Status:   domain.PaymentStatusPending,
	}
	if err := s.payments.Add(ctx, payment); err != nil {
		return nil, status.Error(codes.Internal, "could not create payment")
	}
	return domainToProto(payment), nil
}

func (s *PaymentService) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.Payment, error) {
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}
	payment, ok, err := s.payments.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not get payment")
	}

	if !ok || payment.UserID != claims.UserID {
		return nil, status.Error(codes.NotFound, "payment not found")
	}
	return domainToProto(payment), nil
}

func (s *PaymentService) ProcessPayment(stream pb.PaymentService_ProcessPaymentServer) error {
	ctx := stream.Context()

	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return status.Error(codes.Unauthenticated, "no user in context")
	}
	_ = claims

	cmdCh := make(chan *pb.PaymentCommand)
	errCh := make(chan error, 1)

	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				errCh <- err
				return
			}
			cmdCh <- cmd
		}
	}()

	for {
		select {
		case cmd := <-cmdCh:
			switch cmd.Action {
			case pb.PaymentCommand_ACTION_START:
				payment, ok, err := s.payments.Get(ctx, cmd.PaymentId)
				if err != nil {
					return status.Error(codes.Internal, "could not get payment")
				}
				if !ok || payment.UserID != claims.UserID {
					return status.Error(codes.NotFound, "payment not found")
				}

				if err := s.payments.UpdateStatus(ctx, cmd.PaymentId, domain.PaymentStatusProcessing); err != nil {
					return status.Error(codes.Internal, "could not update payment")
				}

				stream.Send(&pb.PaymentEvent{
					PaymentId: cmd.PaymentId,
					Status:    pb.PaymentStatus_PAYMENT_STATUS_PROCESSING,
					Message:   "payment is being processed",
				})
				// emulate processing
				go func(paymentID string) {
					time.Sleep(2 * time.Second)

					p, ok, err := s.payments.Get(ctx, paymentID)
					if err != nil || !ok {
						return
					}
					if p.Status == domain.PaymentStatusFailed {
						return
					}

					s.payments.UpdateStatus(ctx, paymentID, domain.PaymentStatusSuccess)
					stream.Send(&pb.PaymentEvent{
						PaymentId: paymentID,
						Status:    pb.PaymentStatus_PAYMENT_STATUS_SUCCESS,
						Message:   "payment successful",
					})
				}(cmd.PaymentId)
			case pb.PaymentCommand_ACTION_CANCEL:
				payment, ok, err := s.payments.Get(ctx, cmd.PaymentId)
				if err != nil {
					return status.Error(codes.Internal, "could not get payment")
				}
				if !ok || payment.UserID != claims.UserID {
					return status.Error(codes.NotFound, "payment not found")
				}

				if payment.Status == domain.PaymentStatusProcessing {
					if err := s.payments.UpdateStatus(ctx, cmd.PaymentId, domain.PaymentStatusFailed); err != nil {
						return status.Error(codes.Internal, "could not update payment")
					}
					stream.Send(&pb.PaymentEvent{
						PaymentId: cmd.PaymentId,
						Status:    pb.PaymentStatus_PAYMENT_STATUS_FAILED,
						Message:   "payment failed",
					})
				}
			}
		case err := <-errCh:
			return err
		case <-ctx.Done():
			return nil
		}
	}
}
