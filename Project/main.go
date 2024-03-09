package main

import (
	"Project/masterDummyBackup/mbdFSM"
	"Project/masterDummyBackup/roleDistributor"
	"Project/network/messages"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevio"
	"Project/singleElevator/fsm"
	"fmt"
	"net"
)

func singleElevatorProcess() {
	fmt.Println("In single elevator process")
	elevio.Init("localhost:15657", elevio.N_FLOORS) //port

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)

	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	MBDCh := make(chan string)
	masterIPCh := make(chan net.IP)
	sortedAliveElevIPsCh := make(chan []net.IP)
	jsonMessageCh := make(chan []byte)
	toFSMCh := make(chan []byte)
	toMbdFSMCh := make(chan []byte)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, MBDCh, sortedAliveElevIPsCh)

	go fsm.CheckForDoorTimeOut() //Denne vil kjøre kontinuerlig

	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors()
	}

	fsm.InitLights()
	go mbdFSM.MBD_FSM(MBDCh, sortedAliveElevIPsCh, jsonMessageCh, toMbdFSMCh, masterIPCh)
	go fsm.FSM(buttonsCh, floorsCh, obstrCh, masterIPCh, jsonMessageCh, toFSMCh)
	go messages.DistributeMessages(jsonMessageCh, toFSMCh, toMbdFSMCh)
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
