package main

import (
	"Project/singleElevator/elevio"
	"Project/singleElevator/fsm"
)

func main() {
	elevio.Init("localhost:15657", elevio.N_FLOORS)

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)
	stopCh := make(chan bool)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go elevio.PollStopButton(stopCh)

	//go fsm.CheckForTimeout() Denne kjører bare en gang
	go fsm.CheckForTimeout_continously() //Denne vil kjøre kontinuerlig

	if elevio.GetFloor() == -1 {
		fsm.OnInitBetweenFloors()
	}

	fsm.InitLights()

	fsm.FSM(buttonsCh, floorsCh, obstrCh, stopCh)
}
