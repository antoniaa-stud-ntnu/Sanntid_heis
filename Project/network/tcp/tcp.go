package tcp

import (
	"Project/network/messages"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type ExistingIPsAndConn struct {
	ExistingIP string
	Conn       net.Conn
}

// Master opening a listening server and saving+handling incomming connections from all the elevators
func TCPListenForConnectionsAndHandle(masterPort string, jsonMessageCh chan []byte, iPToConnMap *map[string]net.Conn, allHallReqAndStates messages.HRAInput, existingIPsAndConnCh chan ExistingIPsAndConn, quitCh chan bool) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+masterPort)
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

	fmt.Printf("Server is listening on port %s\n", masterPort)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Could not accept connection: %s\n", err)
			continue
		}
		//fmt.Println("Master accepted new connection: ", conn)

		// Add connection to map with active connections
		connectionsIP := ((conn.RemoteAddr().(*net.TCPAddr)).IP).String()
		fmt.Printf("Master accepted new connection-%d from IP-%s \n", conn, connectionsIP)

		// Hvis conn ikke allerede eksisterer i allHallReqAndStates fra roleFSM
		if _, exists := allHallReqAndStates.States[connectionsIP]; exists {
			existingIPsAndConnCh <- ExistingIPsAndConn{connectionsIP, conn}
			// I master, gå gjennom allHallReqAndStates med denne IP og sender tcp med cabrequests til ipadressen
		}

		(*iPToConnMap)[connectionsIP] = conn
		fmt.Printf("iPToConnMap is updated to: IP-%s, Conn-%d", connectionsIP, conn)

		// Handle client connection in a goroutine
		go TCPRecieveMessage(conn, jsonMessageCh, quitCh)
	}
}

func TCPLookForClosedConns(iPToConnMap *map[string]net.Conn) {
	for ip, conn := range *iPToConnMap {

		_, err := conn.Read(make([]byte, 1024))
		if err != nil {

			fmt.Println("Deleting a conn: ", (*iPToConnMap))
			delete((*iPToConnMap), ip)
		}
	}
}

// Recieving messages and sending them on a channel for to be handeled else where

// Function for connecting to the listening server
func TCPMakeMasterConnection(host string, port string) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host+":"+port)
	if err != nil {
		fmt.Printf("Could not resolve address: %s\n", err)
	}

	var conn net.Conn
	
	// Looping until connection is established
	for {
		conn, err = net.Dial("tcp", tcpAddr.String())
		if err != nil {
			fmt.Println("Could not connect to server: ", err)
			time.Sleep(50*time.Millisecond)
		}else {
			break
		}
	}
	return conn, err
}

func TCPRecieveMessage(conn net.Conn, jsonMessageCh chan<- []byte, quitCh <-chan bool) {
	defer conn.Close()
	//fmt.Println("In TCP recieve message")

	buffer := make([]byte, 65536) // Create a buffer to read data into

	for {
		select {
		case <- quitCh: // Stopping goroutine
			return
		default:
			data, err := conn.Read(buffer) // Read data from the client
			if err != nil {
				conn.Close() // Remove the connection from iPToConnMap of active connections
				if err == io.EOF {
					fmt.Println("Client closed the connection.")
				} else {
					fmt.Println("Error:", err)
				}
				return
			}

		msg := make([]byte, data)
		copy(msg, buffer[:data])
		jsonMessageCh <- msg
		}
	}
}

func TCPSendMessage(conn net.Conn, sendingData []byte) {
	// Send data to the other side of the connection
	_, err := conn.Write(sendingData)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Connection closed by the server.")
		} else {
			fmt.Println("Error sending data to server:", err)
		}
		return
	}
}