package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"net"
	"strconv"
	"strings"
)

func ParseFlags() (map[string]string, error) {
	var addr = pflag.StringP("address", "a", "localhost:8080", "Bind address for the server in the format host:port")
	var pollInterval = pflag.StringP("pollInterval", "p", "2", "pollInterval in seconds")
	var reportInterval = pflag.StringP("reportInterval", "r", "10", "reportInterval in seconds")
	pflag.Parse()

	args := map[string]string{
		"reportInterval": *reportInterval,
		"pollInterval":   *pollInterval,
	}

	for arg, value := range args {
		if err := validateArgs(value); err != nil {
			return nil, fmt.Errorf("invalid %s: %v", arg, err)
		}
	}
	err := validateAddress(*addr)
	if err != nil {
		return nil, err
	}
	args["addr"] = *addr
	return args, nil
}

func validateArgs(s string) error {
	_, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("value must be in seconds")
	}
	return nil
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
		fmt.Errorf("port must be between 1 and 65535")
	}
	if net.ParseIP(host) == nil {
		if _, err := net.LookupHost(host); err != nil {
			return fmt.Errorf("invalid host: %v", err)
		}
	}
	return nil
}
