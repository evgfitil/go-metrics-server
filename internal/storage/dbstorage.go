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
		db.updateCounter(ctx, metric)
	case "gauge":
		db.updateGauge(ctx, metric)
	}
}

func (db *DBStorage) updateCounter(ctx context.Context, metric *metrics.Metrics) {
}

func (db *DBStorage) updateGauge(ctx context.Context, metric *metrics.Metrics) {
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
