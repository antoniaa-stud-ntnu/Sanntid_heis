package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {

	var counter int = 0
	type PB_Mode int

	const (
		primary PB_Mode = 0
		backup  PB_Mode = 1
	)

	var mode PB_Mode = 0
	
	switch mode {
		case primary:
			primary_mode(counter)
		case backup:
			backup_mode(counter)
	}

}

func primary_mode(counter int) {
	// Primary mode
	fmt.Println("Primary mode started")
	// Start backup in a new terminal window
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "Ex4.go")
	fmt.Println("Åpnet ny backup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running a new backup:", err)
		return
	}

	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", "localhost:20006")
	if err != nil {
		log.Fatal(err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Send increasing counter messages
	for {
		_, err = conn.Write([]byte(strconv.Itoa(counter))) // Converting byte from int --> string --> byte
		if err != nil {
			log.Fatal(err)
		}
		counter++
		time.Sleep(50 * time.Millisecond)
		fmt.Println(counter)
	}
}

func backup_mode(counter int) {
	// Backup mode
	fmt.Println("Backup mode started")
	// Create a UDP connection for receiving
	recvAddr, err := net.ResolveUDPAddr("udp", "localhost:20006")
	if err != nil {
		log.Fatal(err)
	}

	recvConn, err := net.ListenUDP("udp", recvAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer recvConn.Close()

	fmt.Println("Backup listening on UDP")

	// Listen for responses until primary dies
	for {
		responseBuf := make([]byte, 1024)
		recvConn.SetReadDeadline(time.Now().Add(3 * time.Second)) // Set timeout
		n, addr, err := recvConn.ReadFromUDP(responseBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout: Server hasn't sent a message in 1 second")
				// Primary is dead
				// This program now becomes primary
				
				break

			}
			fmt.Println("Error reading from UDP:", err)
			return
		}
		fmt.Printf("Received message from %s: %s\n", addr, string(responseBuf[:n]))
		counter, _ = strconv.Atoi(string(responseBuf[:n]))
	}
	recvConn.Close()
}