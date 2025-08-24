package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Dirigera2MQTT) GetRouter() *gin.Engine {
	return s.router
}

func (s *Dirigera2MQTT) getFirmware(c *gin.Context) {
	mcu := c.Param("mcu")
	var req FirmwareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	ssid := req.SSID
	pwd := req.PWD

	var filename string
	var firmwareBytes []byte
	switch mcu {
	case "esp32c6":
		firmwareBytes = s.firmwareESP32C6
		filename = "dirigerac2mqtt_esp32c6.bin"
	default:
		c.String(404, "Resource not found, Firmware not supported")
		return
	}
	fmt.Printf("Request: {mcu: %s, ssid: %s, pwd: %s}\n", mcu, ssid, pwd)
	patchedFirmware := PatchFirmware(firmwareBytes[:], ssid, pwd, 0x20000)
	ComputeAndSetFirmwareChecksum(patchedFirmware[:], 0x20000)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", patchedFirmware[:])
}

func (s *Dirigera2MQTT) GetRoutes() gin.RoutesInfo {
	routes := s.router.Routes()
	return routes
}
