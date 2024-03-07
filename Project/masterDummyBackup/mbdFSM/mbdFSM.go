package mbdFSM

import (
	"Project/network/tcp"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"Project/network/udpBroadcasr/udpNetwork/localip"
	"Project/singleElevator/elevio"
)

const MasterPort = "27300"
const BackupPort = "27301"
const DummyPort = "27302"

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
	ipAddr		string			`json:"ipAdress"`
	elevState	HRAElevState	`json:"elevState"`
}

type msgHallReq struct {
	tAdd_fRemove 	bool				`json:"tAdd_fRemove"`
	floor			int					`json:"floor"`
	button			elevio.ButtonType	`json:"button"`
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

//var hraInput HRAInput
var hraInput HRAInput
var hraOutput map[string][][2]bool //string (nøkler) vil være IP-adressene


func MBD_FSM(MBDCh chan string, SortedAliveElevIPsCh chan []net.IP, jsonMessageCh chan []byte, hraOutputCh chan [][2]bool, lightsCh [][]bool) {
	sortedAliveElevs := <-SortedAliveElevIPsCh
	MBD := <-MBDCh
	//jsonMessage := <- jsonMessageCh
	for {
		switch MBD {
		case "Master":
			// Connection with elevators
			iPToConnMap = make(map[net.Addr]net.Conn)
			tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMessageCh)
			for {
				select {
				case jsonElevUpdate := <-jsonMessageCh:
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
				case sortedAliveElevs := <-SortedAliveElevIPsCh:
					//skal bare bruke heisene som er i live når den kaller på cost fn
					//send tilbake cab requests til heis hvis den kommer tilbake igjen og hadde eksisterende requests

				case roleChange := <- MBDCh:
					/*
						if roleChange != MBD {
							MBD = roleChange
							break
						}*/
					MBD = roleChange
					break
				}
			}
			//Unmarchal til riktig format

			// Update state variable when new information is available - retry if failed

			// Send information to backup
			//go listener, for å ta imot states
			//ta i mot elevator state og oppdater
			// sjekke hraOutput og hvilke requests seg selv skal ta
		case "Backup":
			//ta imot hraInput og lagre
			BackUpIP := sortedAliveElevs[1]
			BackUpIPStr := BackUpIP.String()
			for {
				select {
				case jsonMessage := <-jsonMessageCh:
					var tempMap map[string]interface{}
					err := json.Unmarshal(jsonMessage, &tempMap)
					if err != nil {
						fmt.Println("Error unmarshaling to a map:", err)
					}
					if _, exists := tempMap["hallRequests"]; exists { //hvis nøkkelen hallRequests finnes -> en struct av typen HRAInput
						json.Unmarshal(jsonMessage, &hraInput)
					} else if _, exists := tempMap["hraOutput"]; exists { //hvis nøkkelen inneholder hraOutput
						json.Unmarshal(jsonMessage, hraOutput)
						myHallRequests := hraOutput[BackUpIPStr]
						hraOutputCh <- myHallRequests
						// maa kalle på noe for å kjøre single elevator med de nye requestene
					} else if _, exists := tempMap["lights"]; exists {
						json.Unmarshal(jsonMessage, )
						lightsMsg := hraOutput[BackUpIPStr]
						lightsCh <- lightsMsg

					} 
				case changeInAliveElevs := <-SortedAliveElevIPsCh:
					sortedAliveElevs = changeInAliveElevs
					BackUpIP = sortedAliveElevs[1]
					BackUpIPStr = BackUpIP.String()

				case roleChange := <-MBDCh:
					/*if roleChange != MBD{
						MBD = roleChange
						break
					}*/
					MBD = roleChange
					break
				}
			}

		case "Dummy":
			// Need to move some of these outside the for-loop, or delete things...
			var dummyState HRAElevState
			DummyIPs := sortedAliveElevs[2:] // Slice of IP-adresses for dummies
			//dummyStateUpdateMap := make(map[string]HRAElevState)
			stateUpdateCh := make(chan bool)
			var dummyIPStr string

			// This is the new koooode
			var msgElevState msgElevState
			dummyIP, _ := localip.LocalIP();  // Get elevator IP

			// Set up TCP-connection with Master
			masterConn := tcp.TCPMakeConncetion(sortedAliveElevs[0].String(), strconv.Itoa(MasterPort))

			// How to get Dummy states?
			// type msgElevState struct {
			// 	ipAddr		string
			// 	elevState	HRAElevState
			// }
			for {
				select {
				case <-stateUpdateCh: // Når det skjer en endring i elevator sine states, send ip + state til master
					msgElevState.ipAddr = dummyIP

					// Convert to marshal
					jsonDummyStateUpdate, err := json.Marshal(msgElevState)
					if err != nil {
						fmt.Println("Error marshaling:", err)
					}
					// Send on channel
					jsonMessageCh <- jsonDummyStateUpdate

				case jsonMessage := <-jsonMessageCh:
					tempMap := make(map[string]interface{})      // Temporarily map to hold the JSON data in order to determine its type
					err := json.Unmarshal(jsonMessage, &tempMap) // Unmarshal the data sent from Master (via TCP) into the map
					if err != nil {
						fmt.Println("Error unmarshaling to a map:", err)
					}
					if _, exists := tempMap["hraOutput"]; exists { // If map contains the key "hraOutput"
						json.Unmarshal(jsonMessage, hraOutput) // Unmarshal the data related to the dummy IP
						if err != nil {
							fmt.Println("Error unmarshaling:", err)
						}
						for _, dummyIP := range DummyIPs {
							dummyIPStr = dummyIP.string()          // Gjør om IP til string-IP
							myHallRequests = hraOutput[dummyIPStr] // Oppdaterer dummyState på angitt IP
						}
					}

				case changeInAliveElevs := <-SortedAliveElevIPsCh: // Handles changes in the list of alive elevators.
					sortedAliveElevs = changeInAliveElevs
					DummyIPs = sortedAliveElevs[2:]

				case roleChange := <-MBDCh: // Deals with a change in the role of the program
					/*if roleChange != MBD {
						MBD = roleChange
						break
					}*/
					MBD = roleChange
					break
				}
			}

		}
	}
}
