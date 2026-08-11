package tasks

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/xdars/grpc-tasks/gen/pb"
	"github.com/xdars/grpc-tasks/internal/db"
	"github.com/xdars/grpc-tasks/internal/domain"
	"github.com/xdars/grpc-tasks/internal/grpc/interceptors"
	jwt "github.com/xdars/grpc-tasks/internal/jwt"
	"github.com/xdars/grpc-tasks/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskService struct {
	pb.UnimplementedTaskServiceServer
	tasks *db.TaskRepository
	store *store.Store
}

func NewTaskService(tasks *db.TaskRepository, s *store.Store) *TaskService {
	return &TaskService{tasks: tasks, store: s}
}

func (s *TaskService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	task := &domain.Task{
		ID:     uuid.NewString(),
		UserID: claims.UserID,
		Title:  req.Title,
	}
	if err := s.tasks.Add(ctx, task); err != nil {
		return nil, status.Error(codes.Internal, "could not create task")
	}
	s.store.Publish(&pb.TaskEvent{
		EventType: pb.TaskEvent_EVENT_TYPE_CREATED,
		Task:      domainToProto(task),
	})
	return domainToProto(task), nil
}

func (s *TaskService) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	task, ok, err := s.tasks.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not get task")
	}

	if !ok || task.UserID != claims.UserID {
		return nil, status.Error(codes.NotFound, "task not found")
	}
	return domainToProto(task), nil
}

func (s *TaskService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	tasks, err := s.tasks.GetByUser(ctx, claims.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not get tasks")
	}
	pbTasks := make([]*pb.Task, len(tasks))
	for i, t := range tasks {
		pbTasks[i] = domainToProto(t)
	}
	return &pb.ListTasksResponse{Tasks: pbTasks}, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	task, ok, err := s.tasks.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get task")
	}
	if !ok || task.UserID != claims.UserID {
		return nil, status.Error(codes.NotFound, "task not found")
	}
	if err := s.tasks.Delete(ctx, req.Id); err != nil {
		return &pb.DeleteTaskResponse{Success: false}, nil
	}

	return &pb.DeleteTaskResponse{Success: true}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	claims, ok := ctx.Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}

	task, ok, err := s.tasks.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "task not found")
	}
	if !ok || task.UserID != claims.UserID {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != pb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		task.Status = domain.TaskStatus(req.Status)
	}

	if err := s.tasks.Update(ctx, task); err != nil {
		return nil, status.Error(codes.NotFound, "failed to update task")
	}
	return domainToProto(task), nil
}
