package roleFSM

import (
	"Project/masterDummyBackup/master"
	"Project/masterDummyBackup/roleDistributor"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/singleElevator/elevio"
	"fmt"
	"net"
	"sync"
	"time"
)

// const MasterPort = "27300"
const MasterPort = "20025"
const BackupPort = "27301"
const DummyPort = "27302"

var mutexIPConn = &sync.Mutex{}
var iPToConnMap map[string]net.Conn

var mutexAllHallAndStates = &sync.Mutex{}

var allHallReqAndStates = messages.HRAInput{
	HallRequests: make([][2]bool, elevio.N_FLOORS),
	States:       make(map[string]messages.HRAElevState),
}

func RoleFSM(jsonMsgCh chan []byte, RoleAndSortedAliveElevsCh chan roleDistributor.RoleAndSortedAliveElevs, toRoleFSMCh chan []byte, masterIPCh chan net.IP, existingIPsAndConnCh chan tcp.ExistingIPsAndConn) {
	roleAndSortedAliveElevs := <-RoleAndSortedAliveElevsCh
	role := roleAndSortedAliveElevs.Role
	sortedAliveElevs := roleAndSortedAliveElevs.SortedAliveElevs

	fmt.Println("Received role and sorted alive: ", sortedAliveElevs)
	var mutexSortedElevs = &sync.Mutex{}

	if role == "Master" {
		//var mutexIPConn = &sync.Mutex{}
		iPToConnMap = make(map[string]net.Conn)
		// Setting up masters listening server and keeping iPConnMap up to date
		go tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMsgCh, &iPToConnMap, mutexIPConn, allHallReqAndStates, existingIPsAndConnCh)
		go tcp.TCPLookForClosedConns(&iPToConnMap, mutexIPConn)
	}

	time.Sleep(3 * time.Second)
	masterIP := sortedAliveElevs[int(roleDistributor.Master)]
	masterIPCh <- masterIP

	for {
		select {
		case jsonMsg := <-toRoleFSMCh:
			switch role {
			case "Master":
				go master.HandlingMsg(jsonMsg, &iPToConnMap, mutexIPConn, &sortedAliveElevs, mutexSortedElevs, mutexAllHallAndStates, &allHallReqAndStates)
			case "Backup":
				go backupHandlingMsg(jsonMsg)
			default:
				fmt.Println("RoleFSM recieved a message as a dummy or not assigned role")
			}

		case changedRoleAndSortedAliveElevs := <-RoleAndSortedAliveElevsCh:
			changedRole := changedRoleAndSortedAliveElevs.Role

			mutexSortedElevs.Lock()
			sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs
			mutexSortedElevs.Unlock()

			if changedRole != "" && changedRole != role {
				role = changedRole
				fmt.Println("Elevator recieved role change to ", role)
				switch role {
				case "Master":
					//var mutexIPConn = &sync.Mutex{}
					iPToConnMap := make(map[string]net.Conn)

					// Setting up masters listening server and keeping iPConnMap up to date
					go tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMsgCh, &iPToConnMap, mutexIPConn, allHallReqAndStates, existingIPsAndConnCh)
					go tcp.TCPLookForClosedConns(&iPToConnMap, mutexIPConn)

				default:
					// Default case for "Backup" and "Dummy"
					// No initialization needed
				}

				// Waiting for master to set up listening server before elevators try to connect to it
				time.Sleep(3 * time.Second)
				masterIP := sortedAliveElevs[int(roleDistributor.Master)]
				masterIPCh <- masterIP
				//fmt.Println(masterIP)

			}
		case existingIPsAndConn := <-existingIPsAndConnCh:
			cabRequests := allHallReqAndStates.States[existingIPsAndConn.ExistingIP].CabRequests
			cabReqestMsg := messages.ToBytes(messages.MsgRestoreCabReq, cabRequests)
			tcp.TCPSendMessage(existingIPsAndConn.Conn, cabReqestMsg)
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
