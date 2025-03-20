package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salmarsumi/recipes/internal/authz/store"
)

type pgPool interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
}

// PostgresPolicyManager is a Postgres implementation of the PolicyManager interface.
type PostgresPolicyManager struct {
	db     pgPool
	logger *slog.Logger
}

// NewPostgresPolicyManager creates a new PostgresPolicyManager instance.
func NewPostgresPolicyManager(db pgPool, logger *slog.Logger) *PostgresPolicyManager {
	return &PostgresPolicyManager{db: db, logger: logger}
}

// UpdateGroupPermissions updates the permissions for the specified group.
func (manager *PostgresPolicyManager) UpdateGroupPermissions(ctx context.Context, groupId int, permissions []string) error {
	logger := manager.logger.With("group_id", groupId)
	conn, err := manager.db.Acquire(ctx)
	if err != nil {
		logger.Error("failed to acquire database connection", "error", err)
		return store.NewDataBaseError()
	}
	defer conn.Release()

	var version int
	err = conn.QueryRow(ctx, "SELECT version FROM groups WHERE id = $1", groupId).Scan(&version)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Error("group not found")
			return store.NewGroupNotFoundError()
		}

		logger.Error("failed to query group version", "error", err)
		return store.NewDataBaseError()
	}

	// start a new transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		logger.Error("failed to start transaction", "error", err)
		return store.NewDataBaseError()
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil {
			logger.Error("failed to rollback transaction", "error", err)
		}
	}()

	// update the group permissions
	_, err = tx.Exec(ctx, "DELETE FROM groups_permissions WHERE group_id = $1", groupId)
	if err != nil {
		logger.Error("failed to delete group permissions", "error", err)
		return store.NewDataBaseError()
	}

	_, err = tx.Exec(ctx, "INSERT INTO groups_permissions (group_id, permission_id) VALUES $1", groupId, permissions)
	if err != nil {
		logger.Error("failed to insert group permissions", "error", err)
		return store.NewDataBaseError()
	}

	// update the group version
	tags, err := tx.Exec(ctx, "UPDATE groups SET version = version + 1 WHERE id = $1 AND version = $2", groupId, version)
	if err != nil {
		logger.Error("failed to update group version", "error", err)
		return store.NewDataBaseError()
	}
	if tags.RowsAffected() == 0 {
		logger.Error("failed to update group version due to concurrency issue")
		return store.NewConcurrencyError()
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error("failed to commit transaction", "error", err)
		return store.NewDataBaseError()
	}

	return nil
}
