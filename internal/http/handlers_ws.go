package httpserver

import (
	"log"
	"net/http"

	"github.com/xdars/grpc-tasks/internal/grpc/interceptors"
	"github.com/xdars/grpc-tasks/internal/jwt"
)

func (h *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(interceptors.UserClaimsKey).(*jwt.Claims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}()

	ch := h.store.Subscribe(claims.UserID)
	defer h.store.Unsubscribe(claims.UserID)

	for event := range ch {
		if err := conn.WriteJSON(event); err != nil {
			log.Println("websocket write error:", err)
			break
		}
	}
}
