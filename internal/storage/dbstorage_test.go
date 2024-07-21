package storage

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

// setupMockDB sets up a mock database and logger for testing
func setupMockDB(t *testing.T) (*DBStorage, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	return &DBStorage{connPool: db}, mock
}

func TestDBStorage_Close(t *testing.T) {
	storage, mock := setupMockDB(t)
	defer storage.connPool.Close()

	mock.ExpectClose()

	err := storage.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_Ping(t *testing.T) {
	storage, mock := setupMockDB(t)
	defer storage.connPool.Close()

	mock.ExpectPing()

	err := storage.Ping(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_Update(t *testing.T) {
	storage, mock := setupMockDB(t)
	defer storage.connPool.Close()

	counterMetric := &metrics.Metrics{
		ID:    "test_counter",
		MType: "counter",
		Delta: func() *int64 { v := int64(10); return &v }(),
	}
	gaugeMetric := &metrics.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: func() *float64 { v := 42.42; return &v }(),
	}

	mock.ExpectExec("INSERT INTO counter").WithArgs(counterMetric.ID, *counterMetric.Delta).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO gauge").WithArgs(gaugeMetric.ID, gaugeMetric.Value).WillReturnResult(sqlmock.NewResult(1, 1))

	storage.Update(context.Background(), counterMetric)
	storage.Update(context.Background(), gaugeMetric)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_Get(t *testing.T) {
	logger.InitLogger()
	storage, mock := setupMockDB(t)
	defer storage.connPool.Close()

	counterMetric := &metrics.Metrics{
		ID:    "test_counter",
		MType: "counter",
		Delta: func() *int64 { v := int64(10); return &v }(),
	}
	gaugeMetric := &metrics.Metrics{
		ID:    "test_gauge",
		MType: "gauge",
		Value: func() *float64 { v := 42.42; return &v }(),
	}

	mock.ExpectQuery("SELECT id, delta FROM counter WHERE id = \\$1").
		WithArgs(counterMetric.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "delta"}).AddRow(counterMetric.ID, *counterMetric.Delta))

	mock.ExpectQuery("SELECT id, value FROM gauge WHERE id = \\$1").
		WithArgs(gaugeMetric.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "value"}).AddRow(gaugeMetric.ID, *gaugeMetric.Value))

	// Successful counter metric retrieval
	m, found := storage.Get(context.Background(), counterMetric.ID, "counter")
	assert.True(t, found)
	assert.Equal(t, counterMetric, m)

	// Successful gauge metric retrieval
	m, found = storage.Get(context.Background(), gaugeMetric.ID, "gauge")
	assert.True(t, found)
	assert.Equal(t, gaugeMetric, m)

	// Test for no rows found
	mock.ExpectQuery("SELECT id, delta FROM counter WHERE id = \\$1").
		WithArgs("non_existent").
		WillReturnError(sql.ErrNoRows)

	m, found = storage.Get(context.Background(), "non_existent", "counter")
	assert.False(t, found)
	assert.Nil(t, m)

	// Test for connection error
	mock.ExpectQuery("SELECT id, delta FROM counter WHERE id = \\$1").
		WithArgs("conn_error").
		WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionException})

	m, found = storage.Get(context.Background(), "conn_error", "counter")
	assert.False(t, found)
	assert.Nil(t, m)

	// Test for other query error
	mock.ExpectQuery("SELECT id, delta FROM counter WHERE id = \\$1").
		WithArgs("other_error").
		WillReturnError(errors.New("some other error"))

	m, found = storage.Get(context.Background(), "other_error", "counter")
	assert.False(t, found)
	assert.Nil(t, m)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_newMetricsCache(t *testing.T) {
	mc := newMetricsCache()

	if mc == nil {
		t.Fatalf("exepcted non-nil metricsCache")
	}

	if mc.cache == nil {
		t.Fatalf("exepected non-nil map for cache")
	}

	if len(mc.cache) != 0 {
		t.Errorf("expected empty map for cache")
	}

	var mu sync.Mutex
	if mc.mu != mu {
		t.Errorf("expected default initialized sync.Mutex")
	}
}
