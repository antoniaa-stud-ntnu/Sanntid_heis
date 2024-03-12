package main

import (
	"Project/masterDummyBackup/roleDistributor"
	"Project/masterDummyBackup/roleFSM"
	"Project/network/messages"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevio"
	"Project/singleElevator/singleElevatorFSM"
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
	MBDCh := make(chan string, 2)
	masterIPCh := make(chan net.IP)
	sortedAliveElevIPsCh := make(chan []net.IP, 2)
	jsonMessageCh := make(chan []byte)
	toFSMCh := make(chan []byte)
	toMbdFSMCh := make(chan []byte)
	existingIPsCh := make(chan string)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, MBDCh, sortedAliveElevIPsCh)

	go singleElevatorFSM.CheckForDoorTimeOut() //Denne vil kjøre kontinuerlig

	if elevio.GetFloor() == -1 {
		singleElevatorFSM.OnInitBetweenFloors()
	}

	singleElevatorFSM.InitLights()
	go roleFSM.MBD_FSM(MBDCh, sortedAliveElevIPsCh, jsonMessageCh, toMbdFSMCh, masterIPCh, existingIPsCh)
	go singleElevatorFSM.FSM(buttonsCh, floorsCh, obstrCh, masterIPCh, jsonMessageCh, toFSMCh)
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
