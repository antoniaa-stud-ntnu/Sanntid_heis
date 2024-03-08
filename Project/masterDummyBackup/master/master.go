package master

import (
	"Project/network/tcp"
	"Project/singleElevator/elevio"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
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

const MasterPort = "27300"

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
	ipAddr    string       `json:"ipAdress"`
	elevState HRAElevState `json:"elevState"`
}

type msgHallReq struct {
	tAdd_fRemove bool              `json:"tAdd_fRemove"`
	floor        int               `json:"floor"`
	button       elevio.ButtonType `json:"button"`
} // Up-0, Down-1, lik med index i HallRequests [][2]

var allHallReqAndStates HRAInput
var hraMasterState HRAElevState
var MBD string
var sortedAliveElevs []net.IP

func OnMaster(MBDCh chan string, SortedAliveElevIPsCh chan []net.IP, jsonMessageCh chan []byte, hraOutputCh chan [][2]bool) {
	iPToConnMap = make(map[net.Addr]net.Conn)
	tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMessageCh)

	var allHallReqAndStates HRAInput
	allHallReqAndStates.States = make(map[string]HRAElevState)

	for {
		select {
		case jsonMsg := <-jsonMessageCh:
			// Kodekvalitet på dette???
			var msgElevState msgElevState
			var msgHallReq msgHallReq

			if err := json.Unmarshal(jsonMsg, &msgElevState); err == nil {
				// Process msgElevState
				allHallReqAndStates.States[msgElevState.ipAddr] = msgElevState.elevState
			} else if err := json.Unmarshal(jsonMsg, &msgHallReq); err == nil {
				// Process msgHallReq
				if msgHallReq.tAdd_fRemove == true {
					// Add the correct hall request in hraInput.HallRequests
					allHallReqAndStates.HallRequests[msgHallReq.floor][msgHallReq.button] = true
				} else {
					// Remove the correct hall request in hraInput.HallRequests
					allHallReqAndStates.HallRequests[msgHallReq.floor][msgHallReq.button] = false
				}
				//Kall på hall request
				var toHRA HRAInput
				toHRA.HallRequests = allHallReqAndStates.HallRequests
				for _, ip := range sortedAliveElevs {
					toHRA.States[ip.String()] = allHallReqAndStates.States[ip.String()]
				}
				output := RunHallRequestAssigner(toHRA)
				fmt.Printf("output: \n")
				for ipAddrString, hallRequest := range output {
					ipAddr, _ := net.ResolveIPAddr("ip", ipAddrString) // String til net.Addr
					jsonHallReq, err := json.Marshal(hallRequest)
					if err != nil {
						fmt.Println("Error marshaling:", err)
					}
					tcp.TCPSendMessage(iPToConnMap[ipAddr], jsonHallReq)
					// starte timer

				}


			} else {
				fmt.Println("Master could not recieve message:", err)
			}
			var updatedElevState map[string]HRAElevState
			if err := json.Unmarshal(jsonMsg, &updatedElevState); err != nil {
				fmt.Println("Error:", err)
				continue
			}
			for elevIP, elevState := range updatedElevState { //ikke veldig elegant fordi mappet bare har en verdi
				// Process key-value pair
				allHallReqAndStates.States[elevIP] = elevState
			}
			//extract hall req
			//update hraInput
		case sortedAliveElevs := <-SortedAliveElevIPsCh:
			//skal bare bruke heisene som er i live når den kaller på cost fn
			//send tilbake cab requests til heis hvis den kommer tilbake igjen og hadde eksisterende requests

		case roleChange := <-MBDCh:
			MBD = roleChange
			break
		}
	}
	//Unmarchal til riktig format

	// Update state variable when new information is available - retry if failed

	// Send to backup

}

// Connection with elevators

func RunHallRequestAssigner(input HRAInput) map[string][][2]bool {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return nil
	}

	ret, err := exec.Command("./hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return nil
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return nil
	}
	/*
		fmt.Printf("output: \n")
		for k, v := range *output {
			fmt.Printf("%6v :  %+v\n", k, v)
		}
	*/
	return *output
}

