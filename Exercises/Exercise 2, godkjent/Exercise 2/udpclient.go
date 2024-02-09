package main

import (
	"log"
	"net"
)

func main() {
	// Set up the server address
	serverAddr, err := net.ResolveUDPAddr("udp", "10.100.23.129:30000")
	if err != nil {
		log.Fatal(err)
	}

	// Create a UDP connection
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
