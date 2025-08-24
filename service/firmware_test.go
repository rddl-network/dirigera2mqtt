package service_test

import (
	"os"
	"testing"

	"github.com/rddl-network/dirigera2mqtt/service"
	"github.com/stretchr/testify/assert"
)

func TestFirmwareIntegrityVerification(t *testing.T) {
	t.Parallel()

	firmware, err := os.ReadFile("../test/energy-intelligence-bridge.bin_merged")
	//firmware := firmware_org[0x20000:]
	offset := 0x0
	assert.NoError(t, err)
	valid := service.VerifyBinaryIntegrity(firmware, offset)
	assert.True(t, valid)

	offset = 0x20000
	assert.NoError(t, err)
	valid = service.VerifyBinaryIntegrity(firmware, offset)
	assert.True(t, valid)
}

func TestFirmwareHandling(t *testing.T) {
	t.Parallel()

	firmware, err := os.ReadFile("../test/energy-intelligence-bridge.bin_merged")
	//firmware := firmware_org[0x20000:]
	offset := 0x20000
	assert.NoError(t, err)
	valid := service.VerifyBinaryIntegrity(firmware, offset)
	assert.True(t, valid)

	patchedFirmware := service.PatchFirmware(firmware, "mynetwork", "mypassword", offset)
	invalid := service.VerifyBinaryIntegrity(patchedFirmware[:], offset)
	assert.False(t, invalid)

	correctedFirmware := service.ComputeAndSetFirmwareChecksum(patchedFirmware, offset)
	valid = service.VerifyBinaryIntegrity(correctedFirmware, offset)
	assert.True(t, valid)
}
