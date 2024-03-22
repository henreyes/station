package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run listener.go <udp_port>")
		os.Exit(1)
	}

	udpPort, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid UDP port")
		os.Exit(1)
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", udpPort))
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Failed to listen on UDP port:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Listening on UDP port %d\n", udpPort)

	buffer := make([]byte, 1500)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		data := buffer[:n]
		os.Stdout.Write(data)
	}
}
