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
	"sync"
	"syscall"
	"time"

	"github.com/dhowden/tag"
)

type Metadata struct {
	Title  string
	Artist string
	Album  string
}

type Station struct {
	Name      string
	Filename  string
	Clients   map[*Client]struct{}
	Broadcast chan []byte
	MetaData  Metadata
	sync.Mutex
}

type Client struct {
	udpAddr *net.UDPAddr
	udpConn *net.UDPConn
	Station *Station
}

const (
	targetRate = 16 * 1024
	bufferSize = 1024
	sleep      = time.Duration(float64(bufferSize) / float64(targetRate) * float64(time.Second))
)

func ExtractMetadata(filePath string) (Metadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Metadata{}, err
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return Metadata{}, err
	}

	return Metadata{
		Title:  metadata.Title(),
		Artist: metadata.Artist(),
		Album:  metadata.Album(),
	}, nil
}

var stations []*Station

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
	initStations(os.Args[2:])

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
			if err == io.EOF {
				fmt.Println("Client closed the connection")
				return
			} else {
				log.Printf("Error reading message type: %v", err)
				return
			}
		}

		switch commandType {

		case protocol.HelloCommandType:

			var udpPort uint16
			err := binary.Read(conn, binary.BigEndian, &udpPort)
			if err != nil {
				log.Printf("Failed to read UDP port: %v", err)
				return
			}

			fmt.Println("recieved hello message from client, ", udpPort)

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

		case protocol.SetStationType:
			buffer := make([]byte, 4)
			_, err := io.ReadFull(conn, buffer)
			if err != nil {
				if err == io.EOF {
					log.Println("Client closed the connection while expecting SetStationType message")
					return
				} else {
					log.Printf("Error reading SetStation message: %v", err)
					return
				}
			}
			stationIndex, err := protocol.ParseSetStationMessage(buffer)
			if err != nil {
				log.Printf("Error parsing SetStation message: %v", err)
				return
			}

			if stationIndex < 0 || int(stationIndex) >= len(stations) {
				log.Printf("Invalid station index received: %d", stationIndex)
				return
			}

			selectedStation := stations[stationIndex]

			if client.Station != nil {
				client.Station.Lock()
				delete(client.Station.Clients, &client)
				client.Station.Unlock()
			}
			client.Station = selectedStation
			selectedStation.Lock()
			selectedStation.Clients[&client] = struct{}{}
			selectedStation.Unlock()

			confirmMsg := protocol.SetStationMessage(stationIndex)
			_, err = conn.Write(confirmMsg)
			if err != nil {
				log.Printf("Failed to send station changed confirmation: %v", err)
				return
			}

		}

	}

}

func initStations(filenames []string) {
	for _, filename := range filenames {
		md, err := ExtractMetadata(filename)
		if err != nil {
			fmt.Println("failed to extract metadata")
			continue
		}
		fmt.Println("[METADATA]:", md)
		station := &Station{
			Name:      filepath.Base(filename),
			Filename:  filename,
			Clients:   make(map[*Client]struct{}),
			Broadcast: make(chan []byte, bufferSize),
			MetaData:  md,
		}
		stations = append(stations, station)
		go station.startBroadcast()
	}
}

func (s *Station) startBroadcast() {
	file, err := os.Open(s.Filename)
	if err != nil {
		log.Printf("Failed to open file %s: %v", s.Filename, err)
		return
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				file.Seek(0, 0)
				continue
			}
			log.Printf("Failed to read file %s: %v", s.Filename, err)
			return
		}

		s.Lock()
		for client := range s.Clients {
			if _, err := client.udpConn.Write(buffer[:bytesRead]); err != nil {
				log.Printf("Failed to send to client: %v", err)
				delete(s.Clients, client)
			}
			fmt.Println("sending data to listener")
		}
		s.Unlock()
		time.Sleep(sleep)
	}
}
