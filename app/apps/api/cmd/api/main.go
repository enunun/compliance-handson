// Command api は，題材のSaaSのAPIサーバーである．
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/enunun/compliance-handson/app/apps/api/internal/server"
	"github.com/enunun/compliance-handson/app/apps/api/internal/store"
)

func main() {
	ctx := context.Background()
	db, err := store.OpenPostgres(ctx, mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	// TEST_USER_PASSWORDがあれば，テスト用の管理者と一般の利用者を作る．
	if password := os.Getenv("TEST_USER_PASSWORD"); password != "" {
		if err := createTestUsers(ctx, db, password); err != nil {
			log.Fatalf("create test users: %v", err)
		}
	}

	addr := ":" + envOr("PORT", "8080")
	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.New(db)))
}

func createTestUsers(ctx context.Context, db *store.Postgres, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := db.EnsureUser(ctx, "admin@example.com", hash, store.RoleAdmin); err != nil {
		return err
	}
	return db.EnsureUser(ctx, "member@example.com", hash, store.RoleMember)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
