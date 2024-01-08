package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"net"
	"os"
	"strconv"
	"strings"
)

func ParseFlags() (map[string]string, error) {
	var addr = pflag.StringP("address", "a", "localhost:8080", "Bind address for the server in the format host:port")
	var pollIntervalArg = pflag.StringP("pollInterval", "p", "2", "pollInterval in seconds")
	var reportIntervalArg = pflag.StringP("reportInterval", "r", "10", "reportInterval in seconds")
	var pollInterval, reportInterval, serverAddress string
	pflag.Parse()

	args := make(map[string]string)
	if pollIntervalEnv := os.Getenv("POLL_INTERVAL"); pollIntervalEnv != "" {
		pollInterval = pollIntervalEnv
	} else {
		pollInterval = *pollIntervalArg
	}
	args["pollInterval"] = pollInterval

	if reportIntervalEnv := os.Getenv("REPORT_INTERVAL"); reportIntervalEnv != "" {
		reportInterval = reportIntervalEnv
	} else {
		reportInterval = *reportIntervalArg
	}
	args["reportInterval"] = reportInterval

	for arg, value := range args {
		if err := validateArgs(value); err != nil {
			return nil, fmt.Errorf("invalid %s: %v", arg, err)
		}
	}
	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	} else {
		serverAddress = *addr
	}
	err := validateAddress(serverAddress)
	if err != nil {
		return nil, err
	}
	args["addr"] = serverAddress
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
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if net.ParseIP(host) == nil {
		if _, err := net.LookupHost(host); err != nil {
			return fmt.Errorf("invalid host: %v", err)
		}
	}
	return nil
}
