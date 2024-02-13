package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

func main() {

	 counter := 0

	// Backup mode
	fmt.Println("Backup mode started")

	// Create a UDP connection for receiving
	addr, err := net.ResolveUDPAddr("udp4", "localhost:20006")
	if err != nil {
		fmt.Println("Failed to create a UDP connection for receiving")
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Failed to ListenUDP")
	}
	defer conn.Close()

	fmt.Println("Backup listening on UDP")
	// Listen for responses until primary dies
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // Set timeout
		responseBuf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(responseBuf)
		if err != nil {
			fmt.Printf("Error reading from UDP: %d\n", err)
			break
			
		}
		fmt.Printf("Received message from %s: %s\n", addr, string(responseBuf[:n]))
		counter, err = strconv.Atoi(string(responseBuf[:n]))
		if err != nil {
			fmt.Printf("Error converting to int: %d\n", err)
		}
	}
	conn.Close()

	// Primary mode
	fmt.Println("Primary mode started")
	// Start backup in a new terminal window
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "Ex4_new.go")
	fmt.Println("Åpnet ny backup")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr

	cmd.Run()


	// Create UDP connection
	conn, err = net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial")
	}
	//defer conn.Close()

	time.Sleep(1 * time.Second)
	// Send increasing counter messages
	for {
		counter++
		v := strconv.Itoa(counter)
		_, err := conn.Write([]byte(v)) // Converting byte from int --> string --> byte
		if err != nil {
			fmt.Println("Failed to send")
		} else {
			fmt.Println(v)
		}
		time.Sleep(2 * time.Second)
	}

}
