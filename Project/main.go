package main

import (
	"Project/masterDummyBackup/roleDistributor"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/fsm"
	"net"
	//"command-line-arguments/home/student/Heis2/Sanntid_heis/Sanntid_heis/Project/network/udp_broadcast/udpBroadcaster.go"
)

// https://prod.liveshare.vsengsaas.visualstudio.com/join?C316A91544D83516CD085E57F58A55C3CD3F
func singleElevatorProcess() {
	elevio.Init("localhost:15657", elevio.N_FLOORS) //port

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)
	stopCh := make(chan bool)

	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	MBDCh := make(chan elevator.MasterBackupDummyType)
	primaryIPCh := make(chan net.IP)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go elevio.PollStopButton(stopCh)
	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, MBDCh, primaryIPCh)

	//go fsm.CheckForTimeout() Denne kjører bare en gang
	go fsm.CheckForDoorTimeOut() //Denne vil kjøre kontinuerlig

	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors()
	}

	fsm.InitLights()

	go fsm.FSM(buttonsCh, floorsCh, obstrCh, stopCh, MBDCh, primaryIPCh)
}

func main() {
	//udp_broadcast.ProcessPairInit()
	singleElevatorProcess()
	for {
		select {}
	}
}
