package main

import (
	"fmt"
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

	welcomeBuf := make([]byte, 3)
	n, err := conn.Read(welcomeBuf)
	if err != nil {
		fmt.Printf("Failed to read Welcome message: %v\n", err)
		os.Exit(1)
	}
	if n != len(welcomeBuf) {
		fmt.Println("Received incomplete Welcome message.")
		os.Exit(1)
	}

	numStations, err := protocol.ParseWelcomeMessage(welcomeBuf)
	if err != nil {
		fmt.Printf("Failed to parse Welcome message: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Received Welcome message with %d stations\n", numStations)

}
