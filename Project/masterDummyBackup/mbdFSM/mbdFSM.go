package mbdFSM

import (
	"Project/network/messages"
	"Project/network/tcp"
	"Project/singleElevator/elevio"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// const MasterPort = "27300"
const MasterPort = "20025"
const BackupPort = "27301"
const DummyPort = "27302"

//var iPToConnMap map[net.Addr]net.Conn

// var hraInput HRAInput
var allHallReqAndStates = messages.HRAInput{
	HallRequests: make([][2]bool, elevio.N_FLOORS),
	States:       make(map[string]messages.HRAElevState),
}



func MBD_FSM(MBDCh chan string, sortedAliveElevIPsCh chan []net.IP, jsonMsgCh chan []byte, toMbdFSMCh chan []byte, masterIPCh chan net.IP) {
	
	//fmt.Println("First time recieving sortedAliveElevs: ", sortedAliveElevs)

	MBD := <-MBDCh
	sortedAliveElevs := <-sortedAliveElevIPsCh
	for {
		//masterIPCh <- sortedAliveElevs[0]
		switch MBD {
		case "Master":
			iPToConnMap := make(map[string]net.Conn)
			var mutex = &sync.Mutex{}
			go tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMsgCh, &iPToConnMap, mutex)
			time.Sleep(3 * time.Second)
			fmt.Println(sortedAliveElevs[0])
			masterIPCh <- sortedAliveElevs[0]
			
			for {
				select {
				case jsonMsg := <-toMbdFSMCh:
					typeMsg, dataMsg := messages.FromBytes(jsonMsg)
					switch typeMsg {
					case messages.MsgElevState:
						//fmt.Println(dataMsg)
						//fmt.Println("Master rceived a MsgElevState on mdbFSMCh")
						//fmt.Println("IpAddr: ", dataMsg.(messages.ElevStateMsg).IpAddr)
						//fmt.Println("State: ", dataMsg.(messages.ElevStateMsg).ElevState)
						updatingIPAddr := dataMsg.(messages.ElevStateMsg).IpAddr
						updatingElevState := dataMsg.(messages.ElevStateMsg).ElevState

						allHallReqAndStates.States[updatingIPAddr] = updatingElevState

						// Sending update to backup if backup exists (will not exist if elevator is witout internet)
						if len(sortedAliveElevs) > 1 {
							backupMsg := messages.ToBytes(messages.MsgHRAInput, allHallReqAndStates)

							mutex.Lock()
							backupConn := iPToConnMap[string(sortedAliveElevs[1].String())]
							mutex.Unlock()

							tcp.TCPSendMessage(backupConn, backupMsg)
						}
						

						//fmt.Println(allHallReqAndStates)
					case messages.MsgHallReq:
						// fmt.Println("Master rceived a MsgHallReq on mdbFSMCh")
						if dataMsg.(messages.HallReqMsg).TAddFRemove {
							// Add the correct hall request in hraInput.HallRequests
							//fmt.Println(allHallReqAndStates)
							allHallReqAndStates.HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = true
						} else {
							// fmt.Println("A hall request should be removed")
							// Remove the correct hall request in hraInput.HallRequests
							allHallReqAndStates.HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = false
						}

						// Sending update to backup if backup exists (will not exist if elevator is witout internet)
						if len(sortedAliveElevs) > 1 {
							backupMsg := messages.ToBytes(messages.MsgHRAInput, allHallReqAndStates)
							backupConn := iPToConnMap[string(sortedAliveElevs[1].String())]
							tcp.TCPSendMessage(backupConn, backupMsg)
						}

						// Sende lysoppdatering til alle heisene
						// selve sendingen skjer i for-loopen lenger ned
						jsonLightMsg := messages.ToBytes(messages.MsgHallLigths, dataMsg)
						//fmt.Println("jsonLightsMsg is: ", string(jsonLightMsg))
						
						var inputToHRA = messages.HRAInput{
							HallRequests: make([][2]bool, elevio.N_FLOORS),           
							States:       make(map[string]messages.HRAElevState), 
						}
						
						inputToHRA.HallRequests = allHallReqAndStates.HallRequests
						for _, ip := range sortedAliveElevs {
							inputToHRA.States[ip.String()] = allHallReqAndStates.States[ip.String()]
						}
						// Kall på Hall Request
						output := RunHallRequestAssigner(inputToHRA)
						
						//fmt.Println("Output: ", output)
						for ipAddr, hallRequest := range output {
							jsonHallReq := messages.ToBytes(messages.MsgAssignedHallReq, hallRequest)
							tcp.TCPSendMessage(iPToConnMap[ipAddr], jsonHallReq)
							// fmt.Println("Master sent HallReq to elev: ", string(jsonHallReq))
							tcp.TCPSendMessage(iPToConnMap[ipAddr], jsonLightMsg)
							// fmt.Println("Master sent LightMsg to elev: ", string(jsonLightMsg))
							// starte timer
						}

					}
				case changeInAliveElevs := <-sortedAliveElevIPsCh:
					sortedAliveElevs = changeInAliveElevs
					fmt.Println("Change in sortedAliveElevs: ", sortedAliveElevs)

				case roleChange := <-MBDCh:
					MBD = roleChange
					break
				}
			}

		case "Backup":
			//Sjekk om dette funker eller om man skal ha en wait for å være sikker på at master sin server kjører
			masterIPCh <- sortedAliveElevs[0]
			//ta imot hraInput og lagre

			for {
				select {
				case jsonMsg := <-toMbdFSMCh:
					typeMsg, dataMsg := messages.FromBytes(jsonMsg)
					switch typeMsg {
					case messages.MsgHRAInput:
						allHallReqAndStates = messages.HRAInput{
							HallRequests: dataMsg.(messages.HRAInput).HallRequests,
							States:       dataMsg.(messages.HRAInput).States,
						}
					}
					fmt.Println("Backup revieced all info: ", allHallReqAndStates)
				case changeInAliveElevs := <-sortedAliveElevIPsCh:
					sortedAliveElevs = changeInAliveElevs

				case roleChange := <-MBDCh:
					MBD = roleChange
					break
				}
			}

		case "Dummy":
			masterIPCh <- sortedAliveElevs[0]
			for {
				select {
				case changeInAliveElevs := <-sortedAliveElevIPsCh: // Handles changes in the list of alive elevators.
					sortedAliveElevs = changeInAliveElevs

				case roleChange := <-MBDCh: // Deals with a change in the role of the program
					MBD = roleChange
					break
				}
			}

		}
	}
}

func RunHallRequestAssigner(input messages.HRAInput) map[string][][2]bool {

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

	ret, err := exec.Command(hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
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
