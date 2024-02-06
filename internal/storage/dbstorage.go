package storage

import (
	"context"
	"database/sql"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const driverName = "pgx"

type DBStorage struct {
	connPool *sql.DB
}

func NewDBStorage(databaseDSN string) (*DBStorage, error) {
	var db DBStorage
	conn, err := sql.Open(driverName, databaseDSN)
	if err != nil {
		logger.Sugar.Fatalf("unable to connect to database: %v", err)
		return nil, err
	}
	db = DBStorage{connPool: conn}
	return &db, nil
}

func (db *DBStorage) Ping(ctx context.Context) error {
	return db.connPool.PingContext(ctx)
}

func (db *DBStorage) Update(ctx context.Context, metric *metrics.Metrics) {
	switch metric.MType {
	case "counter":
		err := db.updateCounter(ctx, metric)
		if err != nil {
			logger.Sugar.Errorf("error updating counter metric: %v", err)
		}
	case "gauge":
		err := db.updateGauge(ctx, metric)
		if err != nil {
			logger.Sugar.Errorf("error updating gauge metric: %v", err)
		}
	}
}

func (db *DBStorage) updateCounter(ctx context.Context, metric *metrics.Metrics) error {
	var currentDelta *int64

	row := db.connPool.QueryRowContext(ctx, "SELECT delta FROM counter WHERE id = $1", metric.ID)
	err := row.Scan(&currentDelta)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if currentDelta != nil {
		*currentDelta += *metric.Delta
	} else {
		currentDelta = metric.Delta
	}

	_, err = db.connPool.ExecContext(ctx,
		"INSERT INTO counter (id, delta) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET delta = $2", metric.ID, *currentDelta)
	return err
}

func (db *DBStorage) updateGauge(ctx context.Context, metric *metrics.Metrics) error {
	_, err := db.connPool.ExecContext(ctx,
		"INSERT INTO gauge (id, value) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET value = $2", metric.ID, metric.Value)
	return err
}

func (db *DBStorage) Get(ctx context.Context, metricName string) (*metrics.Metrics, bool) {
	return nil, false
}

func (db *DBStorage) GetAllMetrics(ctx context.Context) map[string]*metrics.Metrics {
	return nil
}

func (db *DBStorage) SaveMetrics(_ context.Context) error {
	return nil
}
