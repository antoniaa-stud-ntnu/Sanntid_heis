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
)

const MasterPort = "27300"
const BackupPort = "27301"
const DummyPort = "27302"

var iPToConnMap map[net.Addr]net.Conn

// var hraInput HRAInput
var allHallReqAndStates messages.HRAInput

func MBD_FSM(MBDCh chan string, sortedAliveElevIPsCh chan []net.IP, jsonMsgCh chan []byte, toMbdFSMCh chan []byte, masterIPCh chan net.IP) {
	iPToConnMap = make(map[net.Addr]net.Conn)
	sortedAliveElevs := <- sortedAliveElevIPsCh

	//var sortedAliveElevs []net.IP
	MBD := <-MBDCh
	for {
		masterIPCh <- sortedAliveElevs[0]
		switch MBD {
		case "Master":
			fmt.Println("Inni master i mbdFSM")
			tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMsgCh, &iPToConnMap)
			//allHallReqAndStates.States = make(map[string]messages.HRAElevState)
			for {
				select {
				case jsonMsg := <-toMbdFSMCh:
					typeMsg, dataMsg := messages.FromBytes(jsonMsg)
					switch typeMsg {
					case messages.MsgElevState:

						allHallReqAndStates.States[dataMsg.(messages.ElevStateMsg).IpAddr] = dataMsg.(messages.ElevStateMsg).ElevState
					case messages.MsgHallReq:
						if dataMsg.(messages.HallReqMsg).TAddFRemove == true {
							// Add the correct hall request in hraInput.HallRequests
							allHallReqAndStates.HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = true
						} else {
							// Remove the correct hall request in hraInput.HallRequests
							allHallReqAndStates.HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = false
						}
						// Sende lysoppdatering til alle heisene
						// selve sendingen skjer i for-loopen lenger ned
						jsonLightMsg := messages.ToBytes(messages.MsgHallLigths, dataMsg)

						var inputToHRA messages.HRAInput
						inputToHRA.HallRequests = allHallReqAndStates.HallRequests
						for _, ip := range sortedAliveElevs {
							inputToHRA.States[ip.String()] = allHallReqAndStates.States[ip.String()]
						}
						// Kall på Hall Request
						output := RunHallRequestAssigner(inputToHRA)
						fmt.Printf("output: \n")
						for ipAddrString, hallRequest := range output {
							ipAddr, _ := net.ResolveIPAddr("ip", ipAddrString) // String til net.Addr
							jsonHallReq := messages.ToBytes(messages.MsgAssignedHallReq, hallRequest)
							tcp.TCPSendMessage(iPToConnMap[ipAddr], jsonHallReq)
							tcp.TCPSendMessage(iPToConnMap[ipAddr], jsonLightMsg)
							// starte timer
						}
					}
				case changeInAliveElevs := <-sortedAliveElevIPsCh:
					fmt.Println("Inni master i mbdFSM, i changeInAliveElevs")
					sortedAliveElevs = changeInAliveElevs

				case roleChange := <-MBDCh:
					MBD = roleChange
					break
				}
			}

		case "Backup":
			//ta imot hraInput og lagre
			fmt.Println("Inni backup i mbdFSM")
			allHallReqAndStates = messages.HRAInput{
				HallRequests: make([][2]bool, elevio.N_FLOORS),
				States:       make(map[string]messages.HRAElevState),
			}

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
				case changeInAliveElevs := <-sortedAliveElevIPsCh:
					sortedAliveElevs = changeInAliveElevs
					fmt.Println("woilajksdoawlk")

				case roleChange := <-MBDCh:
					MBD = roleChange
					break
				}
			}

		case "Dummy":
			fmt.Println("Inni dummy i mbdFSM")
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
