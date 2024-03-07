package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"io"
)

var iPToConnMap map[net.Addr]net.Conn

type HRAElevState struct {
    Behavior    string      `json:"behaviour"`
    Floor       int         `json:"floor"` 
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
}

type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"`
    States          map[string]HRAElevState     `json:"states"`
}

var input = HRAInput{
    HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
    States: map[string]HRAElevState{
        "one": HRAElevState{
            Behavior:       "moving",
            Floor:          2,
            Direction:      "up",
            CabRequests:    []bool{false, false, false, true},
        },
        "two": HRAElevState{
            Behavior:       "idle",
            Floor:          0,
            Direction:      "stop",
            CabRequests:    []bool{false, false, false, false},
        },
    },
}

func TCPListenForConnections(hostPort string, jsonMessageCh chan []byte) { //For master to listen to incomming connections, hostPort=masterPort
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+hostPort)
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

func TCPMakeConnection(host string, port string) (net.Conn, error) { //For dummies og backup
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

func TCPReciveMessage(conn net.Conn, jsonMessageCh chan<- []byte) { //Gjør om til at den mottar FSM-state
	defer conn.Close()

	// Create a buffer to read data into
	buffer := make([]byte, 65536)

	for {
		// Read data from the client
		data, err := conn.Read(buffer) 
		if err != nil {
            // Remove the connection from activeConnections map
            //delete(activeConnections, conn)
			conn.Close()
			delete(iPToConnMap, conn.RemoteAddr())

            if err == io.EOF {
                fmt.Println("Client closed the connection.")
            } else {
                fmt.Println("Error:", err)
            }
            return
        }

		// Process and use the data (here, we'll just print it)
		//fmt.Printf("Received: %s\n", buffer[:data])
		jsonMessageCh <- buffer[:data]
	}
}

func TCPSendMessage(conn net.Conn, sendingData []byte) {
	// Send data to the other side of the conncetion
	_, err := conn.Write(sendingData)
	if err != nil {
		fmt.Println("Error sending data to server:", err)
		return
	}
}


func main() {
	
	args := os.Args

	// Check if there are at least two arguments (the first one is the program name)
	if len(args) < 2 {
		fmt.Println("Usage: ./program_name <input_argument>")
		return
	}

	// The first argument is the program name, so the input argument starts from index 1
	inputArgument := args[1]

	// Print the input argument
	fmt.Println("Input argument:", inputArgument)

	jsonMessageCh := make(chan []byte)
	switch inputArgument{
	case "0":
		iPToConnMap = make(map[net.Addr]net.Conn)
		go TCPListenForConnections("20016", jsonMessageCh)
		fmt.Println("tcp listener running")
		for {
			select{
			case msg := <- jsonMessageCh:
				fmt.Println("New message recieved")
				var input HRAInput
					if err := json.Unmarshal(msg, &input); err != nil {
						fmt.Println("Error:", err)
						continue
					}
				inputJSON, err := json.MarshalIndent(input, "", "    ")
				if err != nil {
					fmt.Println("Error marshaling input:", err)
					return
				}
				fmt.Println(string(inputJSON))
			}
		}
	case "1":
		conn, _ := TCPMakeConnection("localhost", "20016")
		fmt.Println("Connected to server")
		/*
		inputJSON, err := json.MarshalIndent(input, "", "    ")
		if err != nil {
			fmt.Println("Error marshaling input:", err)
			return
		}
		fmt.Println(string(inputJSON)) */
		
		inputBytes, _ := json.Marshal(input)/*
		// Send the JSON data to the server
		_, err := conn.Write(inputBytes)
		if err != nil {
			fmt.Println("Error sending data to server:", err)
			return
		}*/
		TCPSendMessage(conn, inputBytes)
		fmt.Println("Written to server")
		for {

		}
	}
}

/* How do i fix the errors i recieve when running?
Input argument: 0
tcp listener running
Server is listening on port 20016
panic: assignment to entry in nil map

goroutine 19 [running]:
main.TCPListenForConnections({0x52e85b, 0x5}, 0xc00009e0c0)
	/home/student/Heis2/Sanntid_heis/test.go:70 +0x2b4
created by main.main
	/home/student/Heis2/Sanntid_heis/test.go:140 +0x12c
exit status 2
*/