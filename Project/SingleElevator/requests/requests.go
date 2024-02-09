package requests

import (
	"Project/SingleElevator/elevator"
	"Project/SingleElevator/elevio"
)

type DirnBehaviourPair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevator.ElevatorBehaviour
}

func requestsAbove(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requestsHere(e elevator.Elevator) bool {
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirection(e elevator.Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	case elevio.MD_Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	case elevio.MD_Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle} // Cannot write this
	}
}

func ShouldStop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab] || !requestsBelow(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_Cab] || !requestsAbove(e)
	case elevio.MD_Stop: // Maybe don't need this? Someone don't
		fallthrough // Maybe don't need this? Someone don't
	default:
		return true
	}
}

func ShouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.config.clearRequestVariant {
	case elevator.CV_All:
		return e.Floor == btn_floor
	case elevator.CV_InDirn:
		return e.Floor == btn_floor && ((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab)
	default:
		return false
	}
}

func ClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {
	switch e.config.clearRequestVariant {
	case elevator.CV_All:
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case elevator.CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			if !requestsAbove(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			break
		case elevio.MD_Down:
			if !requestsAbove(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			break
		case elevio.MD_Stop:
			break
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
			break
		}
		break
	default:
		break
	}
	return e
}
