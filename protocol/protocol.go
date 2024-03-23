package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	HelloCommandType   = 0
	WelcomeReplyType   = 2
	ChangeStationType  = 3
	StationChangedType = 4
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

func ParseStationChangedMessage(data []byte) (stationIndex int32, err error) {
	buf := bytes.NewBuffer(data)
	commandType, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}
	if commandType != StationChangedType {
		return 0, errors.New("incorrect message type for StationChanged")
	}
	err = binary.Read(buf, binary.BigEndian, &stationIndex)
	return stationIndex, err
}

func StationChangedMessage(stationIndex int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(StationChangedType))
	binary.Write(buf, binary.BigEndian, stationIndex)
	return buf.Bytes()
}
