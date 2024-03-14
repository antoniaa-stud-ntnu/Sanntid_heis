package localElevatorHandler

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast/udpNetwork/localip"
	"fmt"
	"net"
	"time"
)

// "Project/network/udpBroadcast/udpNetwork/localip"
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

//var doorOpenTimerCh <- chan time.Time


func LocalElevatorHandler(buttonsCh chan elevio.ButtonEvent, floorsCh chan int, obstrCh chan bool, masterConnCh chan net.Conn, toNetworkCh chan tcp.SendNetworkMsg, toFSMCh chan []byte, peerTxEnable chan bool) {
	masterConn := <-masterConnCh

	localip, _ := localip.LocalIP()
	msgElevState.IpAddr = localip

	motorAndDoorTimerInit()

	for {
		select {
		case button := <-buttonsCh:
			if button.Button == elevio.Cab {
				msgElevState.ElevState.CabRequests[button.Floor] = true
				//fmt.Println("To master elevState: ", msgElevState.ElevState)
				sendingBytes := messages.PackMessage(messages.MsgElevState, msgElevState)
				toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}

				OnRequestButtonPress(button.Floor, button.Button, peerTxEnable, toNetworkCh, masterConn)
				elevio.SetButtonLamp(button.Floor, button.Button, true)
			} else {
				hallReq := messages.HallReqMsg{true, button.Floor, button.Button}
				sendingBytes := messages.PackMessage(messages.MsgHallReq, hallReq)
				toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
				// fmt.Println("Sent hall request to master: ", button.Floor, button.Button)
			}
		case floor := <-floorsCh:
			removingHallButtons := OnFloorArrival(floor, peerTxEnable)
			
			msgElevState.ElevState.Floor = floor
			msgElevState.ElevState.CabRequests[floor] = false
			sendingBytes := messages.PackMessage(messages.MsgElevState, msgElevState)
			toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
			
			sendHallRemovalToConn(removingHallButtons, floor, toNetworkCh, masterConn)
			
		case obstr := <-obstrCh:
			fmt.Printf("Obstruction is %+v\n", obstr)
			OnObstruction(obstr)
			//elevator.ElevatorPrint(elev)
		case msgFromNetwork := <-toFSMCh:
			msgType, data := messages.UnpackMessage(msgFromNetwork)
			switch msgType {
			case messages.MsgAssignedHallReq:
				newHallRequests := data.([][2]bool) // Dette er ikke nodvendig egentlig
				for floor := 0; floor < len(newHallRequests); floor++ {
					for hallIndex := 0; hallIndex < len(newHallRequests[hallIndex]); hallIndex++ {
						value := newHallRequests[floor][hallIndex]
						if value {
							OnRequestButtonPress(floor, elevio.ButtonType(hallIndex), peerTxEnable, toNetworkCh, masterConn)
							// fmt.Println("Checking if the request is true in elev.Requests: ", elev.Requests[floor][hallIndex])
						} else {
							elev.Requests[floor][hallIndex] = false
						}
					}
				}
				//elevator.ElevatorPrint(elev)
				fmt.Println("Requevec hall request, all requests now: ", elev.Requests)
			case messages.MsgHallLigths:
				elevio.SetButtonLamp(data.(messages.HallReqMsg).Floor, data.(messages.HallReqMsg).Button, data.(messages.HallReqMsg).TAddFRemove)

			case messages.MsgRestoreCabReq:
				for floor := 0; floor < len(data.([]bool)); floor++ {
					elev.Requests[floor][elevio.Cab] = data.([]bool)[floor]
				}
			}
		case newMasterConn := <-masterConnCh:
			masterConn = newMasterConn
			fmt.Println("New masterConn is: ", masterConn)

		//case <-doorOpenTimerCh:
		case <-doorOpenTimer.C:
			fmt.Println("Door timed out")
			OnDoorTimeout(masterConn, toNetworkCh)
			//doorOpenTimer.Reset(24 * time.Hour)

		case <-motorStopTimer.C:
			fmt.Println("Motor timed<-motorStopTimer.C: out")
			peerTxEnable <- false
			fmt.Println("Not visible in peers")
			motorStopTimer.Reset(24 * time.Hour)
		}
	}
}

func SetLight(btn elevio.ButtonType, floor int, onOff bool) {
	elevio.SetButtonLamp(floor, elevio.ButtonType(btn), onOff)
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

func OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.Down)
	elev.Dirn = elevio.Down
	elev.State = elevator.Moving
}

func OnRequestButtonPress(btnFloor int, btnType elevio.ButtonType, peerTxEnable chan bool, toNetworkCh chan tcp.SendNetworkMsg, masterConn net.Conn) {
	//fmt.Printf("\n\n%s(%d, %s)\n", "fsmOnRequestButtonPress", btnFloor, elevator.ButtonToString(btnType))

	switch elev.State {
	case elevator.DoorOpen:
		if requests.ShouldClearImmediately(elev, btnFloor, btnType) {
			//timer.Start(elev.Config.DoorOpenDuration)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			// motorstoptimer ogaå?
			//doorOpenTimerCh = time.After(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
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
			//doorOpenTimerCh = time.After(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			// motorstoptimer ogaå?
			var removingHallButtons [2]bool
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			sendHallRemovalToConn(removingHallButtons, btnFloor, toNetworkCh, masterConn)

		case elevator.Moving:
			elevio.SetMotorDirection(elev.Dirn)
			startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)

		case elevator.Idle:
			//doorOpenTimer.Stop()
			//doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			//motorStopTimer.Reset(24 * time.Hour)
		}
	}

	// SetAllLights(elev)
}

