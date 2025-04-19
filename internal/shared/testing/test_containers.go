package testing

import (
	"context"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

// CreatePostgresContainer creates and starts a PostgreSQL container using the specified
// database name and initialization script. It returns a PostgresContainer instance
// containing the container and its connection string, or an error if the container
// could not be created.
//
// Parameters:
//   - ctx: The context to control the container lifecycle.
//   - dbName: The name of the database to be created in the container.
//   - initScript: The path to the initialization script to be executed in the container.
//
// Returns:
//   - *PostgresContainer: A struct containing the running PostgreSQL container and its
//     connection string.
//   - error: An error if the container creation or connection string retrieval fails.
func CreatePostgresContainer(ctx context.Context, dbName string, initScript string) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithInitScripts(initScript),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(uuid.NewString()),
		postgres.WithPassword(uuid.NewString()),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}
