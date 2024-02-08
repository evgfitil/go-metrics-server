package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"sync"
)

const (
	driverName    = "pgx"
	migrationPath = "db/migrations"
)

var wg sync.WaitGroup

type metricsCache struct {
	cache map[string]*metrics.Metrics
	mu    sync.Mutex
}

func newMetricsCache() *metricsCache {
	return &metricsCache{
		cache: make(map[string]*metrics.Metrics),
		mu:    sync.Mutex{},
	}
}

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
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), databaseDSN)
	if err != nil {
		logger.Sugar.Fatalf("error: %v", err)
	}
	err = m.Up()
	if err == migrate.ErrNoChange {
		logger.Sugar.Info("skipping migrations, no changes")
	}
	if err != nil && err != migrate.ErrNoChange {
		logger.Sugar.Fatalf("error applying migrations: %v", err)
	}
	db = DBStorage{connPool: conn}
	return &db, nil
}

func (db *DBStorage) Ping(ctx context.Context) error {
	err := db.connPool.PingContext(ctx)
	if err != nil {
		logger.Sugar.Errorf("error connecting to database: %v", err)
		return err
	}
	return nil
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

func (db *DBStorage) Get(ctx context.Context, metricName string, metricType string) (*metrics.Metrics, bool) {
	var metric metrics.Metrics
	switch metricType {
	case "counter":
		metric.MType = "counter"
		row := db.connPool.QueryRowContext(ctx, "SELECT id, delta FROM counter where id = $1", metricName)
		err := row.Scan(&metric.ID, &metric.Delta)
		if err == sql.ErrNoRows {
			return nil, false
		}
		if err != nil {
			logger.Sugar.Errorf("error retrieving metric: %v", err)
			return nil, false
		}
		return &metric, true
	case "gauge":
		metricType = "gauge"
		row := db.connPool.QueryRowContext(ctx, "SELECT id, value FROM gauge where id = $1", metricName)
		err := row.Scan(&metric.ID, &metric.Value)

		if err == sql.ErrNoRows {
			return nil, false
		}
		if err != nil {
			logger.Sugar.Errorf("error retrieving metric: %v", err)
			return nil, false
		}
		return &metric, true
	}
	return nil, false
}

func (db *DBStorage) GetAllMetrics(ctx context.Context) map[string]*metrics.Metrics {
	allMetrics := newMetricsCache()

	wg.Add(1)
	go func() {
		db.fetchCounterMetrics(ctx, allMetrics)
	}()

	wg.Add(1)
	go func() {
		db.fetchGaugeMetrics(ctx, allMetrics)
	}()
	wg.Wait()

	return allMetrics.cache
}

func (db *DBStorage) fetchCounterMetrics(ctx context.Context, metricsCache *metricsCache) {
	defer wg.Done()

	rows, err := db.connPool.QueryContext(ctx, "SELECT * FROM counter")
	if err != nil && err != sql.ErrNoRows {
		logger.Sugar.Errorf("error retrieving metrics: %v", err)
	}
	defer rows.Close()

	metricsCache.mu.Lock()
	defer metricsCache.mu.Unlock()

	for rows.Next() {
		var m metrics.Metrics
		m.MType = "counter"
		err = rows.Scan(&m.ID, &m.Delta)
		if err != nil {
			logger.Sugar.Errorf("error retrieving metric: %v", err)
		}
		metricsCache.cache[m.ID] = &m
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("error after row iteration: %v", err)
	}
}

func (db *DBStorage) fetchGaugeMetrics(ctx context.Context, metricsCache *metricsCache) {
	defer wg.Done()

	rows, err := db.connPool.QueryContext(ctx, "SELECT * FROM gauge")
	if err != nil && err != sql.ErrNoRows {
		logger.Sugar.Errorf("error retrieving metrics: %v", err)
	}
	defer rows.Close()

	metricsCache.mu.Lock()
	defer metricsCache.mu.Unlock()

	for rows.Next() {
		var m metrics.Metrics
		m.MType = "gauge"
		err = rows.Scan(&m.ID, &m.Value)
		if err != nil {
			logger.Sugar.Errorf("error retrieving metrics: %v", err)
		}
		metricsCache.cache[m.ID] = &m
	}
	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("error after row iteration: %v", err)
	}
}

func (db *DBStorage) UpdateMetrics(ctx context.Context, metrics []*metrics.Metrics) error {
	tx, err := db.connPool.Begin()
	if err != nil {
		logger.Sugar.Errorln("error starting transaction: %v", err)
	}

	for _, metric := range metrics {
		switch metric.MType {
		case "counter":
			_, err = tx.ExecContext(ctx,
				"INSERT INTO counter (id, delta) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET delta = counter.delta + EXCLUDED.delta",
				metric.ID, *metric.Delta)
			if err != nil {
				tx.Rollback()
				return err
			}
		case "gauge":
			_, err = tx.ExecContext(ctx,
				"INSERT INTO gauge (id, value) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET value = $2",
				metric.ID, *metric.Value)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

func (db *DBStorage) SaveMetrics(_ context.Context) error {
	return nil
}
