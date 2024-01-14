package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	serverAddress  string
	pollInterval   string
	reportInterval string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) setAndValidate(key string, defaultValue string) error {
	var value string
	switch key {
	case "REPORT_INTERVAL":
		value = getEnvOrDefault(key, defaultValue)
		if err := validateIntervalArgs(value); err != nil {
			return err
		}
		c.reportInterval = value
	case "POLL_INTERVAL":
		value = getEnvOrDefault(key, defaultValue)
		if err := validateIntervalArgs(value); err != nil {
			return err
		}
		c.pollInterval = value
	case "ADDRESS":
		value = getEnvOrDefault(key, defaultValue)
		if err := validateAddress(value); err != nil {
			return err
		}
		c.serverAddress = value
	}
	return nil
}

func getEnvOrDefault(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) ParseFlags() error {
	var addr = pflag.StringP("address", "a", "localhost:8080", "Bind address for the server in the format host:port")
	var pollIntervalArg = pflag.StringP("pollInterval", "p", "2", "pollInterval in seconds")
	var reportIntervalArg = pflag.StringP("reportInterval", "r", "10", "reportInterval in seconds")
	pflag.Parse()
	envsAndArgs := map[string]string{
		"ADDRESS":         *addr,
		"POLL_INTERVAL":   *pollIntervalArg,
		"REPORT_INTERVAL": *reportIntervalArg,
	}
	for env, flag := range envsAndArgs {
		if err := c.setAndValidate(env, flag); err != nil {
			return err
		}
	}
	return nil
}

func validateIntervalArgs(s string) error {
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
