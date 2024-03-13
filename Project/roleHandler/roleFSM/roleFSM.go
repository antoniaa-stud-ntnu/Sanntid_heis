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



func RoleFSM(roleAndSortedAliveElevsCh chan roleDistributor.RoleAndSortedAliveElevs, toRoleFSMCh chan []byte, sendNetworkMsgCh chan tcp.SendNetworkMsg, isMasterCh chan bool, editMastersConnMapCh chan tcp.EditConnMap) {
	roleAndSortedAliveElevs := <-roleAndSortedAliveElevsCh
	role := roleAndSortedAliveElevs.Role
	sortedAliveElevs := roleAndSortedAliveElevs.SortedAliveElevs

	fmt.Println("Received role and sorted alive: ", sortedAliveElevs)
	fmt.Println("My role is: ", role)
	
	if role == "Master" {
		isMasterCh <- true
		iPToConnMap = make(map[string]net.Conn)
		//go tcp.ListenForConnectionsAndHandle(master.MasterPort, &iPToConnMap, newConnIPCh, incommingNetworkMsgCh, quitMasterRecieverCh)
		//go tcp.LookForClosedConns(&iPToConnMap)
	}


	for {
		select {
		case newMsg := <-toRoleFSMCh:
			fmt.Println("newMsg")
			switch role {
			case "Master":
				go master.HandlingMessages(newMsg, &iPToConnMap, &sortedAliveElevs, &allHallReqAndStates, sendNetworkMsgCh)
			case "Backup":
				go backupHandlingMessages(newMsg)
			default:
				fmt.Println("RoleFSM recieved a message as a dummy or not assigned role")
			}

		case changedRoleAndSortedAliveElevs := <-roleAndSortedAliveElevsCh:
			changedRole := changedRoleAndSortedAliveElevs.Role
			sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs

			if changedRole != "" && changedRole != role {
				role = changedRole
				fmt.Println("Elevator recieved role change to ", role)
				switch role {
				case "Master":
					//iPToConnMap := make(map[string]net.Conn)
					iPToConnMap = make(map[string]net.Conn)
					//go tcp.ListenForConnectionsAndHandle(master.MasterPort, &iPToConnMap, newConnIPCh, incommingNetworkMsgCh, quitMasterRecieverCh)
					//go tcp.LookForClosedConns(&iPToConnMap)
					isMasterCh <- true

				default:
					isMasterCh <- false
				}

				//masterIP := sortedAliveElevs[int(roleDistributor.Master)]
				//masterIPCh <- masterIP
				//fmt.Println(masterIP)

			}	
		case editMastersConnMap := <- editMastersConnMapCh:
			insert := editMastersConnMap.Insert
			elevatorIP := editMastersConnMap.ClientIP
			elevatorConn := editMastersConnMap.Conn
			if insert {
				(iPToConnMap)[elevatorIP] = elevatorConn
				if _, exists := allHallReqAndStates.States[elevatorIP]; exists {
					cabRequests := allHallReqAndStates.States[elevatorIP].CabRequests
					cabReqestMsg := messages.PackMessage(messages.MsgRestoreCabReq, cabRequests)
					sendNetworkMsgCh <- tcp.SendNetworkMsg{iPToConnMap[elevatorIP], cabReqestMsg}
					// tcp.SendMessage(iPToConnMap[newConnIP], cabReqestMsg)
				}
			} else {
				delete((iPToConnMap), elevatorIP)
			}
			
			// Hvis conn ikke allerede eksisterer i allHallReqAndStates fra roleFSM
			//insert := tcp.editMastersConnMap
			
		}
	
	}
}

func backupHandlingMessages(jsonMsg []byte) {
	typeMsg, dataMsg := messages.UnpackMessage(jsonMsg)
	switch typeMsg {
	case messages.MsgHRAInput:
		allHallReqAndStates = messages.HRAInput{
			HallRequests: dataMsg.(messages.HRAInput).HallRequests,
			States:       dataMsg.(messages.HRAInput).States,
		}
	}
	fmt.Println("Backup revieced all info: ", allHallReqAndStates)
}
