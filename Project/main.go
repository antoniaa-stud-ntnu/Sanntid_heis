package main

import (
	"Project/singleElevator/elevio"
	"Project/singleElevator/fsm"
	"Project/network/Network-go/network/peers"
)

// https://prod.liveshare.vsengsaas.visualstudio.com/join?C316A91544D83516CD085E57F58A55C3CD3F
func singleElevatorProcess() {
	elevio.Init("localhost:15657", elevio.N_FLOORS)

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)
	stopCh := make(chan bool)
	peerUpdateCh := make(chan peers.PeerUpdate)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go elevio.PollStopButton(stopCh)

	//go fsm.CheckForTimeout() Denne kjører bare en gang
	go fsm.CheckForDoorTimeOut() //Denne vil kjøre kontinuerlig

	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors()
	}

	fsm.InitLights()

	go fsm.FSM(buttonsCh, floorsCh, obstrCh, stopCh)
}

func main() {
	//udp_broadcast.ProcessPairInit()
	singleElevatorProcess()
	for {
		select {}
	}
}
