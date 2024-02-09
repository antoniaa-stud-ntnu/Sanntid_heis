package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
	"strconv"
)

func main() {
	fmt.Println("Primary started")

	startBackup()

	// Infinit loop of counting and sending
	var counter int = 0
	for {
		sendMsg(strconv.Itoa(counter))
		log.Println(strconv.Itoa(counter))
		counter++
	}

}

func startBackup() {

	//Execute Backup in a new terminal window
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "A_B/program_A.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running Program A:", err)
		return
	}
}

func sendMsg(msgToSend string) {
	// Set up the udp address
	addr, err := net.ResolveUDPAddr("udp", "localhost:20014")
	if err != nil {
		log.Fatal(err)
	}

	// Create a UDP connection for sending locally
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Message to send
	sendingMessage := []byte(msgToSend)

	// Send the broadcast message
	_, err = conn.Write(sendingMessage)
	if err != nil {
		log.Fatal(err)
	}

	//log.Println("Message sent successfully.")
}

