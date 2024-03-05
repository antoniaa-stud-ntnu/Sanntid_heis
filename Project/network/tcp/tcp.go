package tcp

import (
	"fmt"
	"net"
	"os"
)

func TCP_server(hostIP string, hostPort string. handleClient func) {

	// Listen for incoming connections
	//host := "localhost"
	//port := "8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", hostIP+":"+hostPort)

	if err != nil {
		fmt.Printf("Could not resolve address: %s\n", err)
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Could not open listener: ", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server is listening on port %s\n", hostPort)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		// conn.SetReadDeadline(time.Now().Add(1 * time.Second)) // Set timeout of 1 second
		if err != nil {
			fmt.Printf("Could not accept connection: %s\n", err)
			continue
		}

		// Handle client connection in a goroutine
		go handleClient(conn)
	}
}

func TCP_client(sendingData string, host string, port string) { 
    // Connect to the server
	// host := "localhost"
	// port := "8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host+":"+port)
	if err != nil {
		fmt.Printf("Could not resolve address: %s\n", err)
		os.Exit(1)
	}
    conn, err := net.Dial("tcp", tcpAddr.String())
    if err != nil {
        fmt.Println("Could not connect to server: ", err)
        return
    }
    defer conn.Close()

    // Send data to the server
	data := []byte(sendingData)
	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("Could not send data: %s\n", err)
		return
	}

    // Read and process data from the server
    // ...
}

func handleClient(conn net.Conn) { //Gjør om til at den mottar FSM-state
	defer conn.Close()

	// Create a buffer to read data into
	buffer := make([]byte, 1024)

	for {
		// Read data from the client
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Process and use the data (here, we'll just print it)
		fmt.Printf("Received: %s\n", buffer[:n])
	}
}


