package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run client.go <server_address>")
		return
	}
	serverAddress := os.Args[1]

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("Failed to connect to the server: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Connected to the server at %s\n", serverAddress)

	hello := []byte{0, 0, 0}
	_, err = conn.Write(hello)
	if err != nil {
		fmt.Printf("Failed to send Hello message: %v\n", err)
		return
	}

	welcome := make([]byte, 3)
	_, err = conn.Read(welcome)
	if err != nil {
		fmt.Printf("Failed to receive Welcome message: %v\n", err)
		return
	}

	numStations := uint16(welcome[1])<<8 | uint16(welcome[2])
	fmt.Printf("Received Welcome message. The server has %d stations.\n", numStations)
}
