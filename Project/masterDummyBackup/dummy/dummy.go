package dummy

import (
	"Project/masterDummyBackup/master"
	"Project/network/tcp"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

var masterConn net.Conn

type HRAElevState struct { // StateValues
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct { // statusStruct
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

type msgElevState struct {
	IpAddr    string       `json:"ipAddr"`
	ElevState HRAElevState `json:"elevState"`
}

type msgHallReq struct {
	TAdd_fRemove bool              `json:"tAdd_fRemove"`
	Floor        int               `json:"floor"`
	Button       elevio.ButtonType `json:"button"`
} // Up-0, Down-1, lik med index i HallRequests [][2]

// -----------------------------------------------------------------------

func OnDummy(MBDCh chan string, statusUpdateCh chan HRAElevState, SortedAliveElevIPsCh chan []net.IP, jsonMessageCh chan []byte, peerUpdateToPrimaryHandlerCh chan peers.PeerUpdate) {

	// Initialize an elevator
	var elev elevator.Elevator = elevator.InitElev() // On Elevator-format
	var hraElevState HRAElevState
	dummyIP, _ := localip.LocalIP() // Get elevator IP
	hallreqMatrix = elev.Requests

	hraElevState = initDummy(elev, hraElevState) // On HRAElevState-format
	statusUpdateCh <- hraElevState

	// Listen to primary broadcast (UDP)
	udpBroadcast.StartPeerBroadcasting(peerUpdateToPrimaryHandlerCh)

	// Establish a TCP connection with master
	sortedAliveElevs := <-SortedAliveElevIPsCh
	masterConn, _ = tcp.TCPMakeMasterConnection(sortedAliveElevs[0].String(), master.MasterPort)

	// Send status updates periodically to master via msgElevState
	sendStatusUpdate(dummyIP, hallreqMatrix, masterConn, statusUpdateCh)

	// Receive commands from master via msgHallReq
	receiveHallReq(masterConn, hallreqMatrix, jsonMessageCh, statusUpdateCh)

	// Basic operation of an elevator. Is this done in fsm?

	// Handle cabrequests

	// Send I'm alice message to primary via TCP.
	go sendAliveMessage(masterConn)

}

// -----------------------------------------------------------------------

func SendElevState(msgEl msgElevState, conn net.Conn) {
	jsonElevState, err := json.Marshal(msgEl)
	if err != nil {
		fmt.Println("Error encoding json: ", err)
	}
	tcp.TCPSendMessage(conn, jsonElevState)
}

func SendHallRequests(msgReq msgHallReq, conn net.Conn) {
	jsonHallReq, err := json.Marshal(msgReq)
	if err != nil {
		fmt.Println("Error encoding json: ", err)
	}
	tcp.TCPSendMessage(conn, jsonHallReq)
}

func initDummy(elev elevator.Elevator, elevState HRAElevState) HRAElevState {
	elevState.Behaviour = elevator.EbToString(elev.State)
	elevState.Floor = elev.Floor
	elevState.Direction = elevator.DirnToString(elev.Dirn)
	elevState.CabRequests = make([]bool, elevio.N_FLOORS)  // Use Request from elevator.Elevator?

	return elevState
}

func sendStatusUpdate(ip string, hallreqMatrix [][2]bool, conn net.Conn, statusUpdateCh chan HRAElevState) {
	var msgStatusUpdate msgElevState   // Instance of msgElevState
	var msgHallReqCompleted msgHallReq // Instance of msgHallReq

	for {
		select {
		case statusUpdate := <-statusUpdateCh: // Put the states from channel on a variable statusUpdate
			msgStatusUpdate.IpAddr = ip              // Set IP
			msgStatusUpdate.ElevState = statusUpdate // Set elevator state
			SendElevState(msgStatusUpdate, conn)
			if hallreqMatrix[statusUpdate.Floor][0] || hallreqMatrix[statusUpdate.Floor][1] {
				msgHallReqCompleted.TAdd_fRemove = false
				msgHallReqCompleted.Floor = statusUpdate.Floor
				if statusUpdate.Direction == "up" {
					msgHallReqCompleted.Button = 0
				} else {
					msgHallReqCompleted.Button = 1
				}
				SendHallRequests(msgHallReqCompleted, conn)
			}
		}
	}
}

func receiveHallReq(conn net.Conn, hallreqMatrix [][2]bool, jsonMessageCh chan []byte, statusUpdateCh chan HRAElevState) {
	var msgHallReq msgHallReq

	tcp.TCPReciveMessage(conn, jsonMessageCh)

	for {
		select {
		case jsonMsg := <-jsonMessageCh:
			err := json.Unmarshal(jsonMsg, &msgHallReq) // Unmarshal the data sent from Master
			if err != nil {
				fmt.Println("Error decoding json:", err)
			}
			hallreqMatrix[msgHallReq.Floor][msgHallReq.Button] = msgHallReq.TAdd_fRemove // Add the correct hall request to the hallReqMatrix
		}
	}
}

func sendAliveMessage(conn net.Conn) {
	_, err := conn.Write([]byte("I'm alive"))
	if err != nil {
		fmt.Printf("Failed to send message: %s\n", err)
	}
	time.Sleep(100 * time.Millisecond)
}

// -----------------------------------------------------------------------

// Dummy init
// Elevator init
// Listen to primary broadcast (UDP)
// If primary is dead, become primary
// If primary is alive, save primary IP
// Open TCP connection to primary, by choosing random port number, encapsulating its own IP and sending it to primary
// Send full state to primary
// go sendAliveMessage()

// Send Im alive message to primary (TCP)
