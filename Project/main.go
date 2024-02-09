package main

import (
	"fmt"
	"time"
	"Project/SingleElevator/elevio"
)

func main() {

	elevio.Init("localhost:15657", elevio.N_FLOORS)

	var currentDirn elevio.MotorDirection = elevio.MD_Up

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh  := make(chan int)
	obstrCh   := make(chan bool)
	stopCh    := make(chan bool)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go elevio.PollStopButton(stopCh)

	for {
        select {
        case a := <- buttonsCh:
            fmt.Printf("%+v\n", a)
            elevio.SetButtonLamp(a.Floor, a.Button, true)
            
        case a := <- floorsCh:
            fmt.Printf("%+v\n", a)
            if a == elevio.N_FLOORS-1 {
                currentDirn = elevio.MD_Down
            } else if a == 0 {
                currentDirn = elevio.MD_Up
            }
            elevio.SetMotorDirection(currentDirn)
            
        case a := <- obstrCh:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(currentDirn)
            }
            
        case a := <- stopCh:
            fmt.Printf("%+v\n", a)
            for f := 0; f < elevio.N_FLOORS; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(f, b, false)
                }
            }
		default:
			time.Sleep(500 * time.Millisecond)
        }
    }
}