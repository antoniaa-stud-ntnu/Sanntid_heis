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

var mutexAllHallAndStates = &sync.Mutex{}

var allHallReqAndStates = messages.HRAInput{
	HallRequests: make([][2]bool, elevio.N_FLOORS),
	States:       make(map[string]messages.HRAElevState),
}

func MBD_FSM(MBDCh chan string, sortedAliveElevIPsCh chan []net.IP, jsonMsgCh chan []byte, toMbdFSMCh chan []byte, masterIPCh chan net.IP) {

	//fmt.Println("First time recieving sortedAliveElevs: ", sortedAliveElevs)

	MBD := <-MBDCh

	var mutexSortedElevs = &sync.Mutex{}
	sortedAliveElevs := <-sortedAliveElevIPsCh

Loop:
	for {
		//masterIPCh <- sortedAliveElevs[0]

		fmt.Println("Switching to ", MBD)
		switch MBD {
		case "Master":
			var mutexIPConn = &sync.Mutex{}
			iPToConnMap := make(map[string]net.Conn)

			// Setting up masters listening server and keeping iPConnMap up to date
			go tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMsgCh, &iPToConnMap, mutexIPConn)
			go lookForClosedConns(&iPToConnMap, mutexIPConn)

			time.Sleep(3 * time.Second)
			fmt.Println(sortedAliveElevs[0])
			masterIPCh <- sortedAliveElevs[0]

			for {
				select {
				case jsonMsg := <-toMbdFSMCh:
					go masterHandlingMsg(jsonMsg, &iPToConnMap, mutexIPConn, &sortedAliveElevs, mutexSortedElevs)

				case changeInAliveElevs := <-sortedAliveElevIPsCh:
					go func() {
						mutexSortedElevs.Lock()
						sortedAliveElevs = changeInAliveElevs
						mutexSortedElevs.Unlock()
						fmt.Println("Change in sortedAliveElevs: ", sortedAliveElevs)
					}()

				case roleChange := <-MBDCh:
					fmt.Println("Master recieved role change to ", roleChange)
					MBD = roleChange
					fmt.Printf("MBD: %s, now breaking loop \n", MBD)
					break Loop
				}
			}

		case "Backup":
			//Sjekk om dette funker eller om man skal ha en wait for å være sikker på at master sin server kjører
			masterIPCh <- sortedAliveElevs[0]
			fmt.Println("Backup sent masterIP to elev")
			//ta imot hraInput og lagre

			for {
				select {
				case jsonMsg := <-toMbdFSMCh:
					go backupHandlingMsg(jsonMsg)

				case changeInAliveElevs := <-sortedAliveElevIPsCh:
					sortedAliveElevs = changeInAliveElevs

				case roleChange := <-MBDCh:
					MBD = roleChange
					break Loop
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
					break Loop
				}
			}

		}
	}
}

func backupHandlingMsg(jsonMsg []byte) {
	typeMsg, dataMsg := messages.FromBytes(jsonMsg)
	switch typeMsg {
	case messages.MsgHRAInput:
		mutexAllHallAndStates.Lock()
		allHallReqAndStates = messages.HRAInput{
			HallRequests: dataMsg.(messages.HRAInput).HallRequests,
			States:       dataMsg.(messages.HRAInput).States,
		}
		mutexAllHallAndStates.Unlock()
	}
	//fmt.Println("Backup revieced all info: ", allHallReqAndStates)
}

func masterHandlingMsg(jsonMsg []byte, iPToConnMap *map[string]net.Conn, mutexIPConn *sync.Mutex, sortedAliveElevs *[]net.IP, mutexSortedElevs *sync.Mutex) {
	typeMsg, dataMsg := messages.FromBytes(jsonMsg)
	switch typeMsg {
	case messages.MsgElevState:
		mutexAllHallAndStates.Lock()
		//fmt.Println(dataMsg)
		//fmt.Println("Master rceived a MsgElevState on mdbFSMCh")
		//fmt.Println("IpAddr: ", dataMsg.(messages.ElevStateMsg).IpAddr)
		//fmt.Println("State: ", dataMsg.(messages.ElevStateMsg).ElevState)
		updatingIPAddr := dataMsg.(messages.ElevStateMsg).IpAddr
		updatingElevState := dataMsg.(messages.ElevStateMsg).ElevState

		allHallReqAndStates.States[updatingIPAddr] = updatingElevState

		// Sending update to backup if backup exists (will not exist if elevator is witout internet)
		if len(*sortedAliveElevs) > 1 {
			backupMsg := messages.ToBytes(messages.MsgHRAInput, allHallReqAndStates)

			mutexIPConn.Lock()
			backupConn := (*iPToConnMap)[string((*sortedAliveElevs)[1].String())]
			mutexIPConn.Unlock()

			tcp.TCPSendMessage(backupConn, backupMsg)
		}
		mutexAllHallAndStates.Unlock()

	case messages.MsgHallReq:
		// fmt.Println("Master rceived a MsgHallReq on mdbFSMCh")
		mutexAllHallAndStates.Lock()
		allHallReqAndStates.HallRequests[dataMsg.(messages.HallReqMsg).Floor][dataMsg.(messages.HallReqMsg).Button] = dataMsg.(messages.HallReqMsg).TAddFRemove

		// Sending update to backup if backup exists (will not exist if elevator is witout internet)
		if len(*sortedAliveElevs) > 1 {

			backupMsg := messages.ToBytes(messages.MsgHRAInput, allHallReqAndStates)

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

		inputToHRA.HallRequests = allHallReqAndStates.HallRequests

		for _, ip := range *sortedAliveElevs {
			inputToHRA.States[ip.String()] = allHallReqAndStates.States[ip.String()]
		}
		mutexAllHallAndStates.Unlock()
		output := RunHallRequestAssigner(inputToHRA)

		// Hall lights setting for all elevators
		jsonLightMsg := messages.ToBytes(messages.MsgHallLigths, dataMsg)

		// Sending hall requests and light settnings to the elevators
		for ipAddr, hallRequest := range output {
			jsonHallReq := messages.ToBytes(messages.MsgAssignedHallReq, hallRequest)
			tcp.TCPSendMessage((*iPToConnMap)[ipAddr], jsonHallReq)
			// fmt.Println("Master sent HallReq to elev: ", string(jsonHallReq))
			tcp.TCPSendMessage((*iPToConnMap)[ipAddr], jsonLightMsg)
			// fmt.Println("Master sent LightMsg to elev: ", string(jsonLightMsg))
			// starte timer
		}

	}
}

func lookForClosedConns(iPToConnMap *map[string]net.Conn, mutexIPConn *sync.Mutex) {
	for ip, conn := range *iPToConnMap {

		_, err := conn.Read(make([]byte, 1024))
		if err != nil {

			mutexIPConn.Lock()
			fmt.Println("Deleting a conn: ", (*iPToConnMap))
			delete((*iPToConnMap), ip)
			mutexIPConn.Unlock()
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
