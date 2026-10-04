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

	// 最初の管理者を作る．
	hash, err := bcrypt.GenerateFromPassword([]byte(mustEnv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	if err := db.EnsureUser(ctx, mustEnv("ADMIN_EMAIL"), hash, store.RoleAdmin); err != nil {
		log.Fatalf("create admin: %v", err)
	}

	addr := ":" + envOr("PORT", "8080")
	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.New(db)))
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
