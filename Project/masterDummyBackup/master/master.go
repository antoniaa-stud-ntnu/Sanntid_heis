package master

import (
	"Project/network/tcp"
	"encoding/json"
	"fmt"
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

const MasterPort = 27300
var iPToConnMap map[net.Addr]net.Conn


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

var hraInput HRAInput
var hraMasterState HRAElevState

func OnMaster(jsonMessageCh chan []byte) {
	// Connection with elevators
	tcp.TCPListenForConnections(MasterPort, jsonMessageCh)
	for {
		select {
		case jsonElevUpdate := <- jsonMessageCh:
			var updatedElevState map[string]HRAElevState
			if err := json.Unmarshal(jsonElevUpdate, &updatedElevState); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			for elevIP, elevState := range updatedElevState { //ikke veldig elegant fordi mappet bare har en verdi
				// Process key-value pair
				hraInput.States[elevIP] = elevState
			}
			//extract hall req
			//update hraInput
			
		}
	}
	//Unmarchal til riktig format

	// Update state variable when new information is available - retry if failed

	// Send information to backup
	sendToBackup()

}
