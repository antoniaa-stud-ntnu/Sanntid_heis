package main

import (
	"Project/SingleElevator/elevio"
	"Project/SingleElevator/fsm"
	"fmt"
	"time"
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

    go fsm.CheckForTimeout()

    if elevio.GetFloor() == -1 {
        fsm.FsmOnInitBetweenFloors()
    }

    fsm.InitLights()

	for {
        fsm.CheckForTimeout()
        select {
        case a := <- buttonsCh:
            fmt.Printf("%+v\n", a)
            //elevio.SetButtonLamp(a.Floor, a.Button, true)
            fsm.FsmOnRequestButtonPress(a.Floor, a.Button)
            
        case a := <- floorsCh:
            /*
            fmt.Printf("%+v\n", a)
            if a == elevio.N_FLOORS-1 {
                currentDirn = elevio.MD_Down
            } else if a == 0 {
                currentDirn = elevio.MD_Up
            }
            elevio.SetMotorDirection(currentDirn)*/
            fsm.FsmOnFloorArrival(a)
            
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

/*
for {
    int prev[elevio.N_FLOORS][elevio.N_BUTTONS]
    for f := 0; f < elevio.N_FLOORS; f++ {     
        for b := 0; b < elevio.N_BUTTONS; b++ {
            int v 
        }
    }
}*/