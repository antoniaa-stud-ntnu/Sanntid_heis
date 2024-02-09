package elevator

import (
	"Project/SingleElevator/elevio"
)

type ElevatorBehaviour int

const (
    EB_Idle     ElevatorBehaviour = 0
    EB_DoorOpen ElevatorBehaviour = 1
    EB_Moving   ElevatorBehaviour = 2
)

type ClearRequestVariant int

const (
    CV_All    ClearRequestVariant = 0
    CV_InDirn ClearRequestVariant = 1
)

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDurationS   float64
}

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests[elevio.N_FLOORS][elevio.N_BUTTONS] bool
	Behaviour ElevatorBehaviour
	Config    Config
}

func ebToString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "EB_Idle"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	case EB_Moving:
		return "EB_moving"
	default:
		return "EB_UNDEFINED"
	}
}

func dirnToString(dir elevio.MotorDirection) string {
    switch dir {
    case elevio.MD_up:
        return "MD_Up"
    case elevio.MD_Down:
        return "MD_Down"
    case elevio.MD_Stop:
        return "MD_Stop"
    default:
        return "MD_UNDEFINED"
    }
}

func ButtonToString(btn elevio.ButtonType) string {
    switch btn {
    case elevio.BT_HallUp:
        return "BT_HallUp"
    case elevio.BT_HallDown:
        return "BT_HallDown"
    case elevio.BT_Cab:
        return "BT_Cab"
    default:
        return "BT_UNDEFINED"
    }
}

func ElevatorPrint(es Elevator) {
	fmt.Printf(" +--------------------+\n")
	fmt.Printf(
		" |floor = %-2d        |\n"+
		"  |dirn  = %-12.12s|\n"+
		"  |behav = %-12.12s|\n",
		es.Floor,
		dirnToString(es.Dirn),
		ebToString(es.Behaviour),
	)
	fmt.Printf(" +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for f := elevio.N_FLOORS-1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if (f == elevio.N_FLOORS-1 && btn == elevio.BT_HallUp) || (f == 0 && btn == elevio.BT_HallDown) {
				fmt.Printf("|     ")
			} else {
				if es.Requests[f][btn] {
					fmt.Printf("|  #  ")
				} else {
					fmt.Printf("|  -  ")
				}
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Printf(" +--------------------+\n")
}

func elevatorUninitialized(es Elevator) {
	es.Floor = -1,
	es.Dirn = elevio.MD_Stop,
	es.Behaviour = EB_Idle,
	es.Config = Config {
		ClearRequestVariant: CV_All,
		DoorOpenDurationS:  3.0,
	}
}