package elevator_state

type CallType int

const (
    hallUp CallType = iota
    hallDown
    cab
)

type HallCallType int

const (
    up HallCallType = iota
    down
)

type Dirn int

const (
    down Dirn = -1
    stop
    up
)

type ElevatorBehaviour int

const (
    idle ElevatorBehaviour = iota
    moving
    doorOpen
)

type LocalElevatorState struct {
    Behaviour   ElevatorBehaviour
    Floor       int
    Direction   Dirn
    CabRequests []bool
}

func NewLocalElevatorState(behaviour ElevatorBehaviour, floor int, direction Dirn, cabRequests []bool) LocalElevatorState {
    return LocalElevatorState{
        Behaviour:  behaviour,
        Floor:      floor,
        Direction:  direction,
        CabRequests: cabRequests,
    }
}

type ElevatorState struct {
    Behaviour   ElevatorBehaviour
    Floor       int
    Direction   Dirn
    Requests    [][]bool
}

func NewElevatorState(behaviour ElevatorBehaviour, floor int, direction Dirn, requests [][]bool) ElevatorState {
    return ElevatorState{
        Behaviour:  behaviour,
        Floor:      floor,
        Direction:  direction,
        Requests:   requests,
    }
}

func Local(e ElevatorState) LocalElevatorState {
    return LocalElevatorState{
        Behaviour:  e.Behaviour,
        Floor:      e.Floor,
        Direction:  e.Direction,
        CabRequests: make([]bool, len(e.Requests[0])),
    }
}

func WithRequests(e LocalElevatorState, hallReqs [][]bool) ElevatorState {
    var cabReqs = make([]bool, len(e.CabRequests))
    for i := range cabReqs {
        cabReqs[i] = e.CabRequests[i]
    }

    var newRequests [][]bool
    for i, _ := range hallReqs {
        combined := append(hallReqs[i], cabReqs...)
        newRequests = append(newRequests, combined)
    }

    return ElevatorState{
        Behaviour:  e.Behaviour,
        Floor:      e.Floor,
        Direction:  e.Direction,
        Requests:   newRequests,
    }
}

func HallRequests(e ElevatorState) [][]bool {
    var hallReqs [][]bool
    for _, req := range e.Requests {
        hallReqs = append(hallReqs, req[:2])
    }
    
    return hallReqs
}