package fsm

import (
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/requests"
	"Project/singleElevator/timer"
	"fmt"
	"net"
	"time"
)

//"Project/network/udpBroadcast/udpNetwork/localip"

const MasterPort = "20025"

var elev elevator.Elevator = elevator.InitElev()

var hraElevState = messages.HRAElevState{
	Behaviour:   elevator.EbToString(elev.State),  // hraElevState.Behavior = elevator.EbToString(elev.State)
	Floor:       elev.Floor,                       // hraElevState.Floor = elevio.GetFloor()
	Direction:   elevator.DirnToString(elev.Dirn), // hraElevState.Direction = elevator.DirnToString(elev.Dirn)
	CabRequests: elevator.GetCabRequests(elev),    // hraElevState.CabRequests = elevator.GetCabRequests(elev)
}

var msgElevState = messages.ElevStateMsg{
	IpAddr:    "",
	ElevState: hraElevState,
}

// hraElevState.Behavior = elevator.EbToString(elev.State)
// hraElevState.Floor = elevio.GetFloor()
// hraElevState.Direction = elevator.DirnToString(elev.Dirn)
// hraElevState.CabRequests = elevator.GetCabRequests(elev)

func FSM(buttonsCh chan elevio.ButtonEvent, floorsCh chan int, obstrCh chan bool, masterIPCh chan net.IP, jsonMessageCh chan []byte, toFSMCh chan []byte) {
	// Waiting for master to be set and connecting to it
	masterIP := <-masterIPCh
	time.Sleep(2 * time.Second)
	// Connecting to master
	//masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), MasterPort)

	localip, _ := localip.LocalIP()
	var masterConn net.Conn
	msgElevState.IpAddr = localip
	var err error
	/*
		if masterIP.String() == localip {
			//fmt.Println(masterIP.String(), localip)
			masterConn,_ = tcp.TCPMakeMasterConnection("localhost", MasterPort)
			go tcp.TCPRecieveMasterMsg(masterConn, jsonMessageCh)
		} else {
			masterConn, _ = tcp.TCPMakeMasterConnection(masterIP.String(), MasterPort)
			go tcp.TCPRecieveMasterMsg(masterConn, jsonMessageCh)
		}*/
	masterConn, _ = tcp.TCPMakeMasterConnection(masterIP.String(), MasterPort)
	go tcp.TCPRecieveMasterMsg(masterConn, jsonMessageCh)
	if err != nil {
		fmt.Println("Elevator could not connect to master:", err)
	}

	//fmt.Println("Masterconn is: ", masterConn)
	// Single elevators Finite State Machine
	for {
		select {
		case button := <-buttonsCh:
			// Different approces to cab request and hall request
			//fmt.Printf("%+v\n", button)
			if button.Button == elevio.Cab {
				// Updating master with the new cab request in the elevator
				msgElevState.ElevState.CabRequests[button.Floor] = true
				sendingBytes := messages.ToBytes(messages.MsgElevState, msgElevState)
				//fmt.Println(string(sendingBytes))
				tcp.TCPSendMessage(masterConn, sendingBytes)

				// Handling the cab request
				OnRequestButtonPress(button.Floor, button.Button) // dersom det er cabReq så skal den ta bestillingen selv
				elevio.SetButtonLamp(button.Floor, button.Button, true)
			} else {
				// Sending the hall request to master
				hallReq := messages.HallReqMsg{true, button.Floor, button.Button}
				sendingBytes := messages.ToBytes(messages.MsgHallReq, hallReq)
				tcp.TCPSendMessage(masterConn, sendingBytes)
				//fmt.Println("Sent hall request to master: ", button.Floor, button.Button)
			}
		case floor := <-floorsCh:
			// Updating master with the new state of the elevator
			msgElevState.ElevState.Floor = floor
			sendingBytes := messages.ToBytes(messages.MsgElevState, msgElevState)
			tcp.TCPSendMessage(masterConn, sendingBytes)

			// Handeling floor change and telling master if a hall request was removed
			OnFloorArrival(floor, masterConn)

			//elevator.ElevatorPrint(elev)

		case obstr := <-obstrCh:
			// Reacting to the obstruction button
			fmt.Printf("Obstruction is %+v\n", obstr)
			OnObstruction(obstr)
			//elevator.ElevatorPrint(elev)
		case toFSM := <-toFSMCh:
			// Handeling messages from master
			msgType, data := messages.FromBytes(toFSM)
			switch msgType {
			case messages.MsgAssignedHallReq:
				// Recieve hall requests from master
				newHallRequests := data.([][2]bool) // Dette er ikke nodvendig egentlig
				for floor := 0; floor < len(newHallRequests); floor++ {
					for hallIndex := 0; hallIndex < len(newHallRequests[hallIndex]); hallIndex++ {
						// Overwriting elevators hall requests as master could have given an order to another elevator
						value := newHallRequests[floor][hallIndex]
						if value {
							OnRequestButtonPress(floor, elevio.ButtonType(hallIndex)) // legger til hallrequests fra Master og tar ordrene
						} else {
							//Mutex lock??
							elev.Requests[floor][hallIndex] = false
						}
					}
				}
				//elevator.ElevatorPrint(elev)
			case messages.MsgHallLigths:
				// Setting hall lights as master says
				elevio.SetButtonLamp(data.(messages.HallReqMsg).Floor, data.(messages.HallReqMsg).Button, data.(messages.HallReqMsg).TAddFRemove)
			}

		case masterIP := <-masterIPCh:
			// Master has changed, need to make new connection
			masterConn.Close()
			masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), MasterPort)
			if err != nil {
				fmt.Println("Elevator could not connect to master:", err)
			}
			go tcp.TCPRecieveMasterMsg(masterConn, jsonMessageCh)
			//iPToConnMap := make(map[net.Addr]net.Conn)
			//masterip, _ := net.ResolveIPAddr("ip", masterIP.String()) // String til net.Addr
			//iPToConnMap[masterip] = masterConn
			//go tcp.TCPReciveMessage(masterConn, jsonMessageCh, &iPToConnMap)
			fmt.Println("Master has changed")
		}
	}
}

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(elev) // endres ???
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

