package mbdFSM

import (
	"Project/network/tcp"
	"encoding/json"
	"fmt"
	"net"
)


const MasterPort = 27300
const BackupPort = 27301
const DummyPort = 27302


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


func SendElevState(state HRAElevState, serverIP string, portNr string) {
	stateBytes, _ := json.Marshal(state)
	tcp.TCP_client(stateBytes, serverIP, portNr)
}

func SendHRAInputToBackup(input HRAInput, serverIP string, portNr string) {
	inputBytes, _ := json.Marshal(input)
	fmt.Println(inputBytes)
}

var hraInput HRAInput
var backUpData HRAInput 
var hraOutput map[string][][]bool //string (nøkler) vil være IP-adressene

func MBD_FSM (MBDCh chan string, SortedAliveElevIPsCh chan []net.IP, jsonMessageCh chan []byte) {
	sortedAliveElevs := <- SortedAliveElevIPsCh
	MBD := <- MBDCh
	//jsonMessage := <- jsonMessageCh 
	for {
		switch MBD {
		case "Master":
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
				case sortedAliveElevs := <- SortedAliveElevIPsCh:
					//skal bare bruke heisene som er i live når den kaller på cost fn
					//send tilbake cab requests til heis hvis den kommer tilbake igjen og hadde eksisterende requests
				case roleChange := <- MBDCh:
					if roleChange != MBD {
						MBD = roleChange
						break
					}
					
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
				select{
				case jsonMessage := <- jsonMessageCh:
					var tempMap map[string]interface{} // Først unmarshal til et temp map for å finne ut hva slags type struct/melding det er
					err := json.Unmarshal(jsonMessage, &tempMap)
					if err != nil{
						fmt.Println("Error unmarshaling to a map:", err)
					}
					if _, exists := tempMap["hallRequests"]; exists { //hvis nøkkelen hallRequests finnes -> en struct av typen HRAInput 
						json.Unmarshal(jsonMessage, &backUpData)
					} else if  _, exists := tempMap["hraOutput"]; exists { //hvis nøkkelen inneholder hraOutput 
						json.Unmarshal(jsonMessage, hraOutput)
						myHallRequests := hraOutput[BackUpIPStr]
						// maa kalle på noe for å kjøre single elevator med de nye requestene
					}
				case changeInAliveElevs := <- SortedAliveElevIPsCh:
					sortedAliveElevs = changeInAliveElevs
					BackUpIP = sortedAliveElevs[1]
					BackUpIPStr = BackUpIP.String()

				case roleChange := <- MBDCh:
					if roleChange != MBD{
						MBD = roleChange
						break
					}
					
					
				}
			}
			
			
	
			
		case "Dummy":
			
			//Single elevator prosess kjører kanskje hele tiden?
			//Do dummy stuff
			//Send state update back to master
			//ta imot relevant (?) hraOutput
		}
	}
}

