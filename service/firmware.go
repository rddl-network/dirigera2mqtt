package service

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
)

func ComputeAndSetFirmwareChecksum(patchedBinary []byte, offset int) (correctedBinaryPatch []byte) {
	correctedBinaryPatch = patchedBinary[:]
	patchedBinary = correctedBinaryPatch[offset:]
	binaryChecksum, imageOffset := xorSegments(patchedBinary[:])
	chkOffset := getChecksumOffset(imageOffset)
	patchedBinary[chkOffset] = binaryChecksum

	isHashAppended := patchedBinary[0x17] == 0x1
	if isHashAppended {
		sha256Hash := sha256.Sum256(patchedBinary[0 : chkOffset+1])
		copy(patchedBinary[chkOffset+1:chkOffset+1+32], sha256Hash[:])
	}

	copy(correctedBinaryPatch[offset:], patchedBinary[:])
	return
}

func getChecksumOffset(offset int) int {
	if offset%16 == 0 {
		return offset + 16 - 1
	}

	// slow alternative ist
	// return ((offset + 15) / 16) * 16
	return (offset+15)&^15 - 1
}

func ValidateHash(binary []byte, chkOffset int) bool {
	binarySize := len(binary)
	if chkOffset < 0 || chkOffset+1+32 > binarySize {
		return false
	}
	var binaryHash [32]byte
	copy(binaryHash[:], binary[chkOffset+1:chkOffset+1+32])
	sha256Hash := sha256.Sum256(binary[0 : chkOffset+1])
	return binaryHash == sha256Hash
}

func VerifyBinaryIntegrity(binary []byte, offset int) bool {
	binary = binary[offset:]
	binaryChecksum, offset := xorSegments(binary)
	chkOffset := getChecksumOffset(offset)
	isHashAppended := binary[0x17] == 0x1
	isHashValid := true
	if isHashAppended {
		isHashValid = ValidateHash(binary, chkOffset)
	}
	if binary[chkOffset] == binaryChecksum && isHashValid {
		fmt.Printf("The integrity of the file got verified. The checksum is: %x\n", binaryChecksum)
		return true
	}

	fmt.Printf("Attention: File integrity check FAILED. The files checksum is: %x, the computed checksum is: %x\n", binary[chkOffset], binaryChecksum)
	return false
}

func patchValue(pattern string, value string, firmware []byte) (patchedFirmware []byte) {
	var searchBytes [64]byte
	copy(searchBytes[:], pattern)

	var replacementBuffer [64]byte
	copy(replacementBuffer[:], value)

	patchedFirmware = bytes.Replace(firmware, searchBytes[:], replacementBuffer[:], 1)
	return
}

func PatchFirmware(firmware []byte, ssid string, pwd string, offset int) []byte {
	//toBePatchedImageInFirmware := firmware[offset:]

	SSID_PATTERN := "WIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSIDWIFISSID"

	var searchBytes [64]byte
	copy(searchBytes[:], SSID_PATTERN)

	var replacementBuffer [64]byte
	copy(replacementBuffer[:], ssid)

	patchedSSIDImage := bytes.Replace(firmware[offset:], searchBytes[:], replacementBuffer[:], 1)
	// Write original and patched buffers to files for debugging
	// Verify if ssid is contained in patchedSSIDImage
	if !bytes.Contains(patchedSSIDImage, []byte(ssid)) {
		fmt.Printf("Warning: SSID '%s' not found in patchedSSIDImage\n", ssid)
	}

	pwdPattern := "WIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDWIFIPWDW"
	completlyPatchedImage := patchValue(pwdPattern, pwd, patchedSSIDImage)
	// Verify if ssid and pwd are contained in completlyPatchedImage
	if !bytes.Contains(completlyPatchedImage[:], []byte(ssid)) {
		fmt.Printf("Warning: SSID '%s' not found in completlyPatchedImage\n", ssid)
	}
	if !bytes.Contains(completlyPatchedImage[:], []byte(pwd)) {
		fmt.Printf("Warning: PWD '%s' not found in completlyPatchedImage\n", pwd)
	}

	copy(firmware[offset:], completlyPatchedImage[:])

	return firmware[:]
}

func loadFirmware(filename string) []byte {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic("could not read firmware " + filename)
	}

	if !VerifyBinaryIntegrity(content, 0x0) {
		panic("given firmware integrity check 1 failed: " + filename)
	}
	if !VerifyBinaryIntegrity(content, 0x20000) {
		panic("given firmware integrity check 2 failed: " + filename)
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

func xorSegments(binary []byte) (computedChecksum byte, offset int) {
	// init variables
	numSegments := int(binary[1])
	headerSize := 8
	extHeaderSize := 16
	offset = headerSize + extHeaderSize // that's where the data segments start

	computedChecksum = byte(0)

	for i := 0; i < numSegments; i++ {
		offset += 4 // the segments load address
		length := toInt(binary[:], offset)
		offset += 4 // the read integer
		// xor from here to offset + length for length bytes
		computedChecksum = xorDataBlob(binary[:], offset, length, i == 0, computedChecksum)
		offset += length
	}
	computedChecksum ^= 0xEF
	return
}
