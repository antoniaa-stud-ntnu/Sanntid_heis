package tcp

import (
	"fmt"
	"io"
	"net"
	"os"
)

var iPToConnMap map[net.Addr]net.Conn

// Master opening a listening server and saving+handling incomming connections from all the elevators
func TCPListenForConnectionsAndHandle(masterPort string, jsonMessageCh chan []byte) {
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
		// Add connection to map with active connections
		connectionsIP := conn.RemoteAddr()
		iPToConnMap[connectionsIP] = conn

		// Handle client connection in a goroutine
		go TCPReciveMessage(conn, jsonMessageCh)
	}
}

// Recieving messages and sending them on a channel for to be handeled else where
func TCPReciveMessage(conn net.Conn, jsonMessageCh chan<- []byte) { //Gjør om til at den mottar FSM-state
	defer conn.Close()

	// Create a buffer to read data into
	buffer := make([]byte, 65536)

	for {
		// Read data from the client
		data, err := conn.Read(buffer)
		if err != nil {
			// Remove the connection from iPToConnMap of active connections
			conn.Close()
			delete(iPToConnMap, conn.RemoteAddr())

			if err == io.EOF {
				fmt.Println("Client closed the connection.")
			} else {
				fmt.Println("Error:", err)
			}
			return
		}

		//fmt.Printf("Received: %s\n", buffer[:data])
		jsonMessageCh <- buffer[:data]
	}
}

// Function for connecting to the listening server
func TCPMakeMasterConnection(host string, port string) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host+":"+port)
	if err != nil {
		fmt.Printf("Could not resolve address: %s\n", err)
	}

	conn, err := net.Dial("tcp", tcpAddr.String())
	if err != nil {
		fmt.Println("Could not connect to server: ", err)
	}

	return conn, err
}

func TCPSendMessage(conn net.Conn, sendingData []byte) {
	// Send data to the other side of the conncetion
	_, err := conn.Write(sendingData)
	if err != nil {
		fmt.Println("Error sending data to server:", err)
		return
	}
}

// Forskjellen: TCPReciveMessage Den bruker en fast-størrelse buffer ([]byte) for å lese data fra tilkoblingen.
//  Dette betyr at den leser opptil 1024 byte om gangen, uavhengig av hvor mye data som faktisk er sendt av klienten per melding.
// Mens handleConnection leser og behandler en hel melding avsluttet med en ny linje (/n)

/*
func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		// Read from the connection untill a new line is send
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print the data read from the connection to the terminal
		fmt.Print("> ", string(data))

		// Write back the same message to the client
		conn.Write([]byte("Hello TCP Client\n"))
	}
}
*/
