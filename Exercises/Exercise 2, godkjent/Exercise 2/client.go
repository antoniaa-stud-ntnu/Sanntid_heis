package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	port := "33546"
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
	constantMessage := "Connect to: 10.100.23.32:20022" + "\x00" //Husk å endre ip-adresse og port

	_, err = conn.Write([]byte(constantMessage))
	if err != nil {
		fmt.Println("Error sending constant message:", err)
		os.Exit(1)
	}
	fmt.Println("Message sent:", constantMessage)

	// Resolve the string address to a TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "10.100.23.32:20022") //Husk å endre ip-adresse og port

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start listening for TCP connections on the given address
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}


	// Accept new connections
	accept_conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
	}
	
	// Handle new connections in a Goroutine for concurrency
	defer accept_conn.Close()

	for {
		// Read from the connection untill a new line is send
		data, err := bufio.NewReader(accept_conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

	// Print the data read from the connection to the terminal
	fmt.Print("> ", string(data))

	// Write back the same message to the client
	accept_conn.Write([]byte("Hello TCP Client\n"))
	}


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
