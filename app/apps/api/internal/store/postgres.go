package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            BIGSERIAL PRIMARY KEY,
	email         TEXT NOT NULL UNIQUE,
	password_hash BYTEA NOT NULL,
	role          TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
	token   TEXT PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS notes (
	id         BIGSERIAL PRIMARY KEY,
	owner_id   BIGINT NOT NULL REFERENCES users(id),
	body       TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

// Postgres はPostgreSQLにデータを保存する．
type Postgres struct {
	pool *pgxpool.Pool
}

// OpenPostgres は接続し，テーブルがなければ作る．
func OpenPostgres(ctx context.Context, url string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

// Close は接続を閉じる．
func (p *Postgres) Close() { p.pool.Close() }

// EnsureUser は，メールアドレスの利用者がいなければ作る．
func (p *Postgres) EnsureUser(ctx context.Context, email string, passwordHash []byte, role Role) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
		 ON CONFLICT (email) DO NOTHING`, email, passwordHash, string(role))
	return err
}

func (p *Postgres) UserByEmail(ctx context.Context, email string) (User, error) {
	return p.queryUser(ctx,
		`SELECT id, email, password_hash, role FROM users WHERE email = $1`, email)
}

func (p *Postgres) UserBySession(ctx context.Context, token string) (User, error) {
	return p.queryUser(ctx,
		`SELECT u.id, u.email, u.password_hash, u.role
		 FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token = $1`, token)
}

func (p *Postgres) queryUser(ctx context.Context, sql string, arg any) (User, error) {
	var u User
	var role string
	err := p.pool.QueryRow(ctx, sql, arg).Scan(&u.ID, &u.Email, &u.PasswordHash, &role)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	u.Role = Role(role)
	return u, err
}

func (p *Postgres) CreateSession(ctx context.Context, userID int64, token string) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id) VALUES ($1, $2)`, token, userID)
	return err
}

func (p *Postgres) SetRole(ctx context.Context, userID int64, role Role) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE users SET role = $1 WHERE id = $2`, string(role), userID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (p *Postgres) ListNotes(ctx context.Context, ownerID int64) ([]Note, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, owner_id, body, created_at FROM notes
		 WHERE owner_id = $1 ORDER BY id`, ownerID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Note, error) {
		var n Note
		err := r.Scan(&n.ID, &n.OwnerID, &n.Body, &n.CreatedAt)
		return n, err
	})
}

func (p *Postgres) CreateNote(ctx context.Context, ownerID int64, body string) (Note, error) {
	n := Note{OwnerID: ownerID, Body: body}
	err := p.pool.QueryRow(ctx,
		`INSERT INTO notes (owner_id, body) VALUES ($1, $2) RETURNING id, created_at`,
		ownerID, body).Scan(&n.ID, &n.CreatedAt)
	return n, err
}
