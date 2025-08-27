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
type FirmwareRequest struct {
	SSID           string `json:"ssid"`
	PWD            string `json:"pwd"`
	LiquidAddress  string `json:"liquid_address,omitempty"`
	DirAuthToken   string `json:"dir_auth_token,omitempty"`
	DirURI         string `json:"dir_uri,omitempty"`
}

func NewTrustAnchorAttestationService(cfg *config.Config) *Dirigera2MQTT {
	service := &Dirigera2MQTT{
		cfg:    cfg,
		logger: logger.GetLogger(cfg.LogLevel),
	}

	gin.SetMode(gin.ReleaseMode)
	service.router = gin.New()
	// CORS middleware
	service.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	service.router.POST("/firmware/:mcu", service.getFirmware)

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

//func (s *Dirigera2MQTT) getFirmware(c *gin.Context) {
//	mcu := c.Param("mcu")
//	var req FirmwareRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(400, gin.H{"error": "Invalid request body"})
//		return
//	}
//	ssid := req.SSID
//	pwd := req.PWD
//	// TODO: Add logic to patch and serve firmware using ssid, pwd, and mcu
//	c.JSON(200, gin.H{"message": "Firmware request received", "mcu": mcu, "ssid": ssid, "pwd": pwd})
//}
//
