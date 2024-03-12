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

	jsonMessageCh := make(chan []byte, 5)
	toSingleElevFSMCh := make(chan []byte)
	toRoleFSMCh := make(chan []byte)
	existingIPsAndConnCh := make(chan tcp.ExistingIPsAndConn)

	peerTxEnable := make(chan bool)

	// Start the elevator
	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go singleElevatorFSM.CheckForDoorTimeOut()
	go singleElevatorFSM.CheckForMotorStopTimeOut(peerTxEnable)

	if elevio.GetFloor() == -1 {
		singleElevatorFSM.OnInitBetweenFloors()
	}

	singleElevatorFSM.InitLights()

	go singleElevatorFSM.FSM(buttonsCh, floorsCh, obstrCh, masterIPCh, jsonMessageCh, toSingleElevFSMCh, quitCh, peerTxEnable)

	// Start communication between elevators
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh, peerTxEnable)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, roleAndSortedAliveElevsCh)
	go roleFSM.RoleFSM(jsonMessageCh, roleAndSortedAliveElevsCh, toRoleFSMCh, masterIPCh, existingIPsAndConnCh)
	
	go messages.DistributeMessages(jsonMessageCh, toSingleElevFSMCh, toRoleFSMCh)
}



func main() {
	//udp_broadcast.ProcessPairInit()

	ElevatorProcess()
	for {
		select {}
	}
}
