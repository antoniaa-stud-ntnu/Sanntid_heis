package master

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
)

const MasterPort = "20025"



func HandlingMsg(jsonMsg []byte, iPToConnMap *map[string]net.Conn, mutexIPConn *sync.Mutex, sortedAliveElevs *[]net.IP, mutexSortedElevs *sync.Mutex, mutexAllHallAndStates *sync.Mutex, allHallReqAndStates *messages.HRAInput) {
	typeMsg, dataMsg := messages.FromBytes(jsonMsg)
	fmt.Println("In master.HandlingMsg, and iPToConnMap is: ", iPToConnMap)
	switch typeMsg {
	case messages.MsgElevState:
		mutexAllHallAndStates.Lock() 
		//
		//fmt.Println(dataMsg)
		//fmt.Println("Master rceived a MsgElevState on mdbFSMCh")
		//fmt.Println("IpAddr: ", dataMsg.(messages.ElevStateMsg).IpAddr)
		//fmt.Println("State: ", dataMsg.(messages.ElevStateMsg).ElevState)
		updatingIPAddr := dataMsg.(messages.ElevStateMsg).IpAddr
		updatingElevState := dataMsg.(messages.ElevStateMsg).ElevState

		(*allHallReqAndStates).States[updatingIPAddr] = updatingElevState

		// Sending update to backup if backup exists (will not exist if elevator is without internet)
		if len(*iPToConnMap) > 1 && len(*sortedAliveElevs) > 1 {
			backupMsg := messages.ToBytes(messages.MsgHRAInput, (*allHallReqAndStates))

			mutexIPConn.Lock()
			backupConn := (*iPToConnMap)[(*sortedAliveElevs)[1].String()]
			mutexIPConn.Unlock()

			tcp.TCPSendMessage(backupConn, backupMsg)
		}
		mutexAllHallAndStates.Unlock()

	case messages.MsgHallReq:
		// fmt.Println("Master rceived a MsgHallReq on mdbFSMCh")
		mutexAllHallAndStates.Lock()
		(*allHallReqAndStates).HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = dataMsg.(messages.HallReqMsg).TAddFRemove

		// Sending update to backup if backup exists (will not exist if elevator is witout internet)
		if len(*iPToConnMap) > 1 && len(*sortedAliveElevs) > 1 {

			backupMsg := messages.ToBytes(messages.MsgHRAInput, (*allHallReqAndStates))

			mutexIPConn.Lock()
			backupConn := (*iPToConnMap)[(*sortedAliveElevs)[1].String()]
			tcp.TCPSendMessage(backupConn, backupMsg)
			mutexIPConn.Unlock()
		}

		// Running HallRequest assigner
		var inputToHRA = messages.HRAInput{
			HallRequests: make([][2]bool, elevio.N_FLOORS),
			States:       make(map[string]messages.HRAElevState),
		}

		inputToHRA.HallRequests = (*allHallReqAndStates).HallRequests

		for _, ip := range *sortedAliveElevs {
			inputToHRA.States[ip.String()] = (*allHallReqAndStates).States[ip.String()]
		}
		mutexAllHallAndStates.Unlock()
		output := runHallRequestAssigner(inputToHRA)

		// Hall lights setting for all elevators
		jsonLightMsg := messages.ToBytes(messages.MsgHallLigths, dataMsg)

		// Sending hall requests and light settnings to the elevators
		for ipAddr, hallRequest := range output {
			jsonHallReq := messages.ToBytes(messages.MsgAssignedHallReq, hallRequest)
			fmt.Println("iPToConnMap: ", *iPToConnMap)
			tcp.TCPSendMessage((*iPToConnMap)[ipAddr], jsonHallReq)
			// fmt.Println("Master sent HallReq to elev: ", string(jsonHallReq))
			tcp.TCPSendMessage((*iPToConnMap)[ipAddr], jsonLightMsg)
			// fmt.Println("Master sent LightMsg to elev: ", string(jsonLightMsg))
			// starte timer
		}
	}
}

func runHallRequestAssigner(input messages.HRAInput) map[string][][2]bool {
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
