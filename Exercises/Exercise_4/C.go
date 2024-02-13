package main

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

func sendMessage(j int, conn *net.UDPConn) {
	s := strconv.Itoa(j)
	_, err := conn.Write([]byte(s))
	if err != nil {
		fmt.Println("Failed to send")
	} else {
		//fmt.Printf("Sent message: %s\n", s)
		fmt.Println(s)
	}
}

func main() {
	i := 0
	addr, err := net.ResolveUDPAddr("udp4", "localhost:20006")
	if err != nil {
		fmt.Println("Unable to resolve")
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Unable to listen")
	}
	defer conn.Close()

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			//fmt.Println("Died")
			break
		}

		i, err = strconv.Atoi(string(buf[:n]))
		if err != nil {
			fmt.Println("Unable to convert to int")
		}
		//fmt.Printf("Recieved Message: %s\n", buf[:])
	}
	conn.Close()
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "C.go")
	cmd.Run()
	conn, err = net.DialUDP("udp4", nil, addr)
	if err != nil {
		fmt.Println("Failed to dial")
	}
	time.Sleep(1 * time.Second)
	for {
		i++
		sendMessage(i, conn)

		time.Sleep(2 * time.Second)
	}
}