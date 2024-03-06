package messages

import (
	"encoding/json"
	"Project/network/tcp"
	"fmt"
)


type HRAElevState struct {
	Behaviour    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}


func SendElevState(ipAddr string, state HRAElevState, serverIP string, portNr string) {
	var stateMap map[string]HRAElevState
	stateMap[ipAddr] = state
	stateBytes, _ := json.Marshal(state)
	tcp.TCP_client(stateBytes, serverIP, portNr)
}

func SendHRAInputToBackup(input HRAInput, serverIP string, portNr string) {
	inputBytes, _ := json.Marshal(input)
	fmt.Println(inputBytes)
}

