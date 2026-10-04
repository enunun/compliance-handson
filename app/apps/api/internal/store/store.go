// Package store は，APIが扱うデータの保存先を定める．
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound は，探したデータがないことを表す．
var ErrNotFound = errors.New("not found")

// Role は利用者の権限である．
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// User は利用者である．
type User struct {
	ID           int64
	Email        string
	PasswordHash []byte
	Role         Role
}

// Note は利用者が書くメモである．
type Note struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store はデータの保存先である．
type Store interface {
	UserByEmail(ctx context.Context, email string) (User, error)
	UserBySession(ctx context.Context, token string) (User, error)
	CreateSession(ctx context.Context, userID int64, token string) error
	SetRole(ctx context.Context, userID int64, role Role) error
	ListNotes(ctx context.Context, ownerID int64) ([]Note, error)
	CreateNote(ctx context.Context, ownerID int64, body string) (Note, error)
}
