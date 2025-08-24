package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rddl-network/dirigera2mqtt/config"
	"github.com/rddl-network/go-utils/logger"
)

type Dirigera2MQTT struct {
	cfg             *config.Config
	router          *gin.Engine
	logger          logger.AppLogger
	firmwareESP32C6 []byte
}

func NewTrustAnchorAttestationService(cfg *config.Config) *Dirigera2MQTT {
	service := &Dirigera2MQTT{
		cfg:    cfg,
		logger: logger.GetLogger(cfg.LogLevel),
	}

	gin.SetMode(gin.ReleaseMode)
	service.router = gin.New()
	service.router.GET("/firmware/:mcu/:ssid/:pwd", service.getFirmware)

	return service
}

func (s *Dirigera2MQTT) Run() (err error) {
	s.loadFirmwares()
	err = s.startWebService()
	if err != nil {
		fmt.Print(err.Error())
	}
	return err
}

func (s *Dirigera2MQTT) loadFirmwares() {
	s.firmwareESP32C6 = loadFirmware(s.cfg.FirmwareESP32C6)
}

func (s *Dirigera2MQTT) startWebService() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.ServiceBind, s.cfg.ServicePort)
	err := s.router.Run(addr)

	return err
}
