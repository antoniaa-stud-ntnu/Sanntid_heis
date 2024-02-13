package main

import (
	"fmt"
	"time"
)

type m_Req struct {
	Active     bool
	AssignedTo string
}

type State struct {
	ID    string
	State LocalElevatorState
	Time  time.Duration
}

func isUnassigned(r m_Req) bool {
	return r.Active && r.AssignedTo == ""
}

func filterReq(reqs [][]m_Req, fn func(m_Req) bool) [][]bool {
	filteredReqs := make([][]bool, len(reqs))
	for f, reqsAtFloor := range reqs {
		filteredReqs[f] = make([]bool, 2)
		for b, req := range reqsAtFloor {
			filteredReqs[f][b] = fn(req)
		}
	}
	return filteredReqs
}

func toReq(hallReqs [][]bool) [][]m_Req {
	reqs := make([][]m_Req, len(hallReqs))
	for f, reqsAtFloor := range hallReqs {
		reqs[f] = make([]m_Req, 2)
		for b, req := range reqsAtFloor {
			reqs[f][b] = m_Req{req, ""}
		}
	}
	return reqs
}

func withReqs(s State, reqs [][]m_Req, fn func(m_Req) bool) ElevatorState {
	e := s.State.WithRequests(reqs.filterReq(fn))
	return e
}

func anyUnassigned(reqs [][]m_Req) bool {
	for _, reqsAtFloor := range reqs {
		for _, req := range reqsAtFloor {
			if isUnassigned(req) {
				return true
			}
		}
	}
	return false
}

func initialStates(states map[string]LocalElevatorState) []State {
	initialStates := make([]State, len(states))
	i := 0
	for id, state := range states {
		initialStates[i] = State{id, state, time.Duration(state.Time)}
		i++
	}
	return initialStates
}

func performInitialMove(s *State, reqs *[][]m_Req) {
	fmt.Printf("initial move: %s\n", *s)
	switch s.State.Behaviour {
	case DoorOpen:
		fmt.Printf("  '%s' closing door at floor %d\n", s.ID, s.State.Floor)
		s.Time += doorOpenDuration / 2
		goto case Idle
	case Idle:
		for c := 0; c < 2; c++ {
			if (*reqs)[s.State.Floor][c].Active {
				(*reqs)[s.State.Floor][c].AssignedTo = s.ID
				s.Time += doorOpenDuration
				fmt.Printf("  '%s' taking req %d at floor %d\n", s.ID, c, s.State.Floor)
			}
		}
	case Moving:
		s.State.Floor += int(s.State.Direction)
		s.Time += travelDuration / 2
		fmt.Printf("  '%s' arriving at %d\n", s.ID, s.State.Floor)
	}
}

func performSingleMove(s *State, reqs *[][]m_Req) {
	e := withReqs(*s, *reqs, isUnassigned)
	fmt.Println(e)

	onClearRequest := func(c CallType) {
		switch c {
		case HallUp, HallDown:
			(*reqs)[s.State.Floor][c].AssignedTo = s.ID
		case Cab:
			s.State.CabRequests[s.State.Floor] = false
		}
	}

	switch s.State.Behaviour {
	case Moving:
		if e.ShouldStop {
			s.State.Behaviour = DoorOpen
			s.Time += doorOpenDuration
			e.ClearReqsAtFloor(onClearRequest)
			fmt.Printf("  '%s' stopping at %d\n", s.ID, s.State.Floor)
		} else {
			s.State.Floor += s.State.Direction
			s.Time += travelDuration
			fmt.Printf("  '%s' continuing to %d\n", s.ID, s.State.Floor)
		}
	case Idle, DoorOpen:
		s.State.Direction = e.ChooseDirection()
		if s.State.Direction == DirnStop {
			if e.AnyRequestsAtFloor() {
				e.ClearReqsAtFloor(onClearRequest)
				s.Time += doorOpenDuration
				s.State.Behaviour = DoorOpen
				fmt.Printf("  '%s' taking req in opposite dirn at %d\n", s.ID, s.State.Floor)
			} else {
				s.State.Behaviour = Idle
				fmt.Printf("  '%s' idling at %d\n", s.ID, s.State.Floor)
			}
		} else {
			s.State.Behaviour = Moving
			s.State.Floor += s.State.Direction
			s.Time += travelDuration
			fmt.Printf("  '%s' departing %s to %d\n", s.ID, s.State.Direction, s.State.Floor)
		}
	}

	fmt.Println(withReqs(*s, *reqs, isUnassigned))
}

func unvisitedAreImmediatelyAssignable(reqs [][]m_Req, states []State) bool {
	for f, reqsAtFloor := range reqs {
		if len(reqsAtFloor) == 2 {
			return false
		}
		for _, req := range reqsAtFloor {
			if isUnassigned(req) {
				found := false
				for _, state := range states {
					if state.State.Floor == f && !
