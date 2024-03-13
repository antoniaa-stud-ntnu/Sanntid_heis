package main

import (
	"Project/roleHandler/roleDistributor"
	"Project/roleHandler/roleFSM"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/localElevator/elevio"
	"Project/localElevator/singleElevatorFSM"
	"Project/roleHandler/master"
	"fmt"
	"net"
)

func ElevatorProcess() {
	fmt.Println("In single elevator process")
	elevio.Init("localhost:15657", elevio.N_FLOORS) //port

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)

	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	roleAndSortedAliveElevsCh := make(chan roleDistributor.RoleAndSortedAliveElevs, 2)
	masterIPCh := make(chan net.IP)
	quitCh := make(chan bool)

	msgToMasterCh := make(chan []byte)
	jsonMessageCh := make(chan []byte, 5)
	toSingleElevFSMCh := make(chan []byte)
	toRoleFSMCh := make(chan []byte)
	existingIPsAndConnCh := make(chan tcp.ExistingIPsAndConn)

	peerTxEnable := make(chan bool)

	// Start the elevator
	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)

	if elevio.GetFloor() == -1 {
		singleElevatorFSM.OnInitBetweenFloors()
	}

	singleElevatorFSM.InitLights()

	var masterConn chan net.Conn
	go func() {
		masterIP := <-masterIPCh
		masterConn.Close()
		fmt.Println("Master has changed to IP: ", masterIP.String())
		// Master has changed, need to make new connection
		masterConn.Close()
		masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), master.MasterPort)
		if err != nil {
			fmt.Println("Elevator could not connect to master:", err)
		}
		fmt.Println("New masterConn is: ", masterConn)
		fmt.Println("Closing the old goroutine")
		quitCh <- true // Stopping the old goroutine
		tcp.TCPRecieveMessage(masterConn, jsonMessageCh, quitCh)
	}
	

	// Make masterConn
	masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), master.MasterPort)

	go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, quitCh)
	if err != nil {
		fmt.Println("Elevator could not connect to master:", err)
	}
	fmt.Println("Elevators masterconn is: ", masterConn)

	go func() {
		
	}

	//make master conn (masterConnCh, masterIP, )
	/*case masterIP := <-masterIPCh:
			fmt.Println("Master has changed to IP: ", masterIP.String())
			// Master has changed, need to make new connection
			masterConn.Close()
			masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), master.MasterPort)
			if err != nil {
				fmt.Println("Elevator could not connect to master:", err)
			}
			fmt.Println("New masterConn is: ", masterConn)
			fmt.Println("Closing the old goroutine")
			quitCh <- true // Stopping the old goroutine
			go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, quitCh)
			//iPToConnMap := make(map[net.Addr]net.Conn)
			//masterip, _ := net.ResolveIPAddr("ip", masterIP.String()) // String til net.Addr
			//iPToConnMap[masterip] = masterConn
			//go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, &iPToConnMap)
			fmt.Println("Master has changed")*/

	//tcp send msg(to master, masterconn ch)
	//go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, quitCh)
	go messages.DistributeMessages(jsonMessageCh, toSingleElevFSMCh, toRoleFSMCh)

	//ta in to master ch i FSM
	go singleElevatorFSM.FSM(buttonsCh, floorsCh, obstrCh, masterIPCh, jsonMessageCh, toSingleElevFSMCh, quitCh, peerTxEnable)
	
	// Start communication between elevators
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh, peerTxEnable)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, roleAndSortedAliveElevsCh)
	go roleFSM.RoleFSM(jsonMessageCh, roleAndSortedAliveElevsCh, toRoleFSMCh, masterIPCh, existingIPsAndConnCh)
	
	
}



func main() {
	//udp_broadcast.ProcessPairInit()

	ElevatorProcess()
	for {
		select {}
	}
}
