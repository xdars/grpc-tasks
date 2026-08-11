package wa

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/xdars/grpc-tasks/internal/domain"
)

type WebAuthnUser struct {
	Record      *domain.User
	Credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte                         { return []byte(u.Record.ID) }
func (u *WebAuthnUser) WebAuthnName() string                       { return u.Record.Username }
func (u *WebAuthnUser) WebAuthnDisplayName() string                { return u.Record.Username }
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
