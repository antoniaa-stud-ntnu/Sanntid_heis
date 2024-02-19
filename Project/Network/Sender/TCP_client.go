package TCP_client

import (
	"fmt"
	"net"
	"os"
)

func TCP_client(sendingData string) { 
    // Connect to the server
	host := "localhost"
	port := "8080"
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