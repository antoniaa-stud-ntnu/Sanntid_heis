package requests

import (
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
)

type DirnBehaviourPair struct {
	Dirn  elevio.MotorDirection
	State elevator.Behaviour
}

func requestsAbove(e elevator.Elevator) bool {
	for floor := e.Floor + 1; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(e elevator.Elevator) bool {
	for floor := 0; floor < e.Floor; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Requests[floor][btn] {
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
	case elevio.Up:
		if requestsAbove(e) {
			return DirnBehaviourPair{elevio.Up, elevator.Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.Down, elevator.DoorOpen}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.Down, elevator.Moving}
		} else {
			return DirnBehaviourPair{elevio.Stop, elevator.Idle}
		}
	case elevio.Down:
		if requestsBelow(e) {
			return DirnBehaviourPair{elevio.Down, elevator.Moving}
		} else if requestsHere(e) {
			return DirnBehaviourPair{elevio.Up, elevator.DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.Up, elevator.Moving}
		} else {
			return DirnBehaviourPair{elevio.Stop, elevator.Idle}
		}
	case elevio.Stop:
		if requestsHere(e) {
			return DirnBehaviourPair{elevio.Stop, elevator.DoorOpen}
		} else if requestsAbove(e) {
			return DirnBehaviourPair{elevio.Up, elevator.Moving}
		} else if requestsBelow(e) {
			return DirnBehaviourPair{elevio.Down, elevator.Moving}
		} else {
			return DirnBehaviourPair{elevio.Stop, elevator.Idle}
		}
	default:
		return DirnBehaviourPair{elevio.Stop, elevator.Idle}
	}
}

func ShouldStop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.Down:
		return e.Requests[e.Floor][elevio.HallDown] || e.Requests[e.Floor][elevio.Cab] || !requestsBelow(e)
	case elevio.Up:
		return e.Requests[e.Floor][elevio.HallUp] || e.Requests[e.Floor][elevio.Cab] || !requestsAbove(e)
	default:
		return true
	}
}

func ShouldClearRequest(e elevator.Elevator) elevio.ButtonType {
	if e.Requests[e.Floor][elevio.HallDown] {
		return elevio.HallDown
	} else if e.Requests[e.Floor][elevio.HallUp] {
		return elevio.HallUp
	} else if e.Requests[e.Floor][elevio.Cab]{
		return elevio.Cab
	}
	return -1
}

func ShouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_All:
		return e.Floor == btn_floor
	case elevator.CV_InDirn:
		return e.Floor == btn_floor && ((e.Dirn == elevio.Up && btn_type == elevio.HallUp) ||
			(e.Dirn == elevio.Down && btn_type == elevio.HallDown) ||
			e.Dirn == elevio.Stop || btn_type == elevio.Cab)
	default:
		return false
	}
}

func ClearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_All:
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case elevator.CV_InDirn:
		e.Requests[e.Floor][elevio.Cab] = false
		switch e.Dirn {
		case elevio.Up:
			if !requestsAbove(e) && !e.Requests[e.Floor][elevio.HallUp] {
				e.Requests[e.Floor][elevio.HallDown] = false
			}
			e.Requests[e.Floor][elevio.HallUp] = false
		case elevio.Down:
			if !requestsAbove(e) && !e.Requests[e.Floor][elevio.HallDown] {
				e.Requests[e.Floor][elevio.HallUp] = false
			}
			e.Requests[e.Floor][elevio.HallDown] = false
		case elevio.Stop:
		default:
			e.Requests[e.Floor][elevio.HallUp] = false
			e.Requests[e.Floor][elevio.HallDown] = false
		}
	default:
	}
	return e
}
