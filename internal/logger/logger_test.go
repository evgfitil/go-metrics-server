package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Initialize Logger"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitLogger()
			if Sugar == nil {
				t.Errorf("InitLogger() failed, Sugar logger is nil")
			}
		})
	}
}

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			h http.Handler
		}
		want http.Handler
	}{
		{
			name: "Middleware Logging",
			args: struct{ h http.Handler }{h: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})},
			want: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithLogging(tt.args.h)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			logOutput := captureLogs(func() {
				got.ServeHTTP(rec, req)
			})

			assert := rec.Body.String() == "OK"
			if !assert {
				t.Errorf("WithLogging() got = %v, want %v", rec.Body.String(), "OK")
			}
			if !containsLogDetails(logOutput, "/test", "GET", "200") {
				t.Errorf("WithLogging() log = %v, missing expected details", logOutput)
			}
		})
	}
}

func containsLogDetails(logOutput, uri, method, status string) bool {
	return bytes.Contains([]byte(logOutput), []byte(uri)) &&
		bytes.Contains([]byte(logOutput), []byte(method)) &&
		bytes.Contains([]byte(logOutput), []byte(status))
}

func Test_loggingResponseWriter_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	tests := []struct {
		name   string
		fields struct {
			ResponseWriter http.ResponseWriter
			responseData   *responseData
		}
		args struct {
			b []byte
		}
		want    int
		wantErr bool
	}{
		{
			name: "Write bytes",
			fields: struct {
				ResponseWriter http.ResponseWriter
				responseData   *responseData
			}{
				ResponseWriter: rec,
				responseData:   &responseData{},
			},
			args: struct{ b []byte }{b: []byte("hello")},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &loggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				responseData:   tt.fields.responseData,
			}
			got, err := r.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Write() got = %v, want %v", got, tt.want)
			}
			if r.responseData.size != tt.want {
				t.Errorf("Write() responseData.size = %v, want %v", r.responseData.size, tt.want)
			}
		})
	}
}

func Test_loggingResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	tests := []struct {
		name   string
		fields struct {
			ResponseWriter http.ResponseWriter
			responseData   *responseData
		}
		args struct {
			statusCode int
		}
	}{
		{
			name: "Write header",
			fields: struct {
				ResponseWriter http.ResponseWriter
				responseData   *responseData
			}{
				ResponseWriter: rec,
				responseData:   &responseData{},
			},
			args: struct{ statusCode int }{statusCode: http.StatusAccepted},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &loggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				responseData:   tt.fields.responseData,
			}
			r.WriteHeader(tt.args.statusCode)
			if r.responseData.status != tt.args.statusCode {
				t.Errorf("WriteHeader() responseData.status = %v, want %v", r.responseData.status, tt.args.statusCode)
			}
			if rec.Code != tt.args.statusCode {
				t.Errorf("WriteHeader() rec.Code = %v, want %v", rec.Code, tt.args.statusCode)
			}
		})
	}
}

func captureLogs(f func()) string {
	var buf bytes.Buffer
	writer := zapcore.AddSync(&buf)
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		writer,
		zapcore.DebugLevel,
	)
	logger := zap.New(core)
	Sugar = logger.Sugar()
	defer Sugar.Sync()
	f()
	return buf.String()
}
