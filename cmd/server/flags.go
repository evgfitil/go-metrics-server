package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"net"
	"strconv"
	"strings"
)

func ParseFlags() (string, error) {
	var addr = pflag.StringP("address", "a", "localhost:8080", "Bind address for the server in the format host:port")
	pflag.Parse()

	err := validateAddress(*addr)
	if err != nil {
		return "", err
	}
	return *addr, nil
}

func validateAddress(addr string) error {
	hp := strings.Split(addr, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address must be in the forman host:port")
	}

	host, portString := hp[0], hp[1]
	port, err := strconv.Atoi(portString)
	if err != nil {
		return fmt.Errorf("ivalid port: %v", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if net.ParseIP(host) == nil {
		if _, err := net.LookupHost(host); err != nil {
			return fmt.Errorf("invalid host: %v", err)
		}
	}
	return nil
}
