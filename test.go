package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"Project/singleElevator/elevio"
	//"bytes"
	//"encoding/binary"
)

const MasterPort = "27300"
const HRAInputType = "HRAInput"
const MsgElevStateType = "msgElevState"
const MsgHallReqType = "msgHallReq"

var iPToConnMap map[net.Addr]net.Conn

type HRAElevState struct {
	Behaviour   string `json:"behaviour"`
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

type msgHallReq struct {
	TAddFRemove bool              `json:"tAdd_fRemove"`
	Floor       int               `json:"floor"`
	Button      elevio.ButtonType `json:"button"`
}

type dataWithType struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
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

func ToBytes(structType string, msg interface{}) []byte {
	//fmt.Println(msg.(type))
	msgJsonBytes, _ := json.Marshal(msg)

	dataToSend := dataWithType{
		Type: structType,
		Data: msgJsonBytes,
	}

	finalJSONBytes, _ := json.Marshal(dataToSend)
	return finalJSONBytes
}

func FromBytes(jsonBytes []byte) (string, interface{}) {
	var DataWithType dataWithType
	err := json.Unmarshal(jsonBytes, &DataWithType)
	if err != nil {
		// Handle error
		return DataWithType.Type, nil
	}
	switch DataWithType.Type {
	case "HRAInput":
		var HRAInputData HRAInput
		err = json.Unmarshal(DataWithType.Data, &HRAInputData)
		return DataWithType.Type, HRAInputData
	case "msgElevState":
		var MsgElevStateData msgElevState
		err = json.Unmarshal(DataWithType.Data, &MsgElevStateData)
		return DataWithType.Type, MsgElevStateData
	case "msgHallReq":
		var MsgHallReqData msgHallReq
		err = json.Unmarshal(DataWithType.Data, &MsgHallReqData)
		return DataWithType.Type, MsgHallReqData
	case "msgHallLights":
		var MsgHallLightsData msgHallReq
		err = json.Unmarshal(DataWithType.Data, &MsgHallLightsData)
		return DataWithType.Type, MsgHallLightsData
	default:
		return DataWithType.Type, nil
	}
}

var input = HRAInput{
	HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
	States: map[string]HRAElevState{
		"one": HRAElevState{
			Behaviour:   "moving",
			Floor:       2,
			Direction:   "up",
			CabRequests: []bool{false, false, false, true},
		},
		"two": HRAElevState{
			Behaviour:   "idle",
			Floor:       0,
			Direction:   "stop",
			CabRequests: []bool{false, false, false, false},
		},
	},
}

//var hallReqs := {{false, false}, {true, false}, {false, false}, {false, true}}

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
				msgType, data := FromBytes(msg)
				switch msgType {
				case HRAInputType:
					
					fmt.Println("message is: ", data.(HRAInput).HallRequests)
					
				//fmt.Println("message is: ", msgType.(HRAInput).HallRequests)

				}
			}
		}
	case "1":
		conn, _ := TCPMakeMasterConnection("localhost", "20016")
		fmt.Println("Connected to server")
		//inputBytes := ToBytes("HRAInput", hallReqs)
		TCPSendMessage(conn, inputBytes)
		fmt.Println("Written to server", string(inputBytes[:]))
		conn.Close()
		
	}
}