package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	serverPort = "localhost"
	basePort   = 20014
)

func main() {
	fmt.Println("Backup started")

	var counter int

	// Create a UDP connection for receiving
	recvAddr, err := net.ResolveUDPAddr("udp", "localhost:20014")
	if err != nil {
		log.Fatal(err)
	}

	recvConn, err := net.ListenUDP("udp", recvAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer recvConn.Close()

	// Listen for responses
	responseBuf := make([]byte, 1024)
	recvConn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Set timeout
	n, addr, err := recvConn.ReadFromUDP(responseBuf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("Timeout: Server hasn't sent a message in 2 seconds")
			// This program now becomes primart
			// Start a new backup
			return
		}
		fmt.Println("Error reading from UDP:", err)
		return
	}
	fmt.Printf("Received message from %s: %s\n", addr, string(responseBuf[:n]))

	time.Sleep(time.Second * 5)
	fmt.Println("Program A terminated")

}
