package main

import (
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/spf13/pflag"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	bindAddress     string
	storeInterval   uint16
	fileStoragePath string
	restore         bool
}

func NewConfig() *Config {
	return &Config{}
}

func getEnvOrDefault(envKey, defaultValue interface{}) interface{} {
	if value := os.Getenv(envKey.(string)); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) setAndValidate(key string, defaultValue interface{}) error {
	value := getEnvOrDefault(key, defaultValue)
	switch key {
	case "ADDRESS":
		if err := validateAddress(value.(string)); err != nil {
			logger.Sugar.Fatalf("invalid bind address: %v", err)
			return err
		}
		c.bindAddress = value.(string)
	case "STORE_INTERVAL":
		strValue, ok := value.(string)
		if !ok {
			c.storeInterval = defaultValue.(uint16)
		} else {
			interval, err := strconv.ParseUint(strValue, 10, 16)
			if err != nil {
				logger.Sugar.Fatalf("invalid store interval: %v", err)
				return err
			}
			c.storeInterval = uint16(interval)
		}
	case "FILE_STORAGE_PATH":
		c.fileStoragePath = value.(string)
	case "RESTORE":
		strValue, ok := value.(string)
		if ok {
			boolValue, err := strconv.ParseBool(strValue)
			if err != nil {
				logger.Sugar.Fatalf("invalid restore value: %v", err)
				return err
			}
			c.restore = boolValue
		} else {
			c.restore = defaultValue.(bool)
		}
	}
	return nil
}

func (c *Config) ParseFlags() error {
	var addr = pflag.StringP("address", "a", "localhost:8080", "Bind address for the server in the format host:port")
	var storeIntervalArg = pflag.Uint16P("storeInterval", "i", 300, "Interval in seconds for storage data to a file")
	var fileStoragePathArg = pflag.StringP("fileStoragePath", "f", "/tmp/metrics-db.json", "File path where the server write its data")
	var restoreArg = pflag.BoolP("restore", "r", true, "Controls loading previously saved values from a file at server startup")
	pflag.Parse()

	envsAndArgs := map[string]interface{}{
		"ADDRESS":           *addr,
		"STORE_INTERVAL":    *storeIntervalArg,
		"FILE_STORAGE_PATH": *fileStoragePathArg,
		"RESTORE":           *restoreArg,
	}

	for env, flag := range envsAndArgs {
		if err := c.setAndValidate(env, flag); err != nil {
			return err
		}
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
