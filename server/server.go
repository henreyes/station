package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

type Station struct {
	Name     string
	Filename string
	Clients  map[string]int
}

var stations []Station

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
	fmt.Println("ctrl+c found, closing client connections...")

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

	var udpPort uint16
	err := binary.Read(conn, binary.BigEndian, &udpPort)
	if err != nil {
		log.Printf("Failed to read UDP port: %v", err)
		return
	}
	fmt.Printf("Received UDP port: %d\n", udpPort)

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Failed to read message: %v", err)
		return
	}

	message := string(buffer[:n])
	fmt.Printf("Received message: %s\n", message)

}
