package main

import (
	"log"
	"net"
)

const serverAddress = "10.100.23.129"
const basePort = 20006

func main() {
	// Set up the server address
	serverAddr, err := net.ResolveUDPAddr("udp", ":20006")
	if err != nil {
		log.Fatal(err)
	}

	// Create a UDP connection for sending
	sendConn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer sendConn.Close()

	// Message to send
	message := []byte("Hello, UDP Server!")

	// Send the message
	_, err = sendConn.Write(message)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Message sent successfully.")

	// Create a UDP connection for receiving
	recvAddr, err := net.ResolveUDPAddr("udp", ":20006")
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
	for {
		n, _, err := recvConn.ReadFromUDP(responseBuf)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Received response from server: %s\n", responseBuf[:n])
	}
}

/*package main

import (
	"log"
	"net"
)

const serverAddress = "10.100.23.129"
const basePort = 20006

func main() {
	// Set up the server address
	serverAddr, err := net.ResolveUDPAddr("udp", ":20006")
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



	// Listen for responses
	responseBuf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(responseBuf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Still waiting")
		log.Printf("Received response from server: %s\n", responseBuf[:n])
	}
	
}


package main

import (
	"log"
	"net"
)

const serverAddress = "10.100.23.129"
const basePort = 20006

func main() {
	// Set up the server address
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddress+":20006")
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
	message := []byte("Hello, UDP Server from group 6!")

	// Send the message
	_, err = conn.Write(message)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Message sent successfully.")

	// Listen for responses
	responseBuf := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(responseBuf)
	if err != nil {
		log.Fatal(err)
	}

	// Extract remote IP and port
	remoteIP := remoteAddr.IP.String()
	remotePort := remoteAddr.Port

	log.Printf("Received response from server: %s\n", responseBuf[:n])
	log.Printf("Response from: %s:%d\n", remoteIP, remotePort)
}
*/