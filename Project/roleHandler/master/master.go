package master

import (
	"Project/localElevator/elevio"
	"Project/network/messages"
	"Project/network/tcp"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
)



func HandlingMessages(jsonMsg []byte, iPToConnMap *map[string]net.Conn, sortedAliveElevs *[]net.IP, allHallReqAndStates *messages.HRAInput, sendNetworkMsgCh chan tcp.SendNetworkMsg) {
	typeMsg, dataMsg := messages.UnpackMessage(jsonMsg)
	fmt.Println("In master.HandlingMsg, and iPToConnMap is: ", iPToConnMap)
	switch typeMsg {
	case messages.MsgElevState:
		//fmt.Println(dataMsg)
		fmt.Println("Master rceived a MsgElevState on mdbFSMCh")
		//fmt.Println("IpAddr: ", dataMsg.(messages.ElevStateMsg).IpAddr)
		//fmt.Println("State: ", dataMsg.(messages.ElevStateMsg).ElevState)
		updatingIPAddr := dataMsg.(messages.ElevStateMsg).IpAddr
		updatingElevState := dataMsg.(messages.ElevStateMsg).ElevState

		(*allHallReqAndStates).States[updatingIPAddr] = updatingElevState

		if len(*iPToConnMap) > 1 && len(*sortedAliveElevs) > 1 {
			backupMsg := messages.PackMessage(messages.MsgHRAInput, (*allHallReqAndStates))
			backupConn := (*iPToConnMap)[(*sortedAliveElevs)[1].String()]
			fmt.Println("BackupConn: ", backupConn)
			
			sendNetworkMsgCh <- tcp.SendNetworkMsg{backupConn, backupMsg}
			//tcp.TCPSendMessage(backupConn, backupMsg)
		}
		//fmt.Println("Master finished handling MsgElevState")

	case messages.MsgHallReq:
		fmt.Println("Master rceived a MsgHallReq on mdbFSMCh")
		(*allHallReqAndStates).HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = dataMsg.(messages.HallReqMsg).TAddFRemove

		if len(*iPToConnMap) > 1 && len(*sortedAliveElevs) > 1 {
			backupMsg := messages.PackMessage(messages.MsgHRAInput, (*allHallReqAndStates))
			backupConn := (*iPToConnMap)[(*sortedAliveElevs)[1].String()]
			sendNetworkMsgCh <- tcp.SendNetworkMsg{backupConn, backupMsg}
			// tcp.TCPSendMessage(backupConn, backupMsg)
		}

		var inputToHRA = messages.HRAInput{
			HallRequests: make([][2]bool, elevio.N_FLOORS),
			States:       make(map[string]messages.HRAElevState),
		}

		inputToHRA.HallRequests = (*allHallReqAndStates).HallRequests

		for _, ip := range *sortedAliveElevs {
			inputToHRA.States[ip.String()] = (*allHallReqAndStates).States[ip.String()]
		}
		output := runHallRequestAssigner(inputToHRA)

		jsonLightMsg := messages.PackMessage(messages.MsgHallLigths, dataMsg)

		for ipAddr, hallRequest := range output {
			jsonHallReqMsg := messages.PackMessage(messages.MsgAssignedHallReq, hallRequest)
			fmt.Println("iPToConnMap: ", *iPToConnMap)
			//tcp.TCPSendMessage((*iPToConnMap)[ipAddr], jsonHallReq)
			sendNetworkMsgCh <- tcp.SendNetworkMsg{(*iPToConnMap)[ipAddr], jsonHallReqMsg}
			// fmt.Println("Master sent HallReq to elev: ", string(jsonHallReq))
			sendNetworkMsgCh <- tcp.SendNetworkMsg{(*iPToConnMap)[ipAddr], jsonLightMsg}
			//tcp.TCPSendMessage((*iPToConnMap)[ipAddr], jsonLightMsg)
			// fmt.Println("Master sent LightMsg to elev: ", string(jsonLightMsg))
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
	return *output
}
	
	/*
		fmt.Printf("output: \n")
		for k, v := range *output {
			fmt.Printf("%6v :  %+v\n", k, v)
		}
	*/