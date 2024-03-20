package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage:  <listen_port>")
		return
	}
	port, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid port number")
		return
	}

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("Failed to resolve TCP address: %v\n", err)
		return
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to create listener: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server listening on port %d\n", port)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		// Handle the connection in a new goroutine
		go handleClient(conn, args[2:])
	}
}

func handleClient(conn net.Conn, args []string) {
	defer conn.Close()
	numStations := uint16(len(args) - 2)

	buf := make([]byte, 3)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Failed to read from client: %v\n", err)
		return
	}

	welcome := []byte{2, byte(numStations >> 8), byte(numStations & 0xff)}
	_, err = conn.Write(welcome)
	if err != nil {
		fmt.Printf("Failed to write to client: %v\n", err)
		return
	}

	fmt.Printf("Client connected! The server has %d stations.\n", numStations)
}