// Lysfunksjon med buttonType, floor, og on/off
func SetLight(btn elevio.ButtonType, floor int, onOrOff bool) {
	elevio.SetButtonLamp(floor, elevio.ButtonType(btn), onOrOff)
}

func OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.Down)
	elev.Dirn = elevio.Down
	elev.State = elevator.Moving
}

func OnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {
	//fmt.Printf("\n\n%s(%d, %s)\n", "fsmOnRequestButtonPress", btnFloor, elevator.ButtonToString(btnType))

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
			elev, _ = requests.ClearAtCurrentFloor(elev)

		case elevator.Moving:
			elevio.SetMotorDirection(elev.Dirn)

		case elevator.Idle:
			break
		}
	}

	// SetAllLights(elev)
}

func OnFloorArrival(newFloor int, masterConn net.Conn) {
	//fmt.Printf("\n\n%s(%d)\n", "Arrival on floor ", newFloor)

	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.State {
	case elevator.Moving:
		if requests.ShouldStop(elev) { //I en etasje med request eller ingen requests over/under
			elevio.SetMotorDirection(elevio.Stop)
			elevio.SetDoorOpenLamp(true)
			/*
				for btnIndex, btnValue := range elev.Requests[elev.Floor] {
					buttonType := elevio.ButtonType(btnIndex)
					if btnValue && ((buttonType == elevio.HallDown && elev.Dirn == elevio.Down) || (buttonType == elevio.HallUp && elev.Dirn == elevio.Up)) {
						removeHallReq := messages.HallReqMsg{false, newFloor, buttonType}
						sendingBytes := messages.ToBytes(messages.MsgHallReq, removeHallReq)
						tcp.TCPSendMessage(masterConn, sendingBytes)
					} else if btnValue && buttonType == elevio.Cab{
						elevio.SetButtonLamp(elev.Floor, buttonType, false)
					}

				}*/
			/*
				// Updating the button light contract with the master and the elevator
				for requestsOnfloor := range elev.Requests[elev.Floor] {
					if requestsOnfloor[elevio.HallDown] || requestsOnfloor[elevio.HallUp]{
						removeHallReq := messages.HallReqMsg{false, newFloor, button}
						sendingBytes := messages.ToBytes(messages.MsgHallReq, removeHallReq)
						tcp.TCPSendMessage(masterConn, sendingBytes)
					}
					if requestsOnfloor[elevio.Cab]{

					}
				}


				hallButton := requests.ShouldClearRequest(elev)
				if hallButton == elevio.HallDown || hallButton == elevio.HallUp{
					removeHallReq := messages.HallReqMsg{false, newFloor, hallButton}
					sendingBytes := messages.ToBytes(messages.MsgHallReq, removeHallReq)
					tcp.TCPSendMessage(masterConn, sendingBytes)
				} else if hallButton == elevio.Cab{
					elevio.SetButtonLamp(elev.Floor, hallButton, false)
				}
			*/
			var removingHallButtons [2]bool
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			for btnIndex, btnValue := range removingHallButtons {
				buttonType := elevio.ButtonType(btnIndex)
				if btnValue {
					removeHallReq := messages.HallReqMsg{false, newFloor, buttonType}
					sendingBytes := messages.ToBytes(messages.MsgHallReq, removeHallReq)
					tcp.TCPSendMessage(masterConn, sendingBytes)
					fmt.Println("Sending to master that hall req is complete: ", string(sendingBytes))
				}

			}

			timer.Start(elev.Config.DoorOpenDuration)
			// SetAllLights(elev)
			SetAllCabLights(elev)
			elev.State = elevator.DoorOpen
		}
	default:
		break
	}

}

func OnDoorTimeout() {
	//fmt.Printf("\n\n%s()\n", "Door timeout")

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
			elev, _ = requests.ClearAtCurrentFloor(elev)
			// SetAllLights(elev)
			SetAllCabLights(elev)
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
}

func OnObstruction(obstructionState bool) {
	elev.ObstructionActive = obstructionState
}

func SetAllCabLights(e elevator.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		elevio.SetButtonLamp(floor, elevio.Cab, e.Requests[floor][elevio.Cab])
	}
}
