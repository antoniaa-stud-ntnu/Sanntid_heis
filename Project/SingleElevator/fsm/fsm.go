package fsm

import (
	"fmt"
	"Project/SingleElevator/elevator"
	"Project/SingleElevator/elevio"
	"Project/SingleElevator/requests"
	"Project/SingleElevator/timer"
)

var currentElevator elevator.Elevator;
var outputDevice elevio.ElevOutputDevice;

// Need to look more on this function
func fsmInit() {
	currentElevator = elevator.elevatorUninitialized(currentElevator);
    // outputDevice = getOutputDevice();  // Do we need this?
}

func SetAllLights(es elevator.Elevator){
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(floor, elevio.ButtonType(btn), es.Requests[floor][btn])
		}
	}
}

func FsmOnInitBetweenFloors() {
	// outputDevice.MotorDirection(MD_Down)
	elevio.SetMotorDirection(elevio.MD_Down)
	currentElevator.Dirn = elevio.MD_Down
	currentElevator.Behaviour = elevator.EB_Moving
}

func FsmOnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	fmt.Printf("\n\n%s(%d, %s)\n", "fsmOnRequestButtonPress", btnFloor, elevator.ButtonToString(btnType));
    elevator.ElevatorPrint(currentElevator);

	switch currentElevator.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.ShouldClearImmediately(currentElevator, btnFloor, btnType) {
			timer.TimerStart(5)
		} else {
			currentElevator.Requests[btnFloor][btnType] = true
		}
		break
	case elevator.EB_Moving:
		currentElevator.Requests[btnFloor][btnType] = true
		break

	case elevator.EB_Idle:    
        currentElevator.Requests[btnFloor][btnType] = true
        var pair requests.DirnBehaviourPair = requests.ChooseDirection(currentElevator)
        currentElevator.Dirn = pair.Dirn
        currentElevator.Behaviour = pair.Behaviour

        switch pair.Behaviour {
        case elevator.EB_DoorOpen:
            outputDevice.DoorLight = true
            timer.TimerStart(5)
            currentElevator = requests.ClearAtCurrentFloor(currentElevator)
			break

        case elevator.EB_Moving:
            outputDevice.MotorDirection(currentElevator.Dirn)
			break

        case elevator.EB_Idle:
            break
        }
	}

    SetAllLights(currentElevator)
    fmt.Printf("\nNew state:\n")
    elevator.ElevatorPrint(currentElevator)
}

func FsmOnFloorArrival(newFloor int) {
	fmt.Printf("\n\n%s(%d)\n", "Arrival on floor ", newFloor)

	currentElevator.Floor = newFloor
	elevio.SetFloorIndicator(currentElevator.Floor)

	switch currentElevator.Behaviour {
	case elevator.EB_Moving:
		if requests.ShouldStop(currentElevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.setDoorOpenLamp(true)
			currentElevator = requests.ClearAtCurrentFloor(currentElevator)
			timer.TimerStart(5)
			SetAllLights(currentElevator)
			currentElevator.Behaviour = elevator.EB_DoorOpen
		}
		break

	default:
		break
	}

	fmt.Printf("\nNew state:\n")
}

func FsmOnDoorTimeout() {
	fmt.Printf("\n\n%s()\n", "Door timeout")

	switch currentElevator.Behaviour {
	case elevator.EB_DoorOpen:
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(currentElevator)
		currentElevator.Dirn = pair.Dirn
		currentElevator.Behaviour = pair.Behaviour

		switch currentElevator.Behaviour {
		case elevator.EB_DoorOpen:
			timer.TimerStart(5)
			currentElevator = requests.ClearAtCurrentFloor(currentElevator)
			SetAllLights(currentElevator)
			break
		case elevator.EB_Moving, elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(currentElevator.Dirn)
			break
		}

	default:
		break
	}

	fmt.Printf("\nNew state:\n")
}