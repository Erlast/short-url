package storages

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мокаем зависимости.
type MockMigrationRunner struct {
	mock.Mock
}

func (m *MockMigrationRunner) RunMigrations(dsn string) error {
	args := m.Called(dsn)
	if err := args.Error(0); err != nil {
		return errors.New(err.Error())
	}
	return nil
}

type MockPoolInitializer struct {
	mock.Mock
}

func (m *MockPoolInitializer) InitPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	args := m.Called(ctx, dsn)

	pool, ok := args.Get(0).(*pgxpool.Pool)
	if !ok {
		return nil, errors.New("unexpected type for pgxpool.Pool")
	}

	if err := args.Error(1); err != nil {
		return nil, errors.New(err.Error())
	}

	return pool, nil
}

func TestNewPgStorage(t *testing.T) {
	ctx := context.Background()

	mockMigrationRunner := new(MockMigrationRunner)
	mockPoolInitializer := new(MockPoolInitializer)
	mockPool := &pgxpool.Pool{}

	tests := []struct {
		name               string
		runMigrationsError error
		initPoolError      error
		expectedError      error
	}{
		{
			name:               "Successful initialization",
			runMigrationsError: nil,
			initPoolError:      nil,
			expectedError:      nil,
		},
		//{
		//	name:               "Error running migrations",
		//	runMigrationsError: errors.New("migration error"),
		//	initPoolError:      nil,
		//	expectedError:      errors.New("failed to run DB migrations: migration error"),
		// },
		//{
		//	name:               "Error initializing pool",
		//	runMigrationsError: nil,
		//	initPoolError:      errors.New("connection error"),
		//	expectedError:      errors.New("unable to connect database: connection error"),
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrationRunner.On("RunMigrations", mock.Anything).Return(tt.runMigrationsError)
			mockPoolInitializer.On("InitPool", mock.Anything, mock.Anything).Return(mockPool, tt.initPoolError)

			ps, err := NewPgStorage(ctx, "mock_dsn", mockMigrationRunner, mockPoolInitializer)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, ps)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ps)
			}
		})
	}
}
