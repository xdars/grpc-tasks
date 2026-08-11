package httpserver

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/xdars/grpc-tasks/internal/grpc/auth"
	"github.com/xdars/grpc-tasks/internal/grpc/tasks"
	"github.com/xdars/grpc-tasks/internal/store"
	wa "github.com/xdars/grpc-tasks/internal/webauthn"
)

type Server struct {
	authService  *auth.AuthService
	taskService  *tasks.TaskService
	store        *store.Store
	waService    *wa.Service
	waSessions   map[string]*webauthn.SessionData
	waSessionsMu sync.Mutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func New(a *auth.AuthService, t *tasks.TaskService, s *store.Store, w *wa.Service) *Server {
	return &Server{
		authService: a,
		taskService: t,
		store:       s,
		waService:   w,
		waSessions:  make(map[string]*webauthn.SessionData),
	}
}

func (s *Server) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cmd/server/static/index.html")
	})

	r.Post("/register", s.handleRegister)
	r.Post("/login", s.handleLogin)

	r.Post("/webauthn/register/begin", s.handleWebAuthnRegisterBegin)
	r.Post("/webauthn/register/finish", s.handleWebAuthnRegisterFinish)
	r.Post("/webauthn/login/begin", s.handleWebAuthnLoginBegin)
	r.Post("/webauthn/login/finish", s.handleWebAuthnLoginFinish)

	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Post("/tasks", s.handleCreateTask)
		r.Get("/tasks", s.handleListTasks)
		r.Get("/ws", s.handleWebSocket)
	})

	return r
}
