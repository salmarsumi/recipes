package postgres

import (
	"context"
	"errors"
	"io"
	"testing"

	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/salmarsumi/recipes/internal/authz/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPgDb is a mock implementation of the pgPool interface
type MockPgDb struct {
	mock.Mock
}

func (m *MockPgDb) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.Called(ctx, sql, args).Get(0).(pgx.Row)
}

func (m *MockPgDb) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockPgDb) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// MockTx is a mock implementation of the pgx.Tx interface
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockTx) Commit(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockTx) Conn() *pgx.Conn {
	return m.Called().Get(0).(*pgx.Conn)
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	return m.Called().Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	arguments := m.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Rows), arguments.Error(1)
}
func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.Called(ctx, sql, args).Get(0).(pgx.Row)
}

// MockRow is a mock implementation of the pgx.Row interface
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

func setupMockDbAndManager() (*MockPgDb, *MockTx, *MockRow, *PostgresPolicyManager) {
	mockDb := new(MockPgDb)
	mockTx := new(MockTx)
	mockRow := new(MockRow)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	manager := NewPostgresPolicyManager(mockDb, logger)
	return mockDb, mockTx, mockRow, manager
}

func setupMockQueryRow(mockDb *MockPgDb, mockRow *MockRow, ctx context.Context, groupId int, version int) {
	mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{groupId}).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*(args[0].([]any)[0].(*int)) = version
	}).Return(nil)
}

func assertPolicyStoreError(t *testing.T, err error, exp error) {
	act := &store.PolicyStoreError{}
	assert.ErrorAs(t, err, &act)
	assert.Equal(t, exp, act)
}

func TestUpdateGroupPermissions(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("INSERT 0 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, nil)
		mockTx.On("Commit", ctx).Return(nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assert.NoError(t, err)

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewGroupNotFoundError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on query row", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("db error"))

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on begin transaction", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, errors.New("db error"))

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("database error on exec merge permissions", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("INSERT 0 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, errors.New("db error"))
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on exec update version", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("INSERT 0 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, "UPDATE groups SET version = version + 1 WHERE id = $1 AND version = $2", []any{1, 1}).Return(mockTag, errors.New("db error"))
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("concurrency error", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("INSERT 0 0")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupPermissions(ctx, 1, []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewConcurrencyError())

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
func TestCreateGroup(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockRow := new(MockRow)

		mockDb.On("QueryRow", ctx, "INSERT INTO groups (name, version) VALUES ($1, 1) RETURNING id", []any{"test-group"}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args[0].([]any)[0].(*int)) = 1
		}).Return(nil)

		id, err := manager.CreateGroup(ctx, "test-group")
		assert.NoError(t, err)
		assert.Equal(t, 1, id)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("group name already exists", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockRow := new(MockRow)

		mockDb.On("QueryRow", ctx, "INSERT INTO groups (name, version) VALUES ($1, 1) RETURNING id", []any{"existing-group"}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		id, err := manager.CreateGroup(ctx, "existing-group")
		assertPolicyStoreError(t, err, store.NewNameExistsError())
		assert.Equal(t, 0, id)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockRow := new(MockRow)

		mockDb.On("QueryRow", ctx, "INSERT INTO groups (name, version) VALUES ($1, 1) RETURNING id", []any{"test-group"}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("db error"))

		id, err := manager.CreateGroup(ctx, "test-group")
		assertPolicyStoreError(t, err, store.NewDataBaseError())
		assert.Equal(t, 0, id)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
func TestCreatePermission(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockRow := new(MockRow)

		mockDb.On("QueryRow", ctx, "INSERT INTO permissions (name, version) VALUES ($1, 1) RETURNING id", []any{"test-permission"}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			*(args[0].([]any)[0].(*int)) = 1
		}).Return(nil)

		id, err := manager.CreatePermission(ctx, "test-permission")
		assert.NoError(t, err)
		assert.Equal(t, 1, id)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("permission name already exists", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockRow := new(MockRow)

		mockDb.On("QueryRow", ctx, "INSERT INTO permissions (name, version) VALUES ($1, 1) RETURNING id", []any{"existing-permission"}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		id, err := manager.CreatePermission(ctx, "existing-permission")
		assertPolicyStoreError(t, err, store.NewNameExistsError())
		assert.Equal(t, 0, id)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockRow := new(MockRow)

		mockDb.On("QueryRow", ctx, "INSERT INTO permissions (name, version) VALUES ($1, 1) RETURNING id", []any{"test-permission"}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("db error"))

		id, err := manager.CreatePermission(ctx, "test-permission")
		assertPolicyStoreError(t, err, store.NewDataBaseError())
		assert.Equal(t, 0, id)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
func TestUpdateGroupUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("UPDATE 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, nil)
		mockTx.On("Commit", ctx).Return(nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assert.NoError(t, err)

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assertPolicyStoreError(t, err, store.NewGroupNotFoundError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on query row", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("db error"))

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on begin transaction", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, errors.New("db error"))

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on exec merge users", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("UPDATE 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, errors.New("db error"))
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on exec update version", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("UPDATE 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, "UPDATE groups SET version = version + 1 WHERE id = $1 AND version = $2", []any{1, 1}).Return(mockTag, errors.New("db error"))
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("concurrency error", func(t *testing.T) {
		mockDb, mockTx, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("UPDATE 0")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("Exec", ctx, mock.Anything, mock.Anything).Return(mockTag, nil)
		mockTx.On("Exec", ctx, "UPDATE groups SET version = version + 1 WHERE id = $1 AND version = $2", []any{1, 1}).Return(mockTag, nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := manager.UpdateGroupUsers(ctx, 1, []string{"user1", "user2"})
		assertPolicyStoreError(t, err, store.NewConcurrencyError())

		mockDb.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
