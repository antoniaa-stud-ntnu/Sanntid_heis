package fsm

import (
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/requests"
	"Project/singleElevator/tcp"
	"Project/singleElevator/timer"
	"Project/network/udpBroadcasr/udpNetwork/localip"

	"encoding/json"
	"fmt"
	"net"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type msgElevState struct {
	ipAddr		string
	elevState	HRAElevState
}

type msgHallReq struct {
	tAdd_fRemove 	bool
	floor			int
	button			elevio.ButtonType
}

func SendElevState(msgEl msgElevState, conn net.Conn) {
	jsonElevState, _ := json.Marshal(msgEl)
	tcp.TCPSendMessage(conn, jsonElevState)
}

func SendHallRequests(msgReq msgHallReq, conn net.Conn){
	jsonHallReq, _ := json.Marshal(msgReq)
	tcp.TCPSendMessage(conn, jsonHallReq)
}

var elev elevator.Elevator = elevator.InitElev()
var hraElevState HRAElevState
//var hallReqStruct msgHallReq


func FSM(buttonsCh chan elevio.ButtonEvent, floorsCh chan int, hraOutputCh chan [][2]bool, obstrCh chan bool, lightsCh chan [][2]bool, conn net.Conn) {
	hraElevState.Direction = elevator.DirnToString(elev.Dirn)
	hraElevState.Behavior = elevator.EbToString(elev.State) // burde vi initialisere .Floor og CabReq
	myIP, _ := localip.LocalIP();
	for {
		select {
		case button := <-buttonsCh:
			fmt.Printf("%+v\n", button)
			if button.ButtonType == elevio.Cab {
				hraElevState.CabRequests[button.floor] = true
				SendElevState(myIP, hraElevState, conn)
				OnRequestButtonPress(button.floor, button.ButtonType) // dersom det er cabReq så skal den ta bestillingen selv
			} else {
				hallReq := msgHallReq{true, button.floor, button.ButtonType}
				SendHallRequests(hallReq, conn)
			} 
		case floor := <-floorsCh:
			hraElevState.Floor = floor
			SendElevState(myIP, hraElevState, conn)
			OnFloorArrival(floor)
			elevator.ElevatorPrint(elev)
		case obstr := <-obstrCh:
			fmt.Printf("Obstruction is %+v\n", obstr)
			OnObstruction(obstr)
			elevator.ElevatorPrint(elev)
		case newHallRequests := <-hraOutputCh:
			for floor := 0; floor < len(newHallRequests); floor++ {
				for hallIndex := 0; hallIndex < len(newHallRequests[hallIndex]); hallIndex++ {
					value := newHallRequests[floor][hallIndex]
					if value { 
						OnRequestButtonPress(floor, hallIndex) // legger til hallrequests fra Master og tar ordrene
					}
				}
			}
			elevator.ElevatorPrint(elev)
		case lights := <- lightsCh:
			SetLights(elev, lights)
		}
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

func CheckForDoorTimeOut() {
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

func SetLights(es elevator.Elevator, hallReq [][2]bool) { // maa sende inn hallReq som input
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ { 
			if btn == elevio.ButtonType.Cab {
				elevio.SetButtonLamp(floor, elevio.ButtonType(btn), es.Requests[floor][btn])
			} else {
				elevio.SetButtonLamp(floor, elevio.ButtonType(btn), hallReq[floor][btn])
			}
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

	// SetAllLights(elev)
	fmt.Printf("\nNew state:\n")
}

func OnFloorArrival(newFloor int) {
	fmt.Printf("\n\n%s(%d)\n", "Arrival on floor ", newFloor)

	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.State {
	case elevator.Moving:
		if requests.ShouldStop(elev) { //I en etasje med request
			elevio.SetMotorDirection(elevio.Stop)
			elevio.SetDoorOpenLamp(true)
			//If hall request: send ordre fullført til primary
			elev = requests.ClearAtCurrentFloor(elev)
			timer.Start(elev.Config.DoorOpenDuration)
			// SetAllLights(elev)
			elev.State = elevator.DoorOpen
		}
	default:
		break
	}

	fmt.Printf("\nNew state:\n")
	//Send state to master
}

func OnDoorTimeout() {
	fmt.Printf("\n\n%s()\n", "Door timeout")

	switch elev.State {
	case elevator.DoorOpen:
		if elev.ObstructionActive {
			timer.Start(elev.Config.DoorOpenDuration)
			break
		}
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State

		switch elev.State {
		case elevator.DoorOpen:
			timer.Start(elev.Config.DoorOpenDuration)
			elev = requests.ClearAtCurrentFloor(elev)
			// SetAllLights(elev)

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
	//Send state to master
}

func OnObstruction(obstructionState bool) {
	elev.ObstructionActive = obstructionState
}


//func checkChannels(buttonsCh chan elevio.ButtonEvent, floorsCh chan int) {
	// hraElevState.Direction = elevator.DirnToString(elev.Dirn)
	// hraElevState.Behavior = elevator.EbToString(elev.State)

	// for {
	// 	select {
	// 	case button := <-buttonsCh:
	// 		fmt.Printf("%+v\n", button)
	// 		if button.ButtonType == elevio.Cab {
	// 			hraElevState.CabRequests[button.f] = true
	// 		} else if button.ButtonType == elevio.HallUp {
	// 			hraHallReq[button.f][0] = true
	// 		} else if button.ButtonType == elevio.HallDown {
	// 			hraHallReq[button.f][1] = true
	// 		} // Kall paa sendElevUpdate til master

	// 		SendElevState(hraElevState)
	// 		// FSM(button)

	// 	case floor := <-floorsCh:
	// 		hraElevState.Floor = floor
	// 		SendElevState(hraElevState)
	// 		OnFloorArrival(floor)
	// 		// Kall paa sendElevUpdate til master
	// 		//case obstr := <-obstrCh:
	// 		//fmt.Printf("Obstruction is %+v\n", obstr)
	// 		//OnObstruction(obstr)
	// 		//elevator.ElevatorPrint(elev)
	// 	}
	// }
// }

// Endre slik at det ikke er FSM som tar inn channels som input, skal kunne kalles dersom master gir deg beskjed om det
// Trenger altsaa et mellomledd, slik at statene sendes til master dersom noe skjer paa channels.

