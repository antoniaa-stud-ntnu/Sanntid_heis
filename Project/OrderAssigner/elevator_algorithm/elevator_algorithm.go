package elevator_algorithm

import (
	"fmt"
	"./elevator_state"
)

func requestsAbove(e ElevatorState) bool {
	return e.Requests[e.Floor+1:].Map(func(a ElevatorState) bool {
		return a.Array.Any()
	}).Any()
}

func requestsBelow(e ElevatorState) bool {
	return e.Requests[:e.Floor].Map(func(a ElevatorState) bool {
		return a.Array.Any()
	}).Any()
}

func anyRequests(e ElevatorState) bool {
	return e.Requests.Map(func(a ElevatorState) bool {
		return a.Array.Any()
	}).Any()
}

func anyRequestsAtFloor(e ElevatorState) bool {
	return e.Requests[e.Floor].Array.Any()
}

func shouldStop(e ElevatorState) bool {
	switch e.Direction {
	case Up:
		return e.Requests[e.Floor][HallUp] ||
			e.Requests[e.Floor][Cab] ||
			!requestsAbove(e) ||
			e.Floor == 0 ||
			e.Floor == len(e.Requests)-1
	case Down:
		return e.Requests[e.Floor][HallDown] ||
			e.Requests[e.Floor][Cab] ||
			!requestsBelow(e) ||
			e.Floor == 0 ||
			e.Floor == len(e.Requests)-1
	case Stop:
		return true
	default:
		return false
	}
}

func chooseDirection(e ElevatorState) Dirn {
	switch e.Direction {
	case Up:
		if e.RequestsAbove {
			return Up
		} else if e.AnyRequestsAtFloor {
			return Stop
		} else if e.RequestsBelow {
			return Down
		} else {
			return Stop
		}
	case Down, Stop:
		if e.RequestsBelow {
			return Down
		} else if e.AnyRequestsAtFloor {
			return Stop
		} else if e.RequestsAbove {
			return Up
		} else {
			return Stop
		}
	default:
		return Stop
	}
}

func clearReqsAtFloor(e ElevatorState, onClearedRequest func(CallType) = nil) ElevatorState {
	e2 := e

	clear := func(c CallType) {
		if e2.Requests[e2.Floor][c] {
			if onClearedRequest != nil {
				onClearedRequest(c)
			}
			e2.Requests[e2.Floor][c] = false
		}
	}

	switch ClearRequestType {
	case All:
		for c := CallTypeMin; c < len(e2.Requests[0]); c++ {
			clear(c)
		}
	case InDirn:
		clear(Cab)

		switch e.Direction {
		case Up:
			if e2.Requests[e2.Floor][HallUp] {
				clear(HallUp)
			} else if !requestsAbove(e2) {
				clear(HallDown)
			}
		case Down:
			if e2.Requests[e2.Floor][HallDown] {
				clear(HallDown)
			} else if !requestsBelow(e2) {
				clear(HallUp)
			}
		case Stop:
			clear(HallUp)
			clear(HallDown)
		}
	}

	return e2
}