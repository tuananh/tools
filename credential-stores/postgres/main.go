package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/gptscript-ai/gptscript-helper-sqlite/pkg/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	p, err := NewPostgres(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating postgres: %v\n", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionStore, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionGet, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/erase", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionErase, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		if err := credentials.HandleCommand(p, credentials.ActionList, r.Body, w); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	if err := http.ListenAndServe("127.0.0.1:"+port, mux); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}

func NewPostgres(ctx context.Context) (common.Database, error) {
	dsn := os.Getenv("GPTSCRIPT_POSTGRES_DSN")
	if dsn == "" {
		return common.Database{}, fmt.Errorf("missing GPTSCRIPT_POSTGRES_DSN")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return common.Database{}, fmt.Errorf("failed to open database: %w", err)
	}

	return common.NewDatabase(ctx, db)
}
