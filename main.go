package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// ESP32 Image Header constants
const (
	ESP_IMAGE_HEADER_MAGIC = 0xE9
	ESP_CHECKSUM_MAGIC     = 0xEF
	CHIP_ID_ESP32C6        = 0x000D
)

// ESPImageHeader represents the ESP32 image header
type ESPImageHeader struct {
	Magic          uint8
	SegmentCount   uint8
	SpiMode        uint8
	SpiSpeedSize   uint8
	EntryAddr      uint32
	WpPin          uint8
	SpiPinDrv      [3]uint8
	ChipID         uint16
	MinChipRev     uint8
	MinChipRevFull uint16
	MaxChipRevFull uint16
	Reserved       [4]uint8
	HashAppend     uint8
}

// ESPSegmentHeader represents a segment header
type ESPSegmentHeader struct {
	LoadAddr uint32
	DataLen  uint32
}

// FirmwareInfo holds the parsed firmware information
type FirmwareInfo struct {
	Header           ESPImageHeader
	Segments         []ESPSegmentHeader
	CalculatedCRC    uint32
	StoredChecksum   uint8
	IsValid          bool
	TotalSize        int64
	HeaderValid      bool
	ChipIDValid      bool
	ChecksumOffset   int64
}

// parseESP32Image parses the ESP32 firmware image and extracts checksum info
func parseESP32Image(filename string) (*FirmwareInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info := &FirmwareInfo{}

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	info.TotalSize = stat.Size()

	// Read and parse header
	err = binary.Read(file, binary.LittleEndian, &info.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Validate header magic
	info.HeaderValid = info.Header.Magic == ESP_IMAGE_HEADER_MAGIC
	info.ChipIDValid = info.Header.ChipID == CHIP_ID_ESP32C6

	if !info.HeaderValid {
		return info, fmt.Errorf("invalid header magic: 0x%02X (expected 0x%02X)", 
			info.Header.Magic, ESP_IMAGE_HEADER_MAGIC)
	}

	fmt.Printf("Segment count: %d\n", info.Header.SegmentCount)

	// Read segment headers and calculate total data size
	var totalDataSize uint32
	currentPos := int64(24) // Size of ESPImageHeader

	for i := 0; i < int(info.Header.SegmentCount); i++ {
		var segHeader ESPSegmentHeader
		
		// Seek to current position
		_, err = file.Seek(currentPos, 0)
		if err != nil {
			return info, fmt.Errorf("failed to seek to segment %d: %w", i, err)
		}

		err = binary.Read(file, binary.LittleEndian, &segHeader)
		if err != nil {
			return info, fmt.Errorf("failed to read segment header %d: %w", i, err)
		}

		info.Segments = append(info.Segments, segHeader)
		
		// Move to next segment header (8 bytes for header + data length)
		currentPos += 8 + int64(segHeader.DataLen)
		totalDataSize += segHeader.DataLen
		
		// Ensure 4-byte alignment
		if currentPos%4 != 0 {
			currentPos += 4 - (currentPos % 4)
		}
	}

	// The checksum should be at the end of all segments
	info.ChecksumOffset = currentPos

	// Read the stored checksum
	_, err = file.Seek(info.ChecksumOffset, 0)
	if err != nil {
		return info, fmt.Errorf("failed to seek to checksum: %w", err)
	}

	err = binary.Read(file, binary.LittleEndian, &info.StoredChecksum)
	if err != nil {
		return info, fmt.Errorf("failed to read stored checksum: %w", err)
	}

	// Calculate ESP32-style simple checksum
	calculatedChecksum, err := calculateSimpleChecksum(filename, info.ChecksumOffset)
	if err != nil {
		return info, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	// Also calculate CRC32 for reference
	info.CalculatedCRC, err = calculateImageCRC(filename, info.ChecksumOffset)
	if err != nil {
		return info, fmt.Errorf("failed to calculate CRC: %w", err)
	}

	// Store the ESP32 checksum as uint32 for display consistency
	info.CalculatedCRC = uint32(calculatedChecksum)
	
	// Verify checksum using ESP32 method
	info.IsValid = calculatedChecksum == info.StoredChecksum

	return info, nil
}

// calculateImageCRC calculates the CRC32 of the image up to the specified offset
func calculateImageCRC(filename string, offset int64) (uint32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Create CRC32 hash
	crc := crc32.NewIEEE()
	
	// Copy only up to the checksum offset
	_, err = io.CopyN(crc, file, offset)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("failed to read file for CRC calculation: %w", err)
	}

	return crc.Sum32(), nil
}

// calculateSimpleChecksum calculates ESP32-style simple checksum
func calculateSimpleChecksum(filename string, offset int64) (uint8, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var checksum uint8 = ESP_CHECKSUM_MAGIC
	buffer := make([]byte, 1024)
	var totalRead int64

	for totalRead < offset {
		toRead := int64(len(buffer))
		if totalRead+toRead > offset {
			toRead = offset - totalRead
		}

		n, err := file.Read(buffer[:toRead])
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		for i := 0; i < n; i++ {
			checksum ^= buffer[i]
		}
		totalRead += int64(n)
	}

	return checksum, nil
}

// showHexDump displays a hex dump of the specified file region
func showHexDump(filename string, offset int64, length int) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file for hex dump: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.Seek(offset, 0)
	if err != nil {
		fmt.Printf("Error seeking in file: %v\n", err)
		return
	}

	buffer := make([]byte, length)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fmt.Printf("Offset: 0x%08X\n", offset)
	for i := 0; i < n; i += 16 {
		fmt.Printf("%08X: ", offset+int64(i))
		
		// Hex values
		for j := 0; j < 16 && i+j < n; j++ {
			fmt.Printf("%02X ", buffer[i+j])
		}
		
		// Padding for incomplete lines
		for j := n - i; j < 16; j++ {
			fmt.Printf("   ")
		}
		
		fmt.Printf(" ")
		
		// ASCII representation
		for j := 0; j < 16 && i+j < n; j++ {
			b := buffer[i+j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("\n")
	}
}

// printResults displays the firmware analysis results
func printResults(info *FirmwareInfo) {
	fmt.Printf("ESP32-C6 Firmware Intrinsic Checksum Analysis\n")
	fmt.Printf("=============================================\n\n")

	fmt.Printf("File Size: %d bytes (%.2f KB)\n", info.TotalSize, float64(info.TotalSize)/1024)
	fmt.Printf("\n")

	// Header validation
	fmt.Printf("Header Validation:\n")
	fmt.Printf("  Magic Byte:    0x%02X %s\n", info.Header.Magic, 
		map[bool]string{true: "✓ Valid", false: "✗ Invalid"}[info.HeaderValid])
	fmt.Printf("  Chip ID:       0x%04X %s\n", info.Header.ChipID,
		map[bool]string{true: "✓ ESP32-C6", false: "✗ Not ESP32-C6"}[info.ChipIDValid])
	fmt.Printf("  Segments:      %d\n", info.Header.SegmentCount)
	fmt.Printf("  Entry Point:   0x%08X\n", info.Header.EntryAddr)
	fmt.Printf("  SPI Mode:      %d\n", info.Header.SpiMode)
	fmt.Printf("\n")

	// Segment information
	fmt.Printf("Segments:\n")
	for i, seg := range info.Segments {
		fmt.Printf("  Segment %d:     Load: 0x%08X, Size: %d bytes\n", 
			i, seg.LoadAddr, seg.DataLen)
	}
	fmt.Printf("\n")

	// Checksum analysis
	fmt.Printf("Checksum Analysis:\n")
	fmt.Printf("  Checksum Offset: 0x%08X (%d)\n", info.ChecksumOffset, info.ChecksumOffset)
	fmt.Printf("  Stored Checksum: 0x%02X\n", info.StoredChecksum)
	fmt.Printf("  ESP32 Checksum:  0x%02X\n", uint8(info.CalculatedCRC))
	fmt.Printf("  Status:          %s\n", 
		map[bool]string{true: "✓ Valid", false: "✗ Invalid"}[info.IsValid])

	if info.IsValid {
		fmt.Printf("\n✅ Firmware intrinsic checksum is VALID!\n")
	} else {
		fmt.Printf("\n❌ Firmware intrinsic checksum is INVALID!\n")
		fmt.Printf("   Expected: 0x%02X, Got: 0x%02X\n", 
			info.StoredChecksum, uint8(info.CalculatedCRC))
		
		// Additional diagnostic information
		if info.StoredChecksum == 0x00 {
			fmt.Printf("\n🔍 Diagnostic: Stored checksum is 0x00\n")
			fmt.Printf("   This might indicate:\n")
			fmt.Printf("   - Incomplete firmware image\n")
			fmt.Printf("   - Firmware without embedded checksum\n")
			fmt.Printf("   - Wrong checksum offset calculation\n")
		}
	}
}

func main() {
	var (
		firmwareFile = flag.String("file", "", "Path to the ESP32-C6 firmware file (required)")
		verbose      = flag.Bool("v", false, "Verbose output")
		hexdump      = flag.Bool("hex", false, "Show hex dump around checksum area")
		help         = flag.Bool("help", false, "Show usage information")
	)
	
	flag.Parse()

	if *help || *firmwareFile == "" {
		fmt.Printf("ESP32-C6 Firmware Intrinsic Checksum Verifier\n")
		fmt.Printf("Usage: %s -file <firmware.bin> [options]\n\n", os.Args[0])
		fmt.Printf("This tool verifies the built-in CRC checksum in ESP32-C6 firmware images.\n\n")
		fmt.Printf("Options:\n")
		flag.PrintDefaults()
		fmt.Printf("\nExamples:\n")
		fmt.Printf("  %s -file firmware.bin\n", os.Args[0])
		fmt.Printf("  %s -file bootloader.bin -v\n", os.Args[0])
		
		if *firmwareFile == "" {
			os.Exit(1)
		}
		return
	}

	// Check if file exists
	if _, err := os.Stat(*firmwareFile); os.IsNotExist(err) {
		fmt.Printf("Error: Firmware file '%s' does not exist\n", *firmwareFile)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Analyzing ESP32-C6 firmware: %s\n", *firmwareFile)
		fmt.Printf("Please wait...\n\n")
	}

	// Parse and verify firmware
	info, err := parseESP32Image(*firmwareFile)
	if err != nil {
		fmt.Printf("Error parsing firmware: %v\n", err)
		if info != nil {
			// Still show what we could parse
			printResults(info)
		}
		os.Exit(1)
	}

	// Print results
	printResults(info)
	
	// Show hex dump if requested
	if *hexdump {
		fmt.Printf("\nHex dump around checksum area:\n")
		showHexDump(*firmwareFile, info.ChecksumOffset-16, 32)
	}

	// Exit with error code if checksum is invalid
	if !info.IsValid {
		os.Exit(1)
	}
}
