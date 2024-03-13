package singleElevatorFSM

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	//"Project/localElevator/timer"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/roleHandler/master"
	"fmt"
	"net"
	"time"
)

//"Project/network/udpBroadcast/udpNetwork/localip"
var motorStopTimer *time.Timer
var doorOpenTimer *time.Timer

var elev elevator.Elevator = elevator.InitElev()


var hraElevState = messages.HRAElevState{
	Behaviour:   elevator.EbToString(elev.State),
	Floor:       elev.Floor,
	Direction:   elevator.DirnToString(elev.Dirn),
	CabRequests: elevator.GetCabRequests(elev),
}

var msgElevState = messages.ElevStateMsg{
	IpAddr:    "",
	ElevState: hraElevState,
}

func FSM(buttonsCh chan elevio.ButtonEvent, floorsCh chan int, obstrCh chan bool, masterIPCh chan net.IP, msgToMasterCh chan []byte, toFSMCh chan []byte, quitCh chan bool, peerTxEnable chan bool) {
	// Waiting for master to be set and connecting to it
	// masterIP := <-masterIPCh
	// fmt.Println("Recieved master IP: ", masterIP)
	//time.Sleep(2 * time.Second)
	// Connecting to master
	//masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), MasterPort)

	localip, _ := localip.LocalIP()
	msgElevState.IpAddr = localip

	motorAndDoorTimerInit()
	
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

				// Send on Channel to main
				msgToMasterCh <- sendingBytes 
				// tcp.TCPSendMessage(masterConn, sendingBytes)

				// Handling the cab request
				OnRequestButtonPress(button.Floor, button.Button, peerTxEnable) // dersom det er cabReq så skal den ta bestillingen selv
				elevio.SetButtonLamp(button.Floor, button.Button, true)
			} else {
				// Sending the hall request to master
				hallReq := messages.HallReqMsg{true, button.Floor, button.Button}
				sendingBytes := messages.ToBytes(messages.MsgHallReq, hallReq)
				//tcp.TCPSendMessage(masterConn, sendingBytes)
				msgToMasterCh <- sendingBytes 
				// fmt.Println("Sent hall request to master: ", button.Floor, button.Button)
			}
		case floor := <-floorsCh:
			// Updating master with the new state of the elevator
			msgElevState.ElevState.Floor = floor
			sendingBytes := messages.ToBytes(messages.MsgElevState, msgElevState)
			// tcp.TCPSendMessage(masterConn, sendingBytes)
			msgToMasterCh <- sendingBytes 

			// Handeling floor change and telling master if a hall request was removed
			OnFloorArrival(floor, msgToMasterCh, peerTxEnable)

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
							OnRequestButtonPress(floor, elevio.ButtonType(hallIndex), peerTxEnable) // legger til hallrequests fra Master og tar ordrene
							// fmt.Println("Checking if the request is true in elev.Requests: ", elev.Requests[floor][hallIndex])
						} else {
							elev.Requests[floor][hallIndex] = false
						}
					}
				}
				//elevator.ElevatorPrint(elev)
			case messages.MsgHallLigths:
				// Setting hall lights as master says
				elevio.SetButtonLamp(data.(messages.HallReqMsg).Floor, data.(messages.HallReqMsg).Button, data.(messages.HallReqMsg).TAddFRemove)

				// Restoring the cab requests
			case messages.MsgRestoreCabReq:
				for floor := 0; floor < len(data.([]bool)); floor++ {
					elev.Requests[floor][elevio.Cab] = data.([]bool)[floor]
				}
			}
		case masterIP := <-masterIPCh:
			fmt.Println("Master has changed to IP: ", masterIP.String())
			// Master has changed, need to make new connection
			masterConn.Close()
			masterConn, err := tcp.TCPMakeMasterConnection(masterIP.String(), master.MasterPort)
			if err != nil {
				fmt.Println("Elevator could not connect to master:", err)
			}
			fmt.Println("New masterConn is: ", masterConn)
			fmt.Println("Closing the old goroutine")
			quitCh <- true // Stopping the old goroutine
			go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, quitCh)
			//iPToConnMap := make(map[net.Addr]net.Conn)
			//masterip, _ := net.ResolveIPAddr("ip", masterIP.String()) // String til net.Addr
			//iPToConnMap[masterip] = masterConn
			//go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, &iPToConnMap)
			fmt.Println("Master has changed")
		case <- doorOpenTimer.C:
			fmt.Println("Door timed out")
			OnDoorTimeout(peerTxEnable)
		case <- motorStopTimer.C:
			fmt.Println("Motor timed out")
			peerTxEnable <- false
			//reset
		}
	}
}

