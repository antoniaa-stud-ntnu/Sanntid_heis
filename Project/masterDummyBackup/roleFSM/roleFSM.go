package roleFSM

import (
	"Project/network/messages"
	"Project/network/tcp"
	"Project/singleElevator/elevio"
	"Project/masterDummyBackup/roleDistributor"
	"Project/masterDummyBackup/master"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
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

func roleFSM(jsonMsgCh chan []byte, RoleAndSortedAliveElevs chan roleDistributor.RoleAndSortedAliveElevs, toMbdFSMCh chan []byte, masterIPCh chan net.IP, existingIPsCh chan string) {

	//fmt.Println("First time recieving sortedAliveElevs: ", sortedAliveElevs)

	var mutexSortedElevs = &sync.Mutex{}

	role, sortedAliveElevs := (<-RoleAndSortedAliveElevs).Role, (<-RoleAndSortedAliveElevs).SortedAliveElevs

Loop:
	for {
		//masterIPCh <- sortedAliveElevs[0]

		fmt.Println("Switching to ", role)
		switch role {
		case "Master":
			var mutexIPConn = &sync.Mutex{}
			iPToConnMap := make(map[string]net.Conn)

			// Setting up masters listening server and keeping iPConnMap up to date
			go tcp.TCPListenForConnectionsAndHandle(MasterPort, jsonMsgCh, &iPToConnMap, mutexIPConn, allHallReqAndStates, existingIPsCh)
			go lookForClosedConns(&iPToConnMap, mutexIPConn)

			time.Sleep(3 * time.Second)


			masterIP := sortedAliveElevs[int(roleDistributor.Master)]
			masterIPCh <- masterIP
			fmt.Println(masterIP)

			for {
				select {
				case jsonMsg := <-toMbdFSMCh:
					go master.HandlingMsg(jsonMsg, &iPToConnMap, mutexIPConn, &sortedAliveElevs, mutexSortedElevs)

				
				case changedRoleAndSortedAliveElevs := <-RoleAndSortedAliveElevs:
					changedRole := changedRoleAndSortedAliveElevs.Role

					mutexSortedElevs.Lock()
					sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs
					mutexSortedElevs.Unlock()
					
					if changedRole != "" && changedRole != role {
						role = changedRole
						fmt.Println("Master recieved role change to ", role)
						break Loop
					}
					
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

				case changedRoleAndSortedAliveElevs := <-RoleAndSortedAliveElevs:
					changedRole := changedRoleAndSortedAliveElevs.Role

					mutexSortedElevs.Lock()
					sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs
					mutexSortedElevs.Unlock()
					
					if changedRole != "" && changedRole != role {
						role = changedRole
						fmt.Println("Master recieved role change to ", role)
						break Loop
					}
				
				}
			}

		case "Dummy":
			masterIPCh <- sortedAliveElevs[0]
			for {
				changedRoleAndSortedAliveElevs := <-RoleAndSortedAliveElevs
				changedRole := changedRoleAndSortedAliveElevs.Role

				mutexSortedElevs.Lock()
				sortedAliveElevs = changedRoleAndSortedAliveElevs.SortedAliveElevs
				mutexSortedElevs.Unlock()
				
				if changedRole != "" && changedRole != role {
					role = changedRole
					fmt.Println("Master recieved role change to ", role)
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
