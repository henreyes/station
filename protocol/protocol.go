package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Metadata struct {
	Title  string
	Artist string
	Album  string
}

const (
	HelloCommandType   = 0
	WelcomeReplyType   = 2
	SetStationType     = 3
	StationChangedType = 4
	InvalidCommandType = 5
	MetadataType       = 6
)

func HelloMessage(udpPort uint16) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, uint8(HelloCommandType))
	binary.Write(buf, binary.BigEndian, udpPort)
	return buf.Bytes()
}

func ParseWelcomeMessage(data []byte) (numStations uint16, err error) {
	buf := bytes.NewBuffer(data)

	_, err = buf.ReadByte()
	if err != nil {
		return 0, err
	}

	err = binary.Read(buf, binary.BigEndian, &numStations)
	return numStations, err
}

func WelcomeMessage(numStations uint16) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, uint8(WelcomeReplyType))
	binary.Write(buf, binary.BigEndian, numStations)
	return buf.Bytes()
}

func ParseSetStationMessage(data []byte) (stationIndex int32, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.BigEndian, &stationIndex)
	return stationIndex, err
}

func SetStationMessage(stationIndex int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(SetStationType))
	binary.Write(buf, binary.BigEndian, stationIndex)
	fmt.Printf("sending station message: [%d]", buf.Bytes())
	return buf.Bytes()
}

func MetadataMessage(title, artist, album string) []byte {
	buf := new(bytes.Buffer)

	// First, write the message type
	binary.Write(buf, binary.BigEndian, uint8(MetadataType))

	writeLengthPrefixedString(buf, title)
	writeLengthPrefixedString(buf, artist)
	writeLengthPrefixedString(buf, album)

	return buf.Bytes()
}

func writeLengthPrefixedString(buf *bytes.Buffer, s string) {
	// Convert the string to a byte slice
	strBytes := []byte(s)
	// Write the length of the string as a uint16
	binary.Write(buf, binary.BigEndian, uint16(len(strBytes)))
	// Write the string bytes
	buf.Write(strBytes)
}

func InvalidCommandMessage(reason string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(InvalidCommandType))
	binary.Write(buf, binary.BigEndian, []byte(reason))
	return buf.Bytes()
}

func ParseMetadataMessage(conn net.Conn) (Metadata, error) {
	var metadata Metadata
	// Read the title length and title
	var titleLength uint16
	if err := binary.Read(conn, binary.BigEndian, &titleLength); err != nil {
		return metadata, err
	}
	title := make([]byte, titleLength)
	if _, err := io.ReadFull(conn, title); err != nil {
		return metadata, err
	}
	metadata.Title = string(title)

	var artistLength uint16
	if err := binary.Read(conn, binary.BigEndian, &artistLength); err != nil {
		return metadata, err
	}
	artist := make([]byte, titleLength)
	if _, err := io.ReadFull(conn, artist); err != nil {
		return metadata, err
	}
	metadata.Artist = string(artist)

	var albumLength uint16
	if err := binary.Read(conn, binary.BigEndian, &albumLength); err != nil {
		return metadata, err
	}
	album := make([]byte, albumLength)
	if _, err := io.ReadFull(conn, artist); err != nil {
		return metadata, err
	}
	metadata.Album = string(album)
	return metadata, nil
}
