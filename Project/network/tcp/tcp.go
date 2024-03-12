package tcp

import (
	"Project/network/messages"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type ExistingIPsAndConn struct {
	ExistingIP string
	Conn       net.Conn
}

// Master opening a listening server and saving+handling incomming connections from all the elevators
func TCPListenForConnectionsAndHandle(masterPort string, jsonMessageCh chan []byte, iPToConnMap *map[string]net.Conn, mutex *sync.Mutex, allHallReqAndStates messages.HRAInput, existingIPsAndConnCh chan ExistingIPsAndConn, quitCh chan bool) {
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

		mutex.Lock()
		(*iPToConnMap)[connectionsIP] = conn
		mutex.Unlock()
		fmt.Printf("iPToConnMap is updated to: IP-%s, Conn-%d", connectionsIP, conn)

		// Handle client connection in a goroutine
		go TCPRecieveMessage(conn, jsonMessageCh, quitCh)
	}
}

func TCPLookForClosedConns(iPToConnMap *map[string]net.Conn, mutexIPConn *sync.Mutex) {
	for ip, conn := range *iPToConnMap {

		_, err := conn.Read(make([]byte, 1024))
		if err != nil {

			mutexIPConn.Lock()
			fmt.Println("Deleting a conn: ", (*iPToConnMap))
			delete((*iPToConnMap), ip)
			mutexIPConn.Unlock()
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
	// Create a buffer to read data into

	buffer := make([]byte, 65536)

	for {
		select {
		case <- quitCh: // Stopping goroutine
			return
		default:
			// Read data from the client
			data, err := conn.Read(buffer)
			if err != nil {
				// Remove the connection from iPToConnMap of active connections
				conn.Close()
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

// Panic pga sende til nil pointer -> kan fikses ved å søke på nettet på GO panic ignore......
// While loop (eller noe sånt) på TCPMakeMasterConn
// TCPRecieveMessage skal avslutte en eller annen gang, channelen skal bli født
// Timer kan vi ha på motorstopp, visst den bruker
// Ta bort ALLE mutexer :(

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

// Forskjellen: TCPRecieveMessage Den bruker en fast-størrelse buffer ([]byte) for å lese data fra tilkoblingen.
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
