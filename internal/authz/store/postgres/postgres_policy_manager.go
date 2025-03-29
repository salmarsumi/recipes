package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/salmarsumi/recipes/internal/authz/store"
)

// pgPool is an interface that represents a pool of Postgres connections.
type pgPool interface {
	Acquire(ctx context.Context) (pgConn, error)
}

// pgConn is an interface that represents a Postgres connection.
type pgConn interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Release()
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
func (manager *PostgresPolicyManager) UpdateGroupPermissions(ctx context.Context, groupId int, permissions []int) error {
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

	// merge the new permissions with the existing ones
	_, err = tx.Exec(ctx, `
	WITH new_permissions AS (SELECT unnest($1::int[]) AS permission_id)
	MERGE INTO group_permissions gp
	USING new_permissions np
	ON gp.group_id = $2 AND gp.permission_id = np.permission_id
	WHEN NOT MATCHED BY TARGET THEN
		INSERT (group_id, permission_id) VALUES ($2, np.permission_id)
	WHEN NOT MATCHED BY SOURCE THEN
		DELETE;
	`, permissions, groupId)
	if err != nil {
		logger.Error("failed to merge group permissions", "error", err)
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

// UpdateGroupUsers updates the users for the specified group.
func (manager *PostgresPolicyManager) UpdateGroupUsers(ctx context.Context, groupId int, users []string) error {
	logger := manager.logger.With("group_id", groupId)
	conn, err := manager.db.Acquire(ctx)
	if err != nil {
		logger.Error("failed to acquire database connection", "error", err)
		return store.NewDataBaseError()
	}
	defer conn.Release()

	return nil
}
