package config

import (
	"sync"

	"github.com/rddl-network/go-utils/logger"
)

const DefaultConfigTemplate = `
FIRMWARE_ESP32C6="{{ .FirmwareESP32C6 }}"
SERVICE_BIND="{{ .ServiceBind }}"
SERVICE_PORT={{ .ServicePort }}
LOG_LEVEL="{{ .LogLevel }}"
`

// Config defines TA's top level configuration
type Config struct {
	FirmwareESP32C6 string `json:"firmware-esp32-c6"   mapstructure:"firmware-esp32-c6"`
	ServiceBind     string `json:"service-bind"        mapstructure:"service-bind"`
	ServicePort     int    `json:"service-port"        mapstructure:"service-port"`
	LogLevel        string `json:"log-level"           mapstructure:"log-level"`
}

// global singleton
var (
	config     *Config
	initConfig sync.Once
)

// DefaultConfig returns TA's default configuration.
func DefaultConfig() *Config {
	return &Config{
		FirmwareESP32C6: "./tasmota32c6-rddl.bin",
		ServiceBind:     "localhost",
		ServicePort:     8080,
		LogLevel:        logger.DEBUG,
	}
}

// GetConfig returns the config instance for the SDK.
func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
		config.LogLevel = logger.DEBUG
	})
}

// GetConfig returns the config instance for the SDK.
func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
	})
	return config
}