func InitLights() {
	elevio.SetDoorOpenLamp(false)
	SetAllLights(elev) // endres ???
}



// Slettes ????
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

func OnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, peerTxEnable chan bool) {
	//fmt.Printf("\n\n%s(%d, %s)\n", "fsmOnRequestButtonPress", btnFloor, elevator.ButtonToString(btnType))

	switch elev.State {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(elev, btnFloor, btnType) {
			//timer.Start(elev.Config.DoorOpenDuration)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
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
			//timer.Start(elev.Config.DoorOpenDuration)
			
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			elev, _ = requests.ClearAtCurrentFloor(elev)

		case elevator.Moving:
			elevio.SetMotorDirection(elev.Dirn)
			startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)

		case elevator.Idle:
			break
		}
	}

	// SetAllLights(elev)
}

func OnFloorArrival(newFloor int, msgToMasterCh chan[]byte, peerTxEnable chan bool) {
	fmt.Printf("\n\n%s(%d)\n", "Arrival on floor ", newFloor)

	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	// Reset motor timer
	//startMotorStopTimer(time.Duration(elev.Config.MotorStopDuration) * time.Second, peerTxEnable)

	switch elev.State {
	case elevator.Moving:
		if requests.ShouldStop(elev) { //I en etasje med request eller ingen requests over/under
			elevio.SetMotorDirection(elevio.Stop)
			if !motorStopTimer.Stop() {
				<- motorStopTimer.C
				peerTxEnable <- true
			}
			
			elevio.SetDoorOpenLamp(true)
			fmt.Println("The elevators request list at this floor is: ", elev.Requests[newFloor])
			var removingHallButtons [2]bool
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			for btnIndex, btnValue := range removingHallButtons {
				buttonType := elevio.ButtonType(btnIndex)
				if btnValue {
					removeHallReq := messages.HallReqMsg{false, newFloor, buttonType}
					sendingBytes := messages.ToBytes(messages.MsgHallReq, removeHallReq)
					//tcp.TCPSendMessage(masterConn, sendingBytes)
					msgToMasterCh <- sendingBytes 
					fmt.Println("Sending to master that hall req is complete: ", string(sendingBytes))
				}

			}

			if  !doorOpenTimer.Stop() {
				<- doorOpenTimer.C
			}
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)

			// SetAllLights(elev)
			SetAllCabLights(elev)
			elev.State = elevator.DoorOpen
		} else {
			startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)
		}
	default:
		break
	}

}

func OnDoorTimeout(peerTxEnable chan bool) {
	fmt.Println("Door timeout")

	switch elev.State {
	case elevator.DoorOpen:
		if elev.ObstructionActive {
			//timer.Start(elev.Config.DoorOpenDuration)
			if  !doorOpenTimer.Stop() {
				<- doorOpenTimer.C
			}
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration))
			break
		}
		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State
		startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)

		switch elev.State {
		case elevator.DoorOpen:
			//timer.Start(elev.Config.DoorOpenDuration)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration))
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

// Set motor timer and send true to peerTxEnable when the motor starts (ie. motor direction is set to something else than stop)
// Reset motortimer and send true to peerTxEnable every time the elevator reaches a floor/OnFloorArrival
// In CheckForMotorTimeOut, set peerTxEnable to false if the motor has timed out

func motorAndDoorTimerInit() {
	// Initialize motor timer
	motorStopTimer = time.NewTimer(24 * time.Hour)
	doorOpenTimer = time.NewTimer(24 * time.Hour)
	//Sette peer til true??
}

func startMotorStopTimer(motorStopDuration float64, peerTxEnable chan bool) {
	motorStopTimer.Reset(time.Duration(motorStopDuration) * time.Second)
	peerTxEnable <- true
}

/*
func StartDoorOpenTimer(doorOpenDuration float64) {
	doorOpenTimer.Reset(time.Duration(doorOpenDuration) * time.Second)
}*/

