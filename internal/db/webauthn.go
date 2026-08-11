package db

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebAuthnRepository struct {
	pool *pgxpool.Pool
}

func NewWebAuthnRepository(pool *pgxpool.Pool) *WebAuthnRepository {
	return &WebAuthnRepository{pool: pool}
}

func (r *WebAuthnRepository) SaveCredential(ctx context.Context, userID string, cred *webauthn.Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	credID := base64.StdEncoding.EncodeToString(cred.ID)
	_, err = r.pool.Exec(ctx,
		`INSERT INTO webauthn_credentials (id, user_id, credential) VALUES ($1, $2, $3)`,
		credID, userID, data,
	)
	return err
}

func (r *WebAuthnRepository) GetCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT credential FROM webauthn_credentials WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var cred webauthn.Credential
		if err := json.Unmarshal(data, &cred); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}
