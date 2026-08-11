package tasks

import (
	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func domainToProto(t *domain.Task) *pb.Task {
	return &pb.Task{
		Id:          t.ID,
		UserId:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      pb.TaskStatus(t.Status),
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func protoToDomain(t *pb.Task) *domain.Task {
	return &domain.Task{
		ID:          t.Id,
		UserID:      t.UserId,
		Title:       t.Title,
		Description: t.Description,
		Status:      domain.TaskStatus(t.Status),
	}
}
