package main

import (
	"context"
	"log"

	pb "github.com/xdars/grpc-tasks/gen/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}()

	client := pb.NewAuthServiceClient(conn)

	// Test 1: Public method (Login) - should skip auth interceptor
	log.Println("=== Test 1: Public Method (Login) ===")
	loginResp, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: "demo",
		Password: "demo123",
	})
	if err != nil {
		log.Printf("Login failed: %v", err)
	} else {
		log.Printf("Login success! Token: %s...", loginResp.Token[:20])
	}

	// Test 2: Protected method without auth - should be blocked
	log.Println("\n=== Test 2: Protected Method Without Auth ===")
	taskClient := pb.NewTaskServiceClient(conn)
	_, err = taskClient.ListTasks(context.Background(), &pb.ListTasksRequest{})
	if err != nil {
		log.Printf("ListTasks without auth failed (expected): %v", err)
	}

	// Test 3: Protected method with auth - should work
	log.Println("\n=== Test 3: Protected Method With Auth ===")
	if loginResp != nil {
		// Create context with auth token
		md := metadata.New(map[string]string{"authorization": "Bearer " + loginResp.Token})
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		// Try to list tasks
		listResp, err := taskClient.ListTasks(ctx, &pb.ListTasksRequest{})
		if err != nil {
			log.Printf("ListTasks with auth failed: %v", err)
		} else {
			log.Printf("ListTasks success! Found %d tasks", len(listResp.Tasks))
		}

		// Try to create a task
		createResp, err := taskClient.CreateTask(ctx, &pb.CreateTaskRequest{
			Title:       "Test via gRPC",
			Description: "Created using gRPC client to demonstrate interceptors",
		})
		if err != nil {
			log.Printf("CreateTask failed: %v", err)
		} else {
			log.Printf("CreateTask success! Task ID: %s", createResp.Id)
		}
	}
}
