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
func (m *MockPgDb) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
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

// MockRows is a mock implementation of the pgx.Rows interface
type MockRows struct {
	mock.Mock
}

func (m *MockRows) Next() bool {
	return m.Called().Bool(0)
}
func (m *MockRows) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}
func (m *MockRows) Err() error {
	return m.Called().Error(0)
}
func (m *MockRows) Close() {
	m.Called()
}
func (m *MockRows) CommandTag() pgconn.CommandTag {
	return m.Called().Get(0).(pgconn.CommandTag)
}
func (m *MockRows) Conn() *pgx.Conn {
	return m.Called().Get(0).(*pgx.Conn)
}
func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return m.Called().Get(0).([]pgconn.FieldDescription)
}
func (m *MockRows) Values() ([]any, error) {
	return m.Called().Get(0).([]any), m.Called().Error(1)
}
func (m *MockRows) RawValues() [][]byte {
	return m.Called().Get(0).([][]byte)
}

// MockBatchResults is a mock implementation of the pgx.BatchResults interface
type MockBatchResults struct {
	mock.Mock
}

func (m *MockBatchResults) QueryRow() pgx.Row {
	return m.Called().Get(0).(pgx.Row)
}
func (m *MockBatchResults) Query() (pgx.Rows, error) {
	args := m.Called()
	return args.Get(0).(pgx.Rows), args.Error(1)
}
func (m *MockBatchResults) Exec() (pgconn.CommandTag, error) {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}
func (m *MockBatchResults) Close() error {
	return m.Called().Error(0)
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
func TestUpdateUserGroups(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("MERGE 1")

		mockDb.On("Exec", ctx, mock.Anything, []any{[]int{1, 2, 3}, "user1"}).Return(mockTag, nil)

		err := manager.UpdateUserGroups(ctx, "user1", []int{1, 2, 3})
		assert.NoError(t, err)

		mockDb.AssertExpectations(t)
	})

	t.Run("database error on exec", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()

		mockDb.On("Exec", ctx, mock.Anything, []any{[]int{1, 2, 3}, "user1"}).Return(pgconn.CommandTag{}, errors.New("db error"))

		err := manager.UpdateUserGroups(ctx, "user1", []int{1, 2, 3})
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
	})
}
func TestDeleteGroup(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("DELETE 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Exec", ctx, "DELETE FROM groups WHERE id = $1 AND version = $2", []any{1, 1}).Return(mockTag, nil)

		err := manager.DeleteGroup(ctx, 1)
		assert.NoError(t, err)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		err := manager.DeleteGroup(ctx, 1)
		assertPolicyStoreError(t, err, store.NewGroupNotFoundError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on query row", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("db error"))

		err := manager.DeleteGroup(ctx, 1)
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on delete", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Exec", ctx, "DELETE FROM groups WHERE id = $1 AND version = $2", []any{1, 1}).Return(pgconn.CommandTag{}, errors.New("db error"))

		err := manager.DeleteGroup(ctx, 1)
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("concurrency error", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("DELETE 0")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Exec", ctx, "DELETE FROM groups WHERE id = $1 AND version = $2", []any{1, 1}).Return(mockTag, nil)

		err := manager.DeleteGroup(ctx, 1)
		assertPolicyStoreError(t, err, store.NewConcurrencyError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
func TestChangeGroupName(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("UPDATE 1")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Exec", ctx, "UPDATE groups SET name = $1, version = version + 1 WHERE id = $2 AND version = $3", []any{"new-group-name", 1, 1}).Return(mockTag, nil)

		err := manager.ChangeGroupName(ctx, 1, "new-group-name")
		assert.NoError(t, err)

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		err := manager.ChangeGroupName(ctx, 1, "new-group-name")
		assertPolicyStoreError(t, err, store.NewGroupNotFoundError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on query row", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		mockDb.On("QueryRow", ctx, "SELECT version FROM groups WHERE id = $1", []any{1}).Return(mockRow)
		mockRow.On("Scan", mock.Anything).Return(errors.New("db error"))

		err := manager.ChangeGroupName(ctx, 1, "new-group-name")
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("database error on exec update", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Exec", ctx, "UPDATE groups SET name = $1, version = version + 1 WHERE id = $2 AND version = $3", []any{"new-group-name", 1, 1}).Return(pgconn.CommandTag{}, errors.New("db error"))

		err := manager.ChangeGroupName(ctx, 1, "new-group-name")
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("concurrency error", func(t *testing.T) {
		mockDb, _, mockRow, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("UPDATE 0")

		setupMockQueryRow(mockDb, mockRow, ctx, 1, 1)
		mockDb.On("Exec", ctx, "UPDATE groups SET name = $1, version = version + 1 WHERE id = $2 AND version = $3", []any{"new-group-name", 1, 1}).Return(mockTag, nil)

		err := manager.ChangeGroupName(ctx, 1, "new-group-name")
		assertPolicyStoreError(t, err, store.NewConcurrencyError())

		mockDb.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("DELETE 1")

		mockDb.On("Exec", ctx, "DELETE FROM subjects WHERE id = $1", []any{"user1"}).Return(mockTag, nil)

		err := manager.DeleteUser(ctx, "user1")
		assert.NoError(t, err)

		mockDb.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()

		mockDb.On("Exec", ctx, "DELETE FROM subjects WHERE id = $1", []any{"user1"}).Return(pgconn.CommandTag{}, errors.New("db error"))

		err := manager.DeleteUser(ctx, "user1")
		assertPolicyStoreError(t, err, store.NewDataBaseError())

		mockDb.AssertExpectations(t)
	})

	t.Run("no user records found for deletion", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockTag := pgconn.NewCommandTag("DELETE 0")

		mockDb.On("Exec", ctx, "DELETE FROM subjects WHERE id = $1", []any{"user1"}).Return(mockTag, nil)

		err := manager.DeleteUser(ctx, "user1")
		assertPolicyStoreError(t, err, store.NewNoUserRecordsDeletedError())

		mockDb.AssertExpectations(t)
	})
}
func TestReadPolicy(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)
		mockRowsPermissions := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, nil).Once()
		mockBatchResults.On("Query").Return(mockRowsPermissions, nil).Once()
		mockBatchResults.On("Close").Return(nil)

		// Mock group users query
		mockRowsGroups.On("Next").Return(true).Once()
		mockRowsGroups.On("Next").Return(false).Once()
		mockRowsGroups.On("Scan", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*(args[0].([]any)[0].(*string)) = "group1"
				*(args[0].([]any)[1].(*string)) = "user1"
			}).Return(nil)
		mockRowsGroups.On("Err").Return(nil)

		// Mock permissions query
		mockRowsPermissions.On("Next").Return(true).Once()
		mockRowsPermissions.On("Next").Return(false).Once()
		mockRowsPermissions.On("Scan", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*(args[0].([]any)[0].(*string)) = "permission1"
				*(args[0].([]any)[1].(*string)) = "group1"
			}).Return(nil)
		mockRowsPermissions.On("Err").Return(nil)

		policy, err := manager.ReadPolicy(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Len(t, policy.Groups, 1)
		assert.Len(t, policy.Permissions, 1)
		assert.Equal(t, "group1", policy.Groups[0].Name)
		assert.Equal(t, []string{"user1"}, policy.Groups[0].Users)
		assert.Equal(t, "permission1", policy.Permissions[0].Name)
		assert.Equal(t, []string{"group1"}, policy.Permissions[0].Groups)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
		mockRowsGroups.AssertExpectations(t)
		mockRowsPermissions.AssertExpectations(t)
	})

	t.Run("database error on group users query", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, errors.New("db error")).Once()
		mockBatchResults.On("Close").Return(nil)

		policy, err := manager.ReadPolicy(ctx)
		assertPolicyStoreError(t, err, store.NewDataBaseError())
		assert.Nil(t, policy)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
	})

	t.Run("error scanning group users", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, nil).Once()
		mockBatchResults.On("Close").Return(nil)

		mockRowsGroups.On("Next").Return(true).Once()
		mockRowsGroups.On("Scan", mock.Anything, mock.Anything).Return(errors.New("scan error"))

		policy, err := manager.ReadPolicy(ctx)
		assertPolicyStoreError(t, err, store.NewDefaultError())
		assert.Nil(t, policy)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
		mockRowsGroups.AssertExpectations(t)
	})

	t.Run("error reading group users", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, nil).Once()
		mockBatchResults.On("Close").Return(nil)

		mockRowsGroups.On("Next").Return(false).Once()
		mockRowsGroups.On("Err").Return(errors.New("read error"))

		policy, err := manager.ReadPolicy(ctx)
		assertPolicyStoreError(t, err, store.NewDefaultError())
		assert.Nil(t, policy)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
		mockRowsGroups.AssertExpectations(t)
	})

	t.Run("database error on permissions query", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, nil).Once()
		mockBatchResults.On("Query").Return(mockRowsGroups, errors.New("db error")).Once()
		mockBatchResults.On("Close").Return(nil)

		mockRowsGroups.On("Next").Return(false).Once()
		mockRowsGroups.On("Err").Return(nil)

		policy, err := manager.ReadPolicy(ctx)
		assertPolicyStoreError(t, err, store.NewDataBaseError())
		assert.Nil(t, policy)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
		mockRowsGroups.AssertExpectations(t)
	})

	t.Run("error scanning permissions", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)
		mockRowsPermissions := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, nil).Once()
		mockBatchResults.On("Query").Return(mockRowsPermissions, nil).Once()
		mockBatchResults.On("Close").Return(nil)

		mockRowsGroups.On("Next").Return(false).Once()
		mockRowsGroups.On("Err").Return(nil)

		mockRowsPermissions.On("Next").Return(true).Once()
		mockRowsPermissions.On("Scan", mock.Anything, mock.Anything).
			Return(errors.New("scan error"))
		//mockRowsPermissions.On("Err").Return(nil)

		policy, err := manager.ReadPolicy(ctx)
		assertPolicyStoreError(t, err, store.NewDefaultError())
		assert.Nil(t, policy)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
		mockRowsGroups.AssertExpectations(t)
		mockRowsPermissions.AssertExpectations(t)
	})

	t.Run("error reading permissions", func(t *testing.T) {
		mockDb, _, _, manager := setupMockDbAndManager()
		mockBatchResults := new(MockBatchResults)
		mockRowsGroups := new(MockRows)
		mockRowsPermissions := new(MockRows)

		mockDb.On("SendBatch", ctx, mock.Anything).Return(mockBatchResults)
		mockBatchResults.On("Query").Return(mockRowsGroups, nil).Once()
		mockBatchResults.On("Query").Return(mockRowsPermissions, nil).Once()
		mockBatchResults.On("Close").Return(nil)

		mockRowsGroups.On("Next").Return(false).Once()
		mockRowsGroups.On("Err").Return(nil)
		mockRowsPermissions.On("Next").Return(false).Once()
		mockRowsPermissions.On("Err").Return(errors.New("read error"))

		policy, err := manager.ReadPolicy(ctx)
		assertPolicyStoreError(t, err, store.NewDefaultError())
		assert.Nil(t, policy)

		mockDb.AssertExpectations(t)
		mockBatchResults.AssertExpectations(t)
		mockRowsGroups.AssertExpectations(t)
		mockRowsPermissions.AssertExpectations(t)
	})
}
