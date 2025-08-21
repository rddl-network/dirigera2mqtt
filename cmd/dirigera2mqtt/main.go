package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	"github.com/rddl-network/dirgera2mqtt/config"
	"github.com/rddl-network/dirgera2mqtt/service"
	"github.com/spf13/viper"
)

var libConfig *lib.Config

func init() {
	encodingConfig := app.MakeEncodingConfig()
	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)
}

func loadConfig(path string) (cfg *config.Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err == nil {
		cfg = config.GetConfig()
		cfg.ServiceBind = v.GetString("SERVICE_BIND")
		cfg.ServicePort = v.GetInt("SERVICE_PORT")
		cfg.FirmwareESP32C6 = v.GetString("FIRMWARE_ESP32C6")
		cfg.LogLevel = v.GetString("LOG_LEVEL")
		return
	}
	log.Println("no config file found")

	tmpl := template.New("appConfigFileTemplate")
	configTemplate, err := tmpl.Parse(config.DefaultConfigTemplate)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	if err = configTemplate.Execute(&buffer, config.GetConfig()); err != nil {
		return
	}

	if err = v.ReadConfig(&buffer); err != nil {
		return
	}
	if err = v.SafeWriteConfig(); err != nil {
		return
	}

	log.Println("default config file created. please adapt it and restart the application. exiting...")
	os.Exit(0)
	return
}

func main() {
	cfg, err := loadConfig("./")
	if err != nil {
		log.Fatalf("fatal error reading the configuration %s", err)
	}

	fmt.Println("Web Service mode")

	Dirigera2MQTTService := service.NewTrustAnchorAttestationService(cfg)
	err = Dirigera2MQTTService.Run()
	if err != nil {
		fmt.Print(err.Error())
	}
}
