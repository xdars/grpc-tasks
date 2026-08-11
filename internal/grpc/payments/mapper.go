package payments

import (
	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/domain"
)

func domainToProto(p *domain.Payment) *pb.Payment {
	return &pb.Payment{
		Id:       p.ID,
		UserId:   p.UserID,
		Type:     pb.PaymentType(p.Type),
		Status:   pb.PaymentStatus(p.Status),
		Amount:   p.Amount,
		Currency: p.Currency,
	}
}
