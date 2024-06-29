package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestExecute(t *testing.T) {
	logger.InitLogger()
	defer logger.Sugar.Sync()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "execute without error",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			go func() {
				<-ctx.Done()
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()

			if err := Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_runServer(t *testing.T) {
	logger.InitLogger()
	defer logger.Sugar.Sync()

	type args struct {
		cmd  *cobra.Command
		args []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "run server with in-memory storage",
			args: args{cmd: rootCmd, args: []string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := getFreePort()
			if err != nil {
				fmt.Printf("error getting free port: %v", err)
			}
			address := fmt.Sprintf("localhost:%d", port)

			cfg = &Config{
				BindAddress:     address,
				StoreInterval:   0,
				FileStoragePath: "",
				Restore:         false,
				DatabaseDSN:     "",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			go func() {
				<-ctx.Done()
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()

			runServer(tt.args.cmd, tt.args.args)
		})
	}
}

func Test_validateAddress(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    args{addr: "localhost:8080"},
			wantErr: false,
		},
		{
			name:    "invalid address format",
			args:    args{addr: "localhost"},
			wantErr: true,
		},
		{
			name:    "invalid port",
			args:    args{addr: "localhost:99999"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateAddress(tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("validateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
