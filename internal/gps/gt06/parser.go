// Package gt06 encapsulates all GT06 GPS protocol parsing logic.
// SRP: sole responsibility is decoding/encoding GT06 binary packets.
package gt06

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Protocol message type constants.
const (
	ProtocolLogin     = byte(0x01)
	ProtocolHeartbeat = byte(0x13)
	ProtocolLocation  = byte(0x12)
)

// HeartbeatPacket holds the data extracted from a 0x13 packet.
type HeartbeatPacket struct {
	Battery  int
	Ignition bool
}

// LocationPacket holds all fields extracted from a 0x12 packet.
type LocationPacket struct {
	DeviceUnix int64
	Latitude   float64
	Longitude  float64
	Speed      int32
	Course     int32
	Satellites int
	Battery    int
	Ignition   bool
}

// ParseIMEI extracts the IMEI from a login packet (0x01).
// Returns the IMEI hex string and true on success.
func ParseIMEI(data []byte) (string, bool) {
	if len(data) < 12 {
		return "", false
	}
	return fmt.Sprintf("%x", data[4:12]), true
}

// ParseHeartbeat extracts battery and ignition from a heartbeat packet (0x13).
func ParseHeartbeat(data []byte) (HeartbeatPacket, bool) {
	if len(data) < 10 {
		return HeartbeatPacket{}, false
	}
	termInfo := data[4]
	voltLevel := data[5]
	return HeartbeatPacket{
		Battery:  calculateBatteryPct(int(voltLevel)),
		Ignition: (termInfo & 0x02) != 0,
	}, true
}

// ParseLocation decodes a location packet (0x12) using the GT06 protocol.
// lastBattery and lastIgnition carry over state from previous heartbeats.
func ParseLocation(data []byte, lastBattery int, lastIgnition bool) (LocationPacket, bool) {
	if len(data) < 22 {
		return LocationPacket{}, false
	}

	year, month, day := data[4], data[5], data[6]
	hour, minute, second := data[7], data[8], data[9]
	satellites := int(data[10] & 0x0F)

	latInt := binary.BigEndian.Uint32(data[11:15])
	latitude := float64(latInt) / 1800000.0

	lonInt := binary.BigEndian.Uint32(data[15:19])
	longitude := float64(lonInt) / 1800000.0

	speed := data[19]
	courseStatus := binary.BigEndian.Uint16(data[20:22])
	course := courseStatus & 0x03FF

	isNorth := (courseStatus & 0x0400) != 0
	isWest := (courseStatus & 0x0800) != 0

	if !isNorth {
		latitude = -latitude
	}
	if isWest {
		longitude = -longitude
	}

	// Infer ignition from speed if heartbeat hasn't set it
	if speed > 5 {
		lastIgnition = true
	}

	fullYear := 2000 + int(year)
	deviceUnix := time.Date(
		fullYear, time.Month(month), int(day),
		int(hour), int(minute), int(second),
		0, time.UTC,
	).Unix()

	return LocationPacket{
		DeviceUnix: deviceUnix,
		Latitude:   latitude,
		Longitude:  longitude,
		Speed:      int32(speed),
		Course:     int32(course),
		Satellites: satellites,
		Battery:    lastBattery,
		Ignition:   lastIgnition,
	}, true
}

// CreateAck builds the ACK packet for login (0x01) and heartbeat (0x13) packets.
func CreateAck(packet []byte) []byte {
	resp, _ := createGT06Response(packet)
	return resp
}

func createGT06Response(packet []byte) ([]byte, error) {
	if len(packet) < 10 {
		return nil, fmt.Errorf("paquete muy corto para ser GT06 válido")
	}
	serialIndex := len(packet) - 6
	if serialIndex < 4 {
		return nil, fmt.Errorf("longitud inválida para extraer serial")
	}

	serialBytes := packet[serialIndex : serialIndex+2]
	protocol := packet[3]

	respBody := []byte{0x05, protocol, serialBytes[0], serialBytes[1]}
	crc := checksumCRC16(respBody)

	finalPacket := []byte{0x78, 0x78}
	finalPacket = append(finalPacket, respBody...)

	crcBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(crcBytes, crc)
	finalPacket = append(finalPacket, crcBytes...)
	finalPacket = append(finalPacket, 0x0D, 0x0A)

	return finalPacket, nil
}

func checksumCRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}

func calculateBatteryPct(level int) int {
	switch level {
	case 0:
		return 0
	case 1:
		return 10
	case 2:
		return 20
	case 3:
		return 40
	case 4:
		return 60
	case 5:
		return 80
	case 6:
		return 100
	default:
		return 100
	}
}
