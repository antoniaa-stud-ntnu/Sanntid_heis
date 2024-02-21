package fsm

import (
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/requests"
	"Project/singleElevator/timer"
	"fmt"
)

// LOOK HERE
var elev elevator.Elevator = elevator.InitElev()

func FSM(buttonsCh chan elevio.ButtonEvent, floorsCh chan int, obstrCh chan bool, stopCh chan bool) {

	for {
		select {
		case button := <-buttonsCh:
			fmt.Printf("%+v\n", button)
			OnRequestButtonPress(button.Floor, button.Button)

		case floor := <-floorsCh:
			OnFloorArrival(floor)

		case obstr := <-obstrCh:
			fmt.Printf("Obstruction is %+v\n", obstr)
			if obstr {
				elevio.SetMotorDirection(elevio.Stop)
			} // else {
			// elevio.SetMotorDirection(elev.Dirn)
			//}

		case stop := <-stopCh:
			fmt.Printf("%+v\n", stop)
			for f := 0; f < elevio.N_FLOORS; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(f, b, false)
				}
			}
		}

		/* Vi kaller jo på denne funksjonen i main.go
		if timer.TimedOut() {
			timer.Stop()
			OnDoorTimeout()
		}*/
	}
}

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(elev)
}

func CheckForTimeout() bool {
	if timer.TimedOut() {
		timer.Stop()
		OnDoorTimeout()
		return true
	}
	return false
}

func CheckForTimeout_continously() {
	for {
		if timer.TimedOut() {
			timer.Stop()
			OnDoorTimeout()
		}
	}

}

func SetAllLights(es elevator.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(floor, elevio.ButtonType(btn), es.Requests[floor][btn])
		}
	}
}

func OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.Down)
	elev.Dirn = elevio.Down
	elev.State = elevator.Moving
}

func OnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\n%s(%d, %s)\n", "fsmOnRequestButtonPress", btnFloor, elevator.ButtonToString(btnType))

	switch elev.State {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(elev, btnFloor, btnType) {
			timer.Start(elev.Config.DoorOpenDuration)
		} else {
			elev.Requests[btnFloor][btnType] = true
		}

	case elevator.Moving:
		elev.Requests[btnFloor][btnType] = true

	case elevator.Idle:
		elev.Requests[btnFloor][btnType] = true
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State

		switch pair.State {
		case elevator.DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.Start(elev.Config.DoorOpenDuration)
			elev = requests.ClearAtCurrentFloor(elev)

		case elevator.Moving:
			elevio.SetMotorDirection(elev.Dirn)

		case elevator.Idle:
			break
		}
	}

	SetAllLights(elev)
	fmt.Printf("\nNew state:\n")
}

func OnFloorArrival(newFloor int) {
	fmt.Printf("\n\n%s(%d)\n", "Arrival on floor ", newFloor)

	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.State {
	case elevator.Moving:
		if requests.ShouldStop(elev) {
			elevio.SetMotorDirection(elevio.Stop)
			elevio.SetDoorOpenLamp(true)
			elev = requests.ClearAtCurrentFloor(elev)
			timer.Start(elev.Config.DoorOpenDuration)
			SetAllLights(elev)
			elev.State = elevator.DoorOpen
		}
	default:
		break
	}

	fmt.Printf("\nNew state:\n")
}

func OnDoorTimeout() {
	fmt.Printf("\n\n%s()\n", "Door timeout")

	switch elev.State {
	case elevator.DoorOpen:
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State

		switch elev.State {
		case elevator.DoorOpen:
			timer.Start(elev.Config.DoorOpenDuration)
			elev = requests.ClearAtCurrentFloor(elev)
			SetAllLights(elev)

		case elevator.Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)

		case elevator.Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
		}

	default:
		break
	}

	fmt.Printf("\nNew state:\n")
}