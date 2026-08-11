package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	jwt "github.com/xdars/grpc-tasks/internal/jwt"
)

func (h *Server) handleWebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, ok, err := h.waService.GetUser(r.Context(), req.Username)
	if err != nil || !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	options, session, err := h.waService.BeginRegistration(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.waSessionsMu.Lock()
	h.waSessions[req.Username] = session
	h.waSessionsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(options); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func (h *Server) handleWebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}

	h.waSessionsMu.Lock()
	session, ok := h.waSessions[username]
	h.waSessionsMu.Unlock()
	if !ok {
		http.Error(w, "session not found, call begin first", http.StatusBadRequest)
		return
	}

	user, ok, err := h.waService.GetUser(r.Context(), username)
	if err != nil || !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := h.waService.FinishRegistration(r.Context(), user, session, r); err != nil {
		log.Printf("FinishRegistration error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.waSessionsMu.Lock()
	delete(h.waSessions, username)
	h.waSessionsMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (h *Server) handleWebAuthnLoginBegin(w http.ResponseWriter, r *http.Request) {
	options, session, err := h.waService.BeginLogin(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.waSessionsMu.Lock()
	h.waSessions["discoverable"] = session
	h.waSessionsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(options); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func (h *Server) handleWebAuthnLoginFinish(w http.ResponseWriter, r *http.Request) {
	h.waSessionsMu.Lock()
	session, ok := h.waSessions["discoverable"]
	h.waSessionsMu.Unlock()
	if !ok {
		http.Error(w, "session not found", http.StatusBadRequest)
		return
	}

	userID, err := h.waService.FinishLogin(r.Context(), session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.waSessionsMu.Lock()
	delete(h.waSessions, "discoverable")
	h.waSessionsMu.Unlock()

	user, ok, err := h.waService.GetUserByID(r.Context(), userID)
	if err != nil || !ok {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}

	token, err := jwt.Generate(user.Record.Username, user.Record.ID)
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token":    token,
		"username": user.Record.Username,
	}); err != nil {
		log.Printf("encode error: %v", err)
	}
}
