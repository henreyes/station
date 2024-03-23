package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	HelloCommandType   = 0
	WelcomeReplyType   = 2
	SetStationType     = 3
	StationChangedType = 4
	InvalidCommandType = 5
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

func InvalidCommandMessage(reason string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(InvalidCommandType))
	binary.Write(buf, binary.BigEndian, []byte(reason))
	return buf.Bytes()
}
