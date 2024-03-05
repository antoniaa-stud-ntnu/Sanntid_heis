package master

import (
	"encoding/json"
	"Project/network/tcp"
	"Project/masterDummyBackup/roleDistributor"
	"Project/masterDummyBackup/dummy"
	"Project/singleElevator/elevator"
)

// Connection with elevators
// Needs to update state variable when new information is available - retry if failed
// Sends the information to backup
// Recieves either ACK or NACK from backup that the information is stored
// If NACK, it needs to resend the information
// Run hall_request_assigner
// Send hall_request_assigner output to elevators

// Alive elevators variable, contains ip adresses and port to all alive elevators

// Update all states variable

// Run hall_request_assigner

// Primary init
// Initialize all states
// go sendIPtoPrimaryBroadcast() //Im alive message
// Choose a backup randomly from the alive elevators --> spawn backup.
//


//ret, err := exec.Command("../hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()

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

var hraInput HRAInput
var hraMasterState HRAElevState

func OnMaster() {
	// Connection with elevators
	tcp.handleClient()
	tcp.TCP_server(MasterIP, MasterPort)

	// Update state variable when new information is available - retry if failed

	// Send information to backup
	sendToBackup()
	
}

func sendElevState() {
	stateBytes, _ := json.Marshal(hraMasterState)
	TCP_client(stateBytes, MasterIP, MasterPort)
}

func sendToBackup() {

}