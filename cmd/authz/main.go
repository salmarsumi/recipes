package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmarsumi/recipes/internal/authz/store/postgres"
)

// main is the entry point for the authorization application.
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/authz?sslmode=disable")
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		panic(err)
	}

	manager := postgres.NewPostgresPolicyManager(pool, logger)
	id, err := manager.CreateGroup(context.Background(), "new_group")
	if err != nil {
		logger.Error("failed to create group", "error", err)
		id = 1
	}

	pid, errerr := manager.CreatePermission(context.Background(), "new_permission")
	if errerr != nil {
		logger.Error("failed to create permission", "error", errerr)
		pid = 1
	}

	err = manager.UpdateGroupPermissions(context.Background(), id, []int{pid})
	if err != nil {
		logger.Error("failed to update group permissions", "error", err)
	}

	logger.Info("group permissions updated successfully")

	pool.Close()
}
