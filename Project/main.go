package main

import (
	"Project/masterDummyBackup/mbdFSM"
	"Project/masterDummyBackup/roleDistributor"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/fsm"
	"net"
)

func singleElevatorProcess() {
	elevio.Init("localhost:15657", elevio.N_FLOORS) //port

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)

	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	MBDCh := make(chan string)
	SortedAliveElevIPsCh := make(chan []net.IP)
	jsonMessageCh := make(chan []byte)
	hraOutputCh := make(chan [][2]bool)
	lightsCh := make(chan [][2]bool)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, MBDCh, SortedAliveElevIPsCh)

	//go fsm.CheckForTimeout() Denne kjører bare en gang
	go fsm.CheckForDoorTimeOut() //Denne vil kjøre kontinuerlig

	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors()
	}

	fsm.InitLights()
	go mbdFSM.MBD_FSM(MBDCh, SortedAliveElevIPsCh, jsonMessageCh, lightsCh)
	go fsm.FSM(buttonsCh, floorsCh, hraOutputCh, obstrCh, lightsCh, conn) // endre conn til å være den conn som returneres fra TCPconnection
}

func MBDProsess() {

}

func main() {
	//udp_broadcast.ProcessPairInit()

	singleElevatorProcess()
	for {
		select {}
	}
}
