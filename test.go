package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

const MasterPort = "27300"

var iPToConnMap map[net.Addr]net.Conn

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

type msgElevState struct {
	IpAddr    string       `json:"ipAdress"`
	ElevState HRAElevState `json:"elevState"`
}

// Custom struct to hold type identifier and data
type SendMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type RecieveMessage struct {
	Type string      `json:"type"`
	Data []byte 	`json:"data"`
}

// MarshalStruct marshals any variable of the provided structs into a JSON message
func MarshalStruct(msg interface{}) ([]byte, error) {
	// Determine the type of data
	var typeName string
	switch msg.(type) {
	case HRAInput:
		typeName = "HRAInput"
	case msgElevState:
		typeName = "msgElevState"
	default:
		return nil, fmt.Errorf("unsupported type")
	}

	// Create the message
	message := SendMessage{
		Type: typeName,
		Data: msg,
	}

	// Marshal the message

	return json.Marshal(message)
}

// UnmarshalStruct unmarshals a JSON message into the original struct based on the type identifier
func UnmarshalStruct(jsonMsg []byte) (string, interface{}, error) {
	fmt.Println("Unmarshalling message")
	// Unmarshal the message
	var message SendMessage
	err := json.Unmarshal(jsonMsg, &message)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return "", nil, err
	}

	// Convert the data to a byte array
	//msgData := []byte(fmt.Sprint(message.Data))
	msgData:= []byte(fmt.Sprintf("%v", message.Data))
	fmt.Println("msgData:", string(msgData))

	// Determine the type and unmarshal the msgData
	switch message.Type {
	case "HRAInput":
		var input HRAInput
		err := json.Unmarshal(msgData, &input)
		fmt.Println("Input:", input)
		fmt.Println("Error:", err)
		return message.Type, input, err
	case "msgElevState":
		var msg msgElevState
		err := json.Unmarshal(msgData, &msg)
		return message.Type, msg, err
	default:
		return "nil", nil, fmt.Errorf("unknown type identifier")
	}
}

// UnmarshalStruct unmarshals a JSON message into the original struct based on the type identifier
func UnmarshalStruct2(jsonMsg []byte) (string, interface{}, error) {
	fmt.Println("Unmarshalling message")
	// Unmarshal the message
	var message SendMessage
	err := json.Unmarshal(jsonMsg, &message)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return "", nil, err
	}
	//msgData := []byte(fmt.Sprintf("%v", message.Data))
	//fmt.Println(string(msgData))
	// Determine the type and unmarshal the Data field directly

	// Marshal the Data field into JSON
    msgData, err := json.Marshal(message.Data)
    if err != nil {
        fmt.Println("Error marshalling message data:", err)
        return "", nil, err
    }

    switch message.Type {
    case "HRAElevState":
        var state HRAElevState
        err := json.Unmarshal(msgData, &state)
        return message.Type, state, err
    case "HRAInput":
        var input HRAInput
        err := json.Unmarshal(msgData, &input)
        //fmt.Println("Input:", input)
        //fmt.Println("Error:", err)
        return message.Type, input, err
    case "msgElevState":
        var msg msgElevState
        err := json.Unmarshal(msgData, &msg)
        return message.Type, msg, err
    default:
        return "nil", nil, fmt.Errorf("unknown type identifier")
    }
}

// MarshalStruct marshals any variable of the provided structs into a JSON message
func MarshalStruct3(msg interface{}) ([]byte, error) {
	// Create the message
	message := SendMessage{
		Type: "",
		Data: nil,
	}

	switch m := msg.(type) {
	case HRAInput:
		message.Type = "HRAInput"
		message.Data = m
	case msgElevState:
		message.Type = "msgElevState"
		message.Data = m
	default:
		return nil, fmt.Errorf("unsupported type")
	}

	// Marshal the message
	return json.Marshal(message)
}

// UnmarshalStruct unmarshals a JSON message into the original struct based on the type identifier
func UnmarshalStruct3(jsonMsg []byte) (string, interface{}, error) {
	fmt.Println("Unmarshalling message")
	// Unmarshal the message
	var message SendMessage
	err := json.Unmarshal(jsonMsg, &message)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return "", nil, err
	}

	switch message.Type {
	case "HRAInput":
		var input HRAInput
		err := json.Unmarshal(jsonMsg, &input)
		return message.Type, input, err
	case "msgElevState":
		var msg msgElevState
		err := json.Unmarshal(jsonMsg, &msg)
		return message.Type, msg, err
	default:
		return "nil", nil, fmt.Errorf("unknown type identifier")
	}
}


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

var input = HRAInput{
	HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
	States: map[string]HRAElevState{
		"one": HRAElevState{
			Behavior:    "moving",
			Floor:       2,
			Direction:   "up",
			CabRequests: []bool{false, false, false, true},
		},
		"two": HRAElevState{
			Behavior:    "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: []bool{false, false, false, false},
		},
	},
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
	switch inputArgument {
	case "0":
		iPToConnMap = make(map[net.Addr]net.Conn)
		go TCPListenForConnectionsAndHandle("20016", jsonMessageCh)
		fmt.Println("tcp listener running")
		for {
			select {
			case msg := <-jsonMessageCh:
				fmt.Println("New message received")
				msgType, msgOutput, _ := UnmarshalStruct3(msg)
				fmt.Println(msgType)
				fmt.Println("From client:", msgOutput)
			}
		}
	case "1":
		conn, _ := TCPMakeMasterConnection("localhost", "20016")
		fmt.Println("Connected to server")
		/*
			inputJSON, err := json.MarshalIndent(input, "", "    ")
			if err != nil {
				fmt.Println("Error marshaling input:", err)
				return
			}
			fmt.Println(string(inputJSON)) */
		inputBytes, _ := MarshalStruct3(input)
		TCPSendMessage(conn, inputBytes)
		fmt.Println("Written to server", string(inputBytes[:]))
		conn.Close()
	}
}

/*
Output from terminal when running the program 
Input argument: 1
Connected to server
Written to server {"type":"HRAInput","data":{"hallRequests":[[false,false],[true,false],[false,false],[false,true]],"states":{"one":{"behaviour":"moving","floor":2,"direction":"up","cabRequests":[false,false,false,true]},"two":{"behaviour":"idle","floor":0,"direction":"stop","cabRequests":[false,false,false,false]}}}}

Input argument: 0
tcp listener running
Server is listening on port 20016
New message received
Unmarshalling message
Client closed the connection.
HRAInput
From client: {[[false false] [true false] [false false] [false true]] map[one:{moving 2 up [false false false true]} two:{idle 0 stop [false false false false]}]}
*/