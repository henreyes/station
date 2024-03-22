package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	HelloCommandType = 0
	WelcomeReplyType = 2
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
