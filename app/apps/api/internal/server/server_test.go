package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/enunun/compliance-handson/app/apps/api/internal/server"
	"github.com/enunun/compliance-handson/app/apps/api/internal/store"
)

// memStore はテスト用の，メモリに保存するStoreである．
type memStore struct {
	mu       sync.Mutex
	users    []store.User
	sessions map[string]int64
	notes    []store.Note
}

func newMemStore(t *testing.T) *memStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return &memStore{
		users: []store.User{
			{ID: 1, Email: "admin@example.com", PasswordHash: hash, Role: store.RoleAdmin},
			{ID: 2, Email: "member@example.com", PasswordHash: hash, Role: store.RoleMember},
		},
		sessions: map[string]int64{},
	}
}

func (m *memStore) find(match func(store.User) bool) (store.User, error) {
	for _, u := range m.users {
		if match(u) {
			return u, nil
		}
	}
	return store.User{}, store.ErrNotFound
}

func (m *memStore) UserByEmail(_ context.Context, email string) (store.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.find(func(u store.User) bool { return u.Email == email })
}

func (m *memStore) UserBySession(_ context.Context, token string) (store.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.sessions[token]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	return m.find(func(u store.User) bool { return u.ID == id })
}

func (m *memStore) CreateSession(_ context.Context, userID int64, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = userID
	return nil
}

func (m *memStore) SetRole(_ context.Context, userID int64, role store.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.users {
		if m.users[i].ID == userID {
			m.users[i].Role = role
			return nil
		}
	}
	return store.ErrNotFound
}

func (m *memStore) ListNotes(_ context.Context, ownerID int64) ([]store.Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []store.Note
	for _, n := range m.notes {
		if n.OwnerID == ownerID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (m *memStore) CreateNote(_ context.Context, ownerID int64, body string) (store.Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := store.Note{ID: int64(len(m.notes) + 1), OwnerID: ownerID, Body: body, CreatedAt: time.Now()}
	m.notes = append(m.notes, n)
	return n, nil
}

func do(t *testing.T, h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func loginAs(t *testing.T, h http.Handler, email string) string {
	t.Helper()
	rec := do(t, h, "POST", "/api/login", "", `{"email":"`+email+`","password":"secret"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: status %d", rec.Code)
	}
	var res struct{ Token string }
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	return res.Token
}

func TestHealthz(t *testing.T) {
	rec := do(t, server.New(newMemStore(t)), "GET", "/healthz", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	rec := do(t, server.New(newMemStore(t)), "POST", "/api/login", "",
		`{"email":"admin@example.com","password":"wrong"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestNotesRequireToken(t *testing.T) {
	rec := do(t, server.New(newMemStore(t)), "GET", "/api/notes", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestCreateAndListOwnNotes(t *testing.T) {
	h := server.New(newMemStore(t))
	admin := loginAs(t, h, "admin@example.com")
	member := loginAs(t, h, "member@example.com")

	if rec := do(t, h, "POST", "/api/notes", admin, `{"body":"hello"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create: status %d", rec.Code)
	}

	var notes []store.Note
	rec := do(t, h, "GET", "/api/notes", admin, "")
	if err := json.NewDecoder(rec.Body).Decode(&notes); err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Body != "hello" {
		t.Fatalf("admin notes = %+v", notes)
	}

	rec = do(t, h, "GET", "/api/notes", member, "")
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("member notes = %s", rec.Body)
	}
}

func TestCreateNoteRejectsEmptyBody(t *testing.T) {
	h := server.New(newMemStore(t))
	token := loginAs(t, h, "admin@example.com")
	if rec := do(t, h, "POST", "/api/notes", token, `{"body":"  "}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestOnlyAdminCanSetRole(t *testing.T) {
	h := server.New(newMemStore(t))
	member := loginAs(t, h, "member@example.com")
	if rec := do(t, h, "PUT", "/api/users/2/role", member, `{"role":"admin"}`); rec.Code != http.StatusForbidden {
		t.Fatalf("member: status %d", rec.Code)
	}
	admin := loginAs(t, h, "admin@example.com")
	if rec := do(t, h, "PUT", "/api/users/2/role", admin, `{"role":"admin"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("admin: status %d", rec.Code)
	}
}
