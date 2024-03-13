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
	masterConnCh := make(chan net.Conn)
	quitOldRecieverCh := make(chan bool)

	sendNetworkMsgCh := make(chan tcp.SendNetworkMsg) 

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

	
	

	go tcp.EstablishConnection(masterIPCh, master.MasterPort, masterConnCh, quitOldRecieverCh)
	go tcp.RecieveMessage(masterConnCh, jsonMessageCh, quitOldRecieverCh)
	go tcp.SendMessage(sendNetworkMsgCh)

	
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
