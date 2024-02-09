package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <port>")
		os.Exit(1)
	}

	port := os.Args[1]
	serverAddr := "10.100.23.129" // Replace with the actual server IP address

	conn, err := connectToServer(serverAddr, port)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to the server!")

	// Receive the welcome message
	welcomeMessage, err := receiveMessage(conn)
	if err != nil {
		fmt.Println("Error receiving welcome message:", err)
		os.Exit(1)
	}
	fmt.Println("Welcome message:", welcomeMessage)

	// Send a constant message
	constantMessage := "Connect to: 10.100.23.16:20006" + "\x00"

	_, err = conn.Write([]byte(constantMessage))
	if err != nil {
		fmt.Println("Error sending constant message:", err)
		os.Exit(1)
	}
	fmt.Println("Message sent:", constantMessage)

	// Receive and print the echoed message
	echoedMessage, err := receiveMessage(conn)
	if err != nil {
		fmt.Println("Error receiving echoed message:", err)
		os.Exit(1)
	}
	fmt.Println("Echoed message:", echoedMessage)
}

func connectToServer(serverAddr, port string) (net.Conn, error) {
	addr := fmt.Sprintf("%s:%s", serverAddr, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func receiveMessage(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)

	// Read until the first '\0' is encountered
	message, err := reader.ReadString('\x00')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(message, "\x00"), nil
}
