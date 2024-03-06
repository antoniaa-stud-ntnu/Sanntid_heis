package main

import (
	"Project/masterDummyBackup/roleDistributor"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/fsm"
	"Project/masterDummyBackup/mbdFSM"
	"net"
)

func singleElevatorProcess() {
	elevio.Init("localhost:15657", elevio.N_FLOORS) //port

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)

	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	MBDCh := make(chan elevator.MasterBackupDummyType)
	SortedAliveElevIPsCh := make(chan []net.IP)
	jsonMessageCh := make(chan []byte)

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
	go mbdFSM.MBD_FSM(MBDCh, SortedAliveElevIPsCh, jsonMessageCh)
	go fsm.FSM(buttonsCh, floorsCh, obstrCh, MBDCh, SortedAliveElevIPsCh)

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
