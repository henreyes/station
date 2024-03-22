package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"station/protocol"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run control.go <server_name> <server_port> <udp_port>")
		os.Exit(1)
	}

	serverName := os.Args[1]
	serverPort := os.Args[2]
	udpPort, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Invalid UDP port")
		os.Exit(1)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", serverName, serverPort))
	if err != nil {
		fmt.Println("Failed to connect to the server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	helloMsg := protocol.HelloMessage(uint16(udpPort))
	_, err = conn.Write(helloMsg)
	if err != nil {
		fmt.Println("Failed to send Hello message:", err)
		os.Exit(1)
	}

	for {
		var messageType uint8
		err := binary.Read(conn, binary.BigEndian, &messageType)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading message type: %v\n", err)
			}
			return
		}

		switch messageType {
		case protocol.WelcomeReplyType:

			var numStations uint16
			err := binary.Read(conn, binary.BigEndian, &numStations)
			if err != nil {
				fmt.Printf("Error reading Welcome message data: %v\n", err)
				break
			}
			fmt.Printf("Received Welcome message with %d stations\n", numStations)
		default:
			fmt.Printf("Unknown message type received: %d\n", messageType)

		}
	}

}
