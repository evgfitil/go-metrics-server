package main

import (
	"reflect"
	"testing"
)

func TestConfig_GetServerURL(t *testing.T) {
	type fields struct {
		BatchMode      bool
		PollInterval   int
		ReportInterval int
		ServerAddress  string
		UseHTTPS       bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "HTTP URL",
			fields: fields{
				ServerAddress: "localhost:8080",
				UseHTTPS:      false,
			},
			want: "http://localhost:8080",
		},
		{
			name: "HTTPS URL",
			fields: fields{
				ServerAddress: "localhost:8080",
				UseHTTPS:      true,
			},
			want: "https://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				BatchMode:      tt.fields.BatchMode,
				PollInterval:   tt.fields.PollInterval,
				ReportInterval: tt.fields.ReportInterval,
				ServerAddress:  tt.fields.ServerAddress,
				UseHTTPS:       tt.fields.UseHTTPS,
			}
			if got := c.GetServerURL(); got != tt.want {
				t.Errorf("GetServerURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "Default Config",
			want: &Config{
				BatchMode:      false,
				PollInterval:   0,
				ReportInterval: 0,
				ServerAddress:  "",
				UseHTTPS:       false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
