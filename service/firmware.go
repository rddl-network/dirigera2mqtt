package service

import (
	"bytes"
	"fmt"
	"os"
)

func ComputeAndSetFirmwareChecksum(patchedBinary []byte) {
	binaryChecksum := xorSegments(patchedBinary)
	binarySize := len(patchedBinary)
	patchedBinary[binarySize-1] = binaryChecksum
}

func VerifyBinaryIntegrity(binary []byte) bool {
	binarySize := len(binary)
	binaryChecksum := xorSegments(binary)
	if binary[binarySize-1] == binaryChecksum {
		fmt.Printf("The integrity of the file got verified. The checksum is: %x\n", binaryChecksum)
		return true
	}
	fmt.Printf("Attention: File integrity check FAILED. The files checksum is: %x, the computed checksum is: %x\n", binary[binarySize-1], binaryChecksum)
	return false
}

func PatchFirmware(firmware []byte, ssid string, pwd string) []byte {
	var searchBytes = []byte("WIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSID")
	var patched_ssid = bytes.Replace(firmware, searchBytes, []byte(ssid), 1)

	searchBytes = []byte("WIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWD")
	var completely_patched = bytes.Replace(patched_ssid, searchBytes, []byte(pwd), 1)

	return completely_patched
}

func loadFirmware(filename string) []byte {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic("could not read firmware " + filename)
	}

	if !VerifyBinaryIntegrity(content) {
		panic("given firmware integrity check failed: " + filename)
	}

	return content
}

func toInt(bytes []byte, offset int) int {
	result := 0
	for i := 3; i > -1; i-- {
		result <<= 8
		result += int(bytes[offset+i])
	}
	return result
}

func xorDataBlob(binary []byte, offset int, length int, is1stSegment bool, checksum byte) byte {
	initializer := 0
	if is1stSegment {
		initializer = 1
		checksum = binary[offset]
	}

	for i := initializer; i < length; i++ {
		checksum ^= binary[offset+i]
	}
	return checksum
}

func xorSegments(binary []byte) byte {
	// init variables
	numSegments := int(binary[1])
	headerSize := 8
	extHeaderSize := 16
	offset := headerSize + extHeaderSize // that's where the data segments start

	computedChecksum := byte(0)

	for i := 0; i < numSegments; i++ {
		offset += 4 // the segments load address
		length := toInt(binary, offset)
		offset += 4 // the read integer
		// xor from here to offset + length for length bytes
		computedChecksum = xorDataBlob(binary, offset, length, i == 0, computedChecksum)
		offset += length
	}
	computedChecksum ^= 0xEF

	return computedChecksum
}
