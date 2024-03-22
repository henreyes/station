package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

	err = binary.Write(conn, binary.BigEndian, uint16(udpPort))
	if err != nil {
		fmt.Println("Failed to send Hello command:", err)
		os.Exit(1)
	}

	var numStations uint16
	err = binary.Read(conn, binary.BigEndian, &numStations)
	if err != nil {
		fmt.Println("Failed to read Welcome reply:", err)
		os.Exit(1)
	}

	fmt.Printf("Welcome to the server! The server has %d stations.\n", numStations)

	go receiveReplies(conn)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter a command (q to quit, number to set station): ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		input = strings.TrimSpace(input)

		if input == "q" {
			break
		}

		stationNumber, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid station number")
			continue
		}

		err = binary.Write(conn, binary.BigEndian, uint16(stationNumber))
		if err != nil {
			fmt.Println("Failed to send SetStation command:", err)
			break
		}
	}
}

func receiveReplies(conn net.Conn) {
	for {
		var replyType uint8
		err := binary.Read(conn, binary.BigEndian, &replyType)
		if err != nil {
			fmt.Println("Failed to read reply type:", err)
			break
		}

		switch replyType {
		case 3: // Announce
			var songNameSize uint8
			err := binary.Read(conn, binary.BigEndian, &songNameSize)
			if err != nil {
				fmt.Println("Failed to read song name size:", err)
				break
			}

			songName := make([]byte, songNameSize)
			_, err = conn.Read(songName)
			if err != nil {
				fmt.Println("Failed to read song name:", err)
				break
			}

			fmt.Printf("New song announced: %s\n", string(songName))

		case 4:
			var replyStringSize uint8
			err := binary.Read(conn, binary.BigEndian, &replyStringSize)
			if err != nil {
				fmt.Println("Failed to read reply string size:", err)
				break
			}

			replyString := make([]byte, replyStringSize)
			_, err = conn.Read(replyString)
			if err != nil {
				fmt.Println("Failed to read reply string:", err)
				break
			}

			fmt.Printf("Invalid command: %s\n", string(replyString))
			os.Exit(1)

		default:
			fmt.Printf("Unknown reply type: %d\n", replyType)
			os.Exit(1)
		}
	}
}
