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

func (d *DBStorage) Ping(ctx context.Context) error {
	return d.connPool.PingContext(ctx)
}

func (d *DBStorage) Update(metric *metrics.Metrics) {
}

func (d *DBStorage) Get(metricName string) (*metrics.Metrics, bool) {
	return nil, false
}

func (d *DBStorage) GetAllMetrics() map[string]*metrics.Metrics {
	return nil
}

func (d *DBStorage) SaveMetrics() error {
	return nil
}
