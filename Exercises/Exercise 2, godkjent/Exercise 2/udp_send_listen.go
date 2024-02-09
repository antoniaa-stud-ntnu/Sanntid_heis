package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	serverBroadcastPort = 30000
	basePort            = 20006
)

func main() {
	// Use separate goroutines for sending and receiving
	go sendBroadcast()
	go receiveMessage()

	// Sleep to allow time for goroutines to run
	time.Sleep(10 * time.Second)
}

func sendBroadcast() {
	// Set up the broadcast address
	broadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:30000")
	if err != nil {
		log.Fatal(err)
	}

	// Create a UDP connection for sending broadcast
	broadcastConn, err := net.DialUDP("udp", nil, broadcastAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer broadcastConn.Close()

	// Message to broadcast
	broadcastMessage := []byte("Hello from UDP server at 10.100.23.129!")

	// Send the broadcast message
	_, err = broadcastConn.Write(broadcastMessage)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Broadcast message sent successfully.")
}

func receiveMessage() {
	// Set up the address for receiving
	recvAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", basePort+1))
	if err != nil {
		log.Fatal(err)
	}

	// Create a UDP connection for receiving
	conn, err := net.ListenUDP("udp", recvAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Receive server IP broadcast
	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Fatal(err)
	}

	serverIP := string(buf[:n])
	fmt.Printf("Received server IP: %s\n", serverIP)

	// Use the received server IP for subsequent communication
	go sendMessage(serverIP)
}

func sendMessage(serverIP string) {
	// Set up the server address for sending
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIP, basePort))
	if err != nil {
		log.Fatal(err)
	}

	// Create a UDP connection for sending
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Message to send
	message := []byte("Hello, UDP Server!")

	// Send the message
	_, err = conn.Write(message)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Message sent successfully.")
}
