package wa

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/xdars/grpc-tasks/internal/db"
	"github.com/xdars/grpc-tasks/internal/domain"
)

type Service struct {
	wa      *webauthn.WebAuthn
	users   *db.UserRepository
	waCreds *db.WebAuthnRepository
}

func NewService(users *db.UserRepository, waCreds *db.WebAuthnRepository) (*Service, error) {
	w, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "gRPC Tasks",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})
	if err != nil {
		return nil, err
	}
	return &Service{wa: w, users: users, waCreds: waCreds}, nil
}

func (s *Service) toWebAuthnUser(ctx context.Context, record *domain.User) (*WebAuthnUser, error) {
	creds, err := s.waCreds.GetCredentials(ctx, record.ID)
	if err != nil {
		return nil, err
	}
	return &WebAuthnUser{Record: record, Credentials: creds}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*WebAuthnUser, bool, error) {
	record, ok, err := s.users.GetByID(ctx, id)
	if err != nil || !ok {
		return nil, false, err
	}
	user, err := s.toWebAuthnUser(ctx, record)
	return user, true, err
}

func (s *Service) GetUser(ctx context.Context, username string) (*WebAuthnUser, bool, error) {
	record, ok, err := s.users.Get(ctx, username)
	if err != nil || !ok {
		return nil, false, err
	}
	user, err := s.toWebAuthnUser(ctx, record)
	return user, true, err
}

func (s *Service) BeginRegistration(ctx context.Context, user *WebAuthnUser) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	return s.wa.BeginRegistration(user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
	)
}

func (s *Service) FinishRegistration(ctx context.Context, user *WebAuthnUser, session *webauthn.SessionData, r *http.Request) error {
	cred, err := s.wa.FinishRegistration(user, *session, r)
	if err != nil {
		return err
	}
	return s.waCreds.SaveCredential(ctx, user.Record.ID, cred)
}

func (s *Service) BeginLogin(ctx context.Context, user *WebAuthnUser) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return s.wa.BeginDiscoverableLogin()
}

func (s *Service) FinishLogin(ctx context.Context, session *webauthn.SessionData, r *http.Request) (string, error) {
	var userID string
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		userID = string(userHandle)
		record, ok, err := s.users.GetByID(ctx, userID)
		if err != nil || !ok {
			return nil, fmt.Errorf("user not found")
		}
		creds, err := s.waCreds.GetCredentials(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &WebAuthnUser{Record: record, Credentials: creds}, nil
	}

	_, err := s.wa.FinishDiscoverableLogin(handler, *session, r)
	return userID, err
}
