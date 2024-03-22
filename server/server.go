package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"station/protocol"
	"strconv"
	"syscall"
)

type Station struct {
	Name     string
	Filename string
	Clients  map[string]int
}

type Client struct {
	udpAddr *net.UDPAddr
	udpConn *net.UDPConn
}

var stations []Station

func main() {

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: <listen_port>")
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

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		fmt.Printf("Failed to create listener: %v\n", err)
		return
	}
	defer listener.Close()

	for _, filename := range args[2:] {
		stations = append(stations, Station{
			Name:     filepath.Base(filename),
			Filename: filename,
			Clients:  make(map[string]int),
		})
	}

	fmt.Printf("Server listening on port %d\n", port)

	ctrlCChan := make(chan os.Signal, 1)
	signal.Notify(ctrlCChan, os.Interrupt, syscall.SIGINT)

	go waitConnections(listener)

	<-ctrlCChan
	fmt.Println("ctrl+c found, closing client connections")

}

func waitConnections(listenConn *net.TCPListener) {
	for {
		conn, err := listenConn.Accept()
		if err != nil {
			log.Fatalln("accept: ", err)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	var client Client
	for {
		var commandType uint8
		err := binary.Read(conn, binary.BigEndian, &commandType)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading command type: %v", err)
			}
			break
		}

		switch commandType {
		case protocol.HelloCommandType:

			var udpPort uint16
			err := binary.Read(conn, binary.BigEndian, &udpPort)
			if err != nil {
				log.Printf("Failed to read UDP port: %v", err)
				return
			}
			station := stations[0]

			fmt.Println("recieved message from client, ", udpPort)

			client.udpAddr = &net.UDPAddr{
				IP:   conn.RemoteAddr().(*net.TCPAddr).IP,
				Port: int(udpPort),
			}

			client.udpConn, err = net.DialUDP("udp", nil, client.udpAddr)
			if err != nil {
				log.Printf("Error setting up UDP connection: %v", err)
				return
			}
			defer client.udpConn.Close()

			welcomeMsg := protocol.WelcomeMessage(uint16(len(stations)))
			_, err = conn.Write(welcomeMsg)
			if err != nil {
				log.Printf("Failed to send Welcome message: %v", err)
				return
			}
			go streamAudio(&client, station.Filename)
		}
	}

}

func streamAudio(client *Client, filename string) {

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Error opening audio file: %v", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		fmt.Println("sending data to listener")

		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {

				_, err = file.Seek(0, 0)
				if err != nil {
					log.Printf("Error seeking audio file: %v", err)
					return
				}
				continue
			}
			log.Printf("Error reading audio file: %v", err)
			return
		}

		_, err = client.udpConn.Write(buffer[:bytesRead])
		if err != nil {
			log.Printf("Error sending UDP packet: %v", err)
			return
		}

	}
}
