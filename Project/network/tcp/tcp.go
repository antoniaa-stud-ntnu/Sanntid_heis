package tcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var iPToConnMap map[net.Addr]net.Conn


func TCPListenForConnections(hostPort string, jsonMessageCh chan []byte) { //For master to listen to incomming connections, hostPort=masterPort
	tcpAddr, err := net.ResolveTCPAddr("tcp4", hostPort)
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
		connectionsIP := conn.RemoteAddr()
		iPToConnMap[connectionsIP] = conn
		go TCPReciveMessage(conn, jsonMessageCh)

	}
}

func TCPMakeConncetion(host string, port string) net.Conn{ //For dummies og backup
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host+":"+port)
	if err != nil {
		fmt.Printf("Could not resolve address: %s\n", err)
		os.Exit(1)
	}
	conn, err := net.Dial("tcp", tcpAddr.String())
	if err != nil {
		fmt.Println("Could not connect to server: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	return conn
}

func TCPSendMessage(conn net.Conn, sendingData []byte) {
	// Send data to the other side of the conncetion
	_, err := conn.Write(sendingData)
	if err != nil {
		//conn.Close()
		//fjern fra lista
		fmt.Printf("Could not send data: %s\n", err)
		os.Exit(1)
	}
}



func TCPReciveMessage(conn net.Conn, jsonMessageCh chan<- []byte) { //Gjør om til at den mottar FSM-state
	defer conn.Close()

	// Create a buffer to read data into
	buffer := make([]byte, 1024)

	for {
		// Read data from the client
		data, err := conn.Read(buffer) 
		if err != nil {
			//conn.Close()
			//fjern fra lista
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// Process and use the data (here, we'll just print it)
		fmt.Printf("Received: %s\n", buffer[:data])
		jsonMessageCh <- buffer[:data]
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
