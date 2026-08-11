package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	pb "github.com/xdars/grpc-tasks/gen/pb"
)

func (h *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.CreateTask(r.Context(), &pb.CreateTaskRequest{
		Title:       req.Title,
		Description: req.Description,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func (h *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	task, err := h.taskService.ListTasks(r.Context(), &pb.ListTasksRequest{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("encode error: %v", err)
	}
}