func OnFloorArrival(newFloor int, peerTxEnable chan bool) [2]bool {
	var removingHallButtons [2]bool
	fmt.Printf("\n\n%s(%d)\n", "Arrival on floor ", newFloor)

	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	// Reset motor timer
	fmt.Println("Reset motor timer")
	startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)

	switch elev.State {
	case elevator.Moving:
		if requests.ShouldStop(elev) { //I en etasje med request eller ingen requests over/under
			elevio.SetMotorDirection(elevio.Stop)
			// if !motorStopTimer.Stop() {
			// 	<-motorStopTimer.C
			// 	peerTxEnable <- true
			// }
			elevio.SetDoorOpenLamp(true)
			//fmt.Println("The elevators request list at this floor is: ", elev.Requests[newFloor])
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			//fmt.Println("The elevators request cab request at this floor is: ", elev.Requests[newFloor][2])

			//if !doorOpenTimer.Stop() {
			//	<-doorOpenTimer.C
			//}
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			//doorOpenTimerCh = time.After(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			SetAllCabLights(elev)
			elev.State = elevator.DoorOpen
			//motorStopTimer.Reset(24 * time.Hour)
		} else {
			//startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)
		}
	default:
		break
	}
	return removingHallButtons
}

func OnDoorTimeout(masterConn net.Conn, toNetworkCh chan tcp.SendNetworkMsg) {
	//fmt.Println("Door timeout------")
	switch elev.State {
	case elevator.DoorOpen:
		if elev.ObstructionActive {
			//timer.Start(elev.Config.DoorOpenDuration)
			//if !doorOpenTimer.Stop() {
			//	<-doorOpenTimer.C
			//}
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration)* time.Second)
			//doorOpenTimerCh = time.After(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			break
		}
		fmt.Println("Elev state before choosing new dir: ", elev.State)
		fmt.Println("Elev dirn before choosing new dir: ", elev.Dirn)

		var pair requests.DirnBehaviourPair = requests.ChooseDirection(elev)
		elev.Dirn = pair.Dirn
		elev.State = pair.State
		fmt.Println("Elev state after choosing new dir: ", elev.State)
		fmt.Println("Elev dirn after choosing new dir: ", elev.Dirn)

		switch elev.State {
		case elevator.DoorOpen:
			fmt.Println("Door open after choosing new dir")
			//timer.Start(elev.Config.DoorOpenDuration)
			doorOpenTimer.Reset(time.Duration(elev.Config.DoorOpenDuration)* time.Second)
			//doorOpenTimerCh = time.After(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
			var removingHallButtons [2]bool
			elev, removingHallButtons = requests.ClearAtCurrentFloor(elev)
			sendHallRemovalToConn(removingHallButtons, elev.Floor, toNetworkCh, masterConn)
			// SetAllLights(elev)
			SetAllCabLights(elev)

			
		case elevator.Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
			//startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)
			motorStopTimer.Reset(time.Duration(elev.Config.MotorStopDuration)* time.Second)
			
		case elevator.Idle:
			elevio.SetDoorOpenLamp(false)
			fmt.Println("Setting motor dir to: ", elev.Dirn)
			elevio.SetMotorDirection(elev.Dirn)
			fmt.Println("Inni idle, restting timer")
			//startMotorStopTimer(elev.Config.MotorStopDuration, peerTxEnable)
			motorStopTimer.Reset(24 * time.Hour)
			
			
		}
		
	default:
		
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

	//motorStopTimer.Stop()
	//doorOpenTimer.Stop()
	//doorOpenTimer = time.NewTimer(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
	//doorOpenTimerCh = time.After(time.Duration(elev.Config.DoorOpenDuration) * time.Second)
	//Sette peer til true??
}

func startMotorStopTimer(motorStopDuration float64, peerTxEnable chan bool) {
	fmt.Println("Start motor stop timer")
	motorStopTimer.Reset(time.Duration(motorStopDuration) * time.Second)
	peerTxEnable <- true
}

func sendHallRemovalToConn(removingHallButtons [2]bool, floor int, toNetworkCh chan tcp.SendNetworkMsg, masterConn net.Conn) {
	for btnIndex, btnValue := range removingHallButtons {
		buttonType := elevio.ButtonType(btnIndex)
		if btnValue {
			removeHallReq := messages.HallReqMsg{false, floor, buttonType}
			sendingBytes := messages.PackMessage(messages.MsgHallReq, removeHallReq)
			toNetworkCh <- tcp.SendNetworkMsg{masterConn, sendingBytes}
			//fmt.Println("Sending to master that hall req is complete: ", string(sendingBytes))
		}
	}
}