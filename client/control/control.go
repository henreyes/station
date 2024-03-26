package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"station/protocol"
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

	helloMsg := protocol.HelloMessage(uint16(udpPort))
	_, err = conn.Write(helloMsg)
	if err != nil {
		fmt.Println("Failed to send Hello message:", err)
		os.Exit(1)
	}

	go func() {
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
				fmt.Printf("Welcome to the radio. There are %d stations\n", numStations)
			case protocol.MetadataType:
				metadata, err := protocol.ParseMetadataMessage(conn)
				if err != nil {
					fmt.Printf("Error parsing metadata: %v\n", err)
					continue
				}

				fmt.Printf("Now playing: '%s' \n",
					metadata.Title)
			default:
				fmt.Printf("Unknown message type received: %d\n", messageType)

			}
		}

	}()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()
		args := strings.Split(input, " ")

		if len(args) == 0 {
			continue
		}

		switch args[0] {

		case "set":
			if len(args) < 3 || args[1] != "station" {
				fmt.Println("Invalid command. Usage: 'set station <index>'")
				continue
			}

			stationIndex, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Printf("Invalid station index: %v\n", err)
				continue
			}
			fmt.Printf("CLIENT INPUT:[%d]", stationIndex)
			setStationMsg := protocol.SetStationMessage(int32(stationIndex))
			_, err = conn.Write(setStationMsg)
			if err != nil {
				fmt.Printf("Failed to send SetStation message: %v\n", err)
				continue
			}

		default:
			fmt.Println("Unknown command:", args[0])
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}
