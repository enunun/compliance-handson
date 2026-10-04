// Package server は，APIのHTTPハンドラを定める．
package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/enunun/compliance-handson/app/apps/api/internal/store"
)

// MaxNoteLength はメモの本文の最大文字数である．packages/sharedの検証と合わせる．
const MaxNoteLength = 1000

type ctxKey struct{}

// New はAPIのハンドラを返す．
func New(s store.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/login", login(s))
	mux.Handle("GET /api/notes", authenticated(s, listNotes(s)))
	mux.Handle("POST /api/notes", authenticated(s, createNote(s)))
	mux.Handle("PUT /api/users/{id}/role", authenticated(s, adminOnly(setRole(s))))
	return mux
}

func login(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		u, err := s.UserByEmail(r.Context(), req.Email)
		if err != nil || bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(req.Password)) != nil {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		token, err := newToken()
		if err == nil {
			err = s.CreateSession(r.Context(), u.ID, token)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create session")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}

func listNotes(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notes, err := s.ListNotes(r.Context(), currentUser(r).ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not list notes")
			return
		}
		if notes == nil {
			notes = []store.Note{}
		}
		writeJSON(w, http.StatusOK, notes)
	}
}

func createNote(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		body := strings.TrimSpace(req.Body)
		if body == "" || len([]rune(body)) > MaxNoteLength {
			writeError(w, http.StatusBadRequest, "body must be 1 to 1000 characters")
			return
		}
		n, err := s.CreateNote(r.Context(), currentUser(r).ID, body)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create note")
			return
		}
		writeJSON(w, http.StatusCreated, n)
	}
}

func setRole(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var req struct {
			Role store.Role `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
			(req.Role != store.RoleAdmin && req.Role != store.RoleMember) {
			writeError(w, http.StatusBadRequest, "role must be admin or member")
			return
		}
		switch err := s.SetRole(r.Context(), id, req.Role); {
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "user not found")
		case err != nil:
			writeError(w, http.StatusInternalServerError, "could not set role")
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// authenticated は，Authorizationヘッダのトークンで利用者を確かめる．
func authenticated(s store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		u, err := s.UserBySession(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
	})
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r).Role != store.RoleAdmin {
			writeError(w, http.StatusForbidden, "admin only")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func currentUser(r *http.Request) store.User {
	return r.Context().Value(ctxKey{}).(store.User)
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
