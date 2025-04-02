package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/salmarsumi/recipes/internal/authz/store"
)

// pgDb is an interface that represents a pool of Postgres connections.
type pgDb interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PostgresPolicyManager is a Postgres implementation of the PolicyManager interface.
type PostgresPolicyManager struct {
	db     pgDb
	logger *slog.Logger
}

// NewPostgresPolicyManager creates a new PostgresPolicyManager instance.
func NewPostgresPolicyManager(db pgDb, logger *slog.Logger) *PostgresPolicyManager {
	return &PostgresPolicyManager{db: db, logger: logger}
}

// UpdateGroupPermissions updates the permissions for the specified group.
func (manager *PostgresPolicyManager) UpdateGroupPermissions(ctx context.Context, groupId int, permissions []int) error {
	logger := manager.logger.With("group_id", groupId)

	var version int
	err := manager.db.QueryRow(ctx, "SELECT version FROM groups WHERE id = $1", groupId).Scan(&version)
	if err != nil {
		return versionError(err, logger)
	}

	// start a new transaction
	tx, err := manager.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to start transaction", "error", err)
		return store.NewDataBaseError()
	}
	defer rollback(tx, ctx, logger)

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

// CreateGroup creates a new group.
func (manager *PostgresPolicyManager) CreateGroup(ctx context.Context, groupName string) (int, error) {
	logger := manager.logger.With("group_name", groupName)
	var id int
	err := manager.db.QueryRow(ctx, "INSERT INTO groups (name, version) VALUES ($1, 1) RETURNING id", groupName).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			logger.Error("group name already exists")
			return 0, store.NewNameExistsError()
		}

		logger.Error("failed to create group", "error", err)
		return 0, store.NewDataBaseError()
	}

	return id, nil
}

// CreatePermission creates a new permission.
func (manager *PostgresPolicyManager) CreatePermission(ctx context.Context, permissionName string) (int, error) {
	logger := manager.logger.With("permission_name", permissionName)
	var id int
	err := manager.db.QueryRow(ctx, "INSERT INTO permissions (name, version) VALUES ($1, 1) RETURNING id", permissionName).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			logger.Error("permission name already exists")
			return 0, store.NewNameExistsError()
		}

		logger.Error("failed to create permission", "error", err)
		return 0, store.NewDataBaseError()
	}

	return id, nil
}

// UpdateGroupUsers updates the users for the specified group.
func (manager *PostgresPolicyManager) UpdateGroupUsers(ctx context.Context, groupId int, users []string) error {
	logger := manager.logger.With("group_id", groupId)

	var version int
	err := manager.db.QueryRow(ctx, "SELECT version FROM groups WHERE id = $1", groupId).Scan(&version)
	if err != nil {
		return versionError(err, logger)
	}

	// start a new transaction
	tx, err := manager.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to start transaction", "error", err)
		return store.NewDataBaseError()
	}
	defer rollback(tx, ctx, logger)

	// merge the new users with the existing ones
	_, err = tx.Exec(ctx, `
	WITH new_users AS (SELECT unnest($1::text[]) AS user_id)
	MERGE INTO subjects sub
	USING new_users nu
	ON sub.group_id = $2 AND sub.id = nu.user_id
	WHEN NOT MATCHED BY TARGET THEN
		INSERT (group_id, id) VALUES ($2, nu.user_id)
	WHEN NOT MATCHED BY SOURCE THEN
		DELETE;
	`, users, groupId)
	if err != nil {
		logger.Error("failed to merge group users", "error", err)
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

// UpdateUserGroups updates the groups for the specified user.
func (manager *PostgresPolicyManager) UpdateUserGroups(ctx context.Context, userId string, groups []int) error {
	logger := manager.logger.With("user_id", userId)

	// merge the new groups with the existing ones
	_, err := manager.db.Exec(ctx, `
	WITH new_groups AS (SELECT unnest($1::int[]) AS group_id)
	MERGE INTO subjects sub
	USING new_groups ng
	ON sub.group_id = ng.group_id AND sub.id = $2
	WHEN NOT MATCHED BY TARGET THEN
		INSERT (id, group_id) VALUES ($2, ng.group_id)
	WHEN NOT MATCHED BY SOURCE THEN
		DELETE;
	`, groups, userId)
	if err != nil {
		logger.Error("failed to merge user groups", "error", err)
		return store.NewDataBaseError()
	}

	return nil
}

// DeleteGroup deletes the group with the specified id.
func (manager *PostgresPolicyManager) DeleteGroup(ctx context.Context, groupId int) error {
	logger := manager.logger.With("group_id", groupId)

	var version int
	err := manager.db.QueryRow(ctx, "SELECT version FROM groups WHERE id = $1", groupId).Scan(&version)
	if err != nil {
		return versionError(err, logger)
	}

	_, err = manager.db.Exec(ctx, "DELETE FROM groups WHERE id = $1", groupId)
	if err != nil {
		logger.Error("failed to delete group", "error", err)
		return store.NewDataBaseError()
	}

	return nil
}

// 	ChangeGroupName(ctx context.Context, groupId TGroupId, newGroupName string) error
// 	DeleteUser(ctx context.Context, userId TUserId) error
// 	ReadPolicy(ctx context.Context) (*authz.Policy, error)

func rollback(tx pgx.Tx, ctx context.Context, logger *slog.Logger) {
	err := tx.Rollback(ctx)
	if err != nil && err != pgx.ErrTxClosed {
		logger.Error("failed to rollback transaction", "error", err)
	}
}

func versionError(err error, logger *slog.Logger) error {
	if err == pgx.ErrNoRows {
		logger.Error("group not found")
		return store.NewGroupNotFoundError()
	}
	logger.Error("failed to query group version", "error", err)
	return store.NewDataBaseError()
}
