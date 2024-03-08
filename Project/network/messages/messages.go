package messages

import (
	"Project/network/tcp"
	"encoding/json"
	"Project/singleElevator/elevio"
	"fmt"
	"net"
)

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


func SendElevState(ipAddr string, state HRAElevState, conn net.Conn) {
	var stateMap map[string]HRAElevState
	stateMap[ipAddr] = state
	stateBytes, _ := json.Marshal(state)
	tcp.TCPSendMessage(conn, stateBytes)
}

func SendHRAInputToBackup(input HRAInput, serverIP string, portNr string) {
	inputBytes, _ := json.Marshal(input)
	fmt.Println(inputBytes)
}

// Custom struct to hold type identifier and data
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// MarshalStruct marshals any variable of the provided structs into a JSON message
func MarshalStruct(msg interface{}) ([]byte, error) {
	// Determine the type of data
	var typeName string
	switch msg.(type) {
	case HRAElevState:
		typeName = "HRAElevState"
	case HRAInput:
		typeName = "HRAInput"
	case msgElevState:
		typeName = "msgElevState"
	case msgHallReq:
		typeName = "msgHallReq"
	default:
		return nil, fmt.Errorf("unsupported type")
	}

	// Create the message
	message := Message{
		Type: typeName,
		Data: msg,
	}

	// Marshal the message
	return json.Marshal(message)
}

// UnmarshalStruct unmarshals a JSON message into the original struct based on the type identifier
func UnmarshalStruct(jsonMsg []byte) (interface{}, error) {
	// Unmarshal the message
	var message Message
	err := json.Unmarshal(jsonMsg, &message)
	if err != nil {
		return nil, err
	}

	// Convert the data to a byte array
	msgData, ok := message.Data.([]byte)
	if !ok {
		return nil, fmt.Errorf("unexpected data type")
	}

	// Determine the type and unmarshal the msgData
	switch message.Type {
	case "HRAElevState":
		var state HRAElevState
		
		err := json.Unmarshal(msgData, &state)
		return state, err
	case "HRAInput":
		var input HRAInput
		err := json.Unmarshal(msgData, &input)
		return input, err
	case "msgElevState":
		var msg msgElevState
		err := json.Unmarshal(msgData, &msg)
		return msg, err
	case "msgHallReq":
		var msg msgHallReq
		err := json.Unmarshal(msgData, &msg)
		return msg, err
	default:
		return nil, fmt.Errorf("unknown type identifier")
	}
}
