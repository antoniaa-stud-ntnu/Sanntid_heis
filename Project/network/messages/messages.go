package messages

import (
	"Project/network/tcp"
	"encoding/json"
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
