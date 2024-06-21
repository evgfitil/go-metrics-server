package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/internal/mocks"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
)

func testMetricsRouter(s storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Route("/value", func(r chi.Router) {
		r.Post("/", GetMetricsJSON(s))
		r.Get("/{type}/{name}", GetMetricsPlain(s))
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", UpdateMetricsJSON(s))
		r.Post("/{type}/{name}/{value}", UpdateMetricsPlain(s))
	})
	return r
}

//func float64Ptr(f float64) *float64 {
//	return &f
//}
//
//func int64Ptr(i int64) *int64 {
//	return &i
//}

func createMockStorageWithMetrics(ctrl *gomock.Controller, mockMetrics map[string]*metrics.Metrics) *mocks.MockStorage {
	mockStorage := mocks.NewMockStorage(ctrl)
	mockStorage.EXPECT().GetAllMetrics(gomock.Any()).Return(mockMetrics)
	return mockStorage
}

func TestGetMetricsJsonHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	metricDelta := int64(100)
	storedMetric := &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: &metricDelta}
	mockStorage.EXPECT().Get(gomock.Any(), "testCounter", "counter").Return(storedMetric, true).AnyTimes()
	mockStorage.EXPECT().Get(gomock.Any(), "testTest", "counter").Return(nil, false).AnyTimes()

	ts := httptest.NewServer(testMetricsRouter(mockStorage))
	defer ts.Close()

	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		requestBody   metrics.Metrics
		want          want
	}{
		{
			name:          "valid request",
			requestMethod: http.MethodPost,
			requestPath:   "/value/",
			requestBody:   metrics.Metrics{ID: "testCounter", MType: "counter"},
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id":"testCounter","type":"counter","delta":100}`,
			},
		},
		{
			name:          "wrong requestMethod",
			requestMethod: http.MethodGet,
			requestPath:   "/value/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:          "wrong requestPath",
			requestMethod: http.MethodPost,
			requestPath:   "/wrong",
			requestBody:   metrics.Metrics{ID: "testCounter", MType: "counter"},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:          "get non existing metric",
			requestMethod: http.MethodPost,
			requestPath:   "/value/",
			requestBody:   metrics.Metrics{ID: "testTest", MType: "counter"},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.JSONEq(t, tt.want.body, string(body))
			}
		})
	}
}

func TestUpdateMetricsJsonHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()

	ts := httptest.NewServer(testMetricsRouter(mockStorage))
	defer ts.Close()
	validCounterValue := int64(100)
	validGaugeValue := 123.12
	validCounterMetric := metrics.Metrics{ID: "testCounter", MType: "counter", Delta: &validCounterValue}
	validGaugeMetric := metrics.Metrics{ID: "testGauge", MType: "gauge", Value: &validGaugeValue}
	invalidMetric := metrics.Metrics{ID: "testInvalid", MType: "someType", Value: &validGaugeValue}

	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		requestBody   metrics.Metrics
		want          want
	}{
		{
			name:          "valid Counter update",
			requestMethod: http.MethodPost,
			requestPath:   "/update/",
			requestBody:   validCounterMetric,
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id":"testCounter","type":"counter","delta":100}`,
			},
		},
		{
			name:          "valid Gauge update",
			requestMethod: http.MethodPost,
			requestPath:   "/update/",
			requestBody:   validGaugeMetric,
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id":"testGauge","type":"gauge","value":123.12}`,
			},
		},
		{
			name:          "invalid request method",
			requestMethod: http.MethodGet,
			requestPath:   "/update/",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:          "invalid path",
			requestMethod: http.MethodPost,
			requestPath:   "/update/u",
			requestBody:   validGaugeMetric,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:          "invalid metric type",
			requestMethod: http.MethodPost,
			requestPath:   "/update/",
			requestBody:   invalidMetric,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.JSONEq(t, tt.want.body, string(body))
			}
		})
	}
}

func TestGetMetricsHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()
	mockMetricValue := int64(100)
	mockMetric := metrics.Metrics{ID: "testCounter", MType: "counter", Delta: &mockMetricValue}
	mockStorage.Update(context.TODO(), &mockMetric)

	ts := httptest.NewServer(testMetricsRouter(mockStorage))
	defer ts.Close()

	type want struct {
		statusCode int
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		want          want
	}{
		{
			name:          "valid request",
			requestMethod: http.MethodGet,
			requestPath:   "/value/counter/testCounter",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "wrong requestMethod",
			requestMethod: http.MethodPost,
			requestPath:   "/value/counter/testCounter",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:          "wrong requestPath",
			requestMethod: http.MethodGet,
			requestPath:   "/wrong",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:          "get non existing metric",
			requestMethod: http.MethodGet,
			requestPath:   "/value/counter/nonExistingMetric",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestUpdateMetricsHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()

	ts := httptest.NewServer(testMetricsRouter(mockStorage))
	defer ts.Close()

	type want struct {
		statusCode int
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		want          want
	}{
		{
			name:          "valid Counter update",
			requestMethod: http.MethodPost,
			requestPath:   "/update/counter/testCounter/123",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "valid Gauge update",
			requestMethod: http.MethodPost,
			requestPath:   "/update/gauge/testGauge/123.12",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "invalid request method",
			requestMethod: http.MethodGet,
			requestPath:   "/update/gauge/testGauge/123.12",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:          "invalid path",
			requestMethod: http.MethodPost,
			requestPath:   "/update/gauge/test/Gauge/123.12",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:          "invalid metric type",
			requestMethod: http.MethodPost,
			requestPath:   "/update/histogram/testHistogram/123.12",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "invalid Counter value",
			requestMethod: http.MethodPost,
			requestPath:   "/update/counter/testCounter/123.12",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name     string
		metrics  map[string]*metrics.Metrics
		validate func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "multiple existing metrics",
			metrics: map[string]*metrics.Metrics{
				"metric1": {ID: "metric1", MType: "counter", Delta: new(int64)},
				"metric2": {ID: "metric2", MType: "gauge", Value: new(float64)},
			},
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Contains(t, rr.Body.String(), "metric1")
				assert.Contains(t, rr.Body.String(), "metric2")
			},
		},
		{
			name:    "no metrics",
			metrics: make(map[string]*metrics.Metrics),
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Empty(t, rr.Body)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := createMockStorageWithMetrics(ctrl, tt.metrics)

			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := GetAllMetrics(mockStorage)
			handler.ServeHTTP(rr, req)
			tt.validate(t, rr)
		})
	}
}

func TestPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)

	type want struct {
		statusCode int
	}
	tests := []struct {
		name           string
		mockStorageErr error
		want           want
	}{
		{
			name:           "successful ping",
			mockStorageErr: nil,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:           "internal server error",
			mockStorageErr: assert.AnError,
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.EXPECT().Ping(gomock.Any()).Return(tt.mockStorageErr)

			req, err := http.NewRequest(http.MethodGet, "/ping", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := Ping(mockStorage)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
		})
	}
}
