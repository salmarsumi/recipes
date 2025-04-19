package testing

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
