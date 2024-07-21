package main

import "testing"

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
		{
			name:    "non-numeric port",
			args:    args{addr: "localhost:port"},
			wantErr: true,
		},
		{
			name:    "port less than 1",
			args:    args{addr: "localhost:0"},
			wantErr: true,
		},
		{
			name:    "non-existent host",
			args:    args{addr: "nonexistenthost:8080"},
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
