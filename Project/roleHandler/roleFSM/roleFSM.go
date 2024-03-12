package roleFSM

import (
	"Project/localElevator/elevio"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/roleHandler/master"
	"Project/roleHandler/roleDistributor"
	"fmt"
	"net"
)

var allHallReqAndStates = messages.HRAInput{
	HallRequests: make([][2]bool, elevio.N_FLOORS),
	States:       make(map[string]messages.HRAElevState),
}

var iPToConnMap map[string]net.Conn

var quitCh chan bool

func RoleFSM(jsonMsgCh chan []byte, RoleAndSortedAliveElevsCh chan roleDistributor.RoleAndSortedAliveElevs, toRoleFSMCh chan []byte, masterIPCh chan net.IP, existingIPsAndConnCh chan tcp.ExistingIPsAndConn) {
	roleAndSortedAliveElevs := <-RoleAndSortedAliveElevsCh
	role := roleAndSortedAliveElevs.Role
	sortedAliveElevs := roleAndSortedAliveElevs.SortedAliveElevs

	fmt.Println("Received role and sorted alive: ", sortedAliveElevs)
	fmt.Println("My role is: ", role)

	if role == "Master" {
		iPToConnMap = make(map[string]net.Conn)
		// Setting up masters listening server and keeping iPConnMap up to date
		go tcp.TCPListenForConnectionsAndHandle(master.MasterPort, jsonMsgCh, &iPToConnMap, allHallReqAndStates, existingIPsAndConnCh, quitCh)
		go tcp.TCPLookForClosedConns(&iPToConnMap)
	}

	// time.Sleep(3 * time.Second)
	masterIP := sortedAliveElevs[int(roleDistributor.Master)]
	masterIPCh <- masterIP

	for {
		select {
		case jsonMsg := <-toRoleFSMCh:
			switch role {
			case "Master":
				go master.HandlingMsg(jsonMsg, &iPToConnMap, &sortedAliveElevs, &allHallReqAndStates)
			case "Backup":
				go backupHandlingMsg(jsonMsg)
			default:
				fmt.Println("RoleFSM recieved a message as a dummy or not assigned role")
			}

		case changedRoleAndSortedAliveElevs := <-RoleAndSortedAliveElevsCh:
			changedRole := changedRoleAndSortedAliveElevs.Role
			sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs

			if changedRole != "" && changedRole != role {
				// Quitting masters reciever
				if role == "Master" {
					quitCh <- true
				}

				role = changedRole
				fmt.Println("Elevator recieved role change to ", role)
				switch role {
				case "Master":
					iPToConnMap := make(map[string]net.Conn)

					// Setting up masters listening server and keeping iPConnMap up to date
					go tcp.TCPListenForConnectionsAndHandle(master.MasterPort, jsonMsgCh, &iPToConnMap, allHallReqAndStates, existingIPsAndConnCh, quitCh)
					go tcp.TCPLookForClosedConns(&iPToConnMap)

				default:
					// Default case for "Backup" and "Dummy"
					// No initialization needed
				}

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
		allHallReqAndStates = messages.HRAInput{
			HallRequests: dataMsg.(messages.HRAInput).HallRequests,
			States:       dataMsg.(messages.HRAInput).States,
		}
	}
	fmt.Println("Backup revieced all info: ", allHallReqAndStates)
}
