package main

import (
	"Project/localElevator/elevio"
	"Project/localElevator/localElevatorHandler"
	"Project/network/messages"
	"Project/network/tcp"
	"Project/network/udpBroadcast"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/roleHandler/roleDistributor"
	"Project/roleHandler/roleFSM"
	"net"
)

func ElevatorProcess() {

}

func main() {

	elevio.Init("localhost:15657", elevio.N_FLOORS)

	buttonsCh := make(chan elevio.ButtonEvent)
	floorsCh := make(chan int)
	obstrCh := make(chan bool)

	peerUpdateToRoleDistributorCh := make(chan peers.PeerUpdate)
	roleAndSortedAliveElevsCh := make(chan roleDistributor.RoleAndSortedAliveElevs, 2)
	isMasterCh := make(chan bool)
	editMastersConnMapCh := make(chan tcp.EditConnMap)

	masterIPCh := make(chan net.IP)
	masterConnCh := make(chan net.Conn)
	//quitOldRecieverCh := make(chan bool)

	sendNetworkMsgCh := make(chan tcp.SendNetworkMsg, 100)
	incommingNetworkMsgCh := make(chan []byte, 100)
	
	toSingleElevFSMCh := make(chan []byte, 5)
	toRoleFSMCh := make(chan []byte, 5)

	peerTxEnable := make(chan bool)

	go elevio.PollRequestButtons(buttonsCh)
	go elevio.PollFloorSensor(floorsCh)
	go elevio.PollObstructionSwitch(obstrCh)

	if elevio.GetFloor() == -1 {
		localElevatorHandler.OnInitBetweenFloors()
	}

	localElevatorHandler.InitLights()

	go udpBroadcast.StartPeerBroadcasting(peerUpdateToRoleDistributorCh, peerTxEnable)
	go roleDistributor.RoleDistributor(peerUpdateToRoleDistributorCh, roleAndSortedAliveElevsCh, masterIPCh)
	go roleFSM.RoleFSM(roleAndSortedAliveElevsCh, toRoleFSMCh, sendNetworkMsgCh, isMasterCh, editMastersConnMapCh)

	go tcp.SetUpMaster(isMasterCh, tcp.MasterPort, editMastersConnMapCh, incommingNetworkMsgCh)
	go tcp.EstablishConnectionAndListen(masterIPCh, tcp.MasterPort, masterConnCh, incommingNetworkMsgCh)
	//go tcp.RecieveMessage(masterConnCh, jsonMessageCh, quitOldRecieverCh)
	go tcp.SendMessage(sendNetworkMsgCh)

	//tcp send msg(to master, masterconn ch)
	//go tcp.TCPRecieveMessage(masterConn, jsonMessageCh, quitCh)
	go messages.DistributeMessages(incommingNetworkMsgCh, toSingleElevFSMCh, toRoleFSMCh)

	//ta in to master ch i FSM
	go localElevatorHandler.LocalElevatorHandler(buttonsCh, floorsCh, obstrCh, masterConnCh, sendNetworkMsgCh, toSingleElevFSMCh, peerTxEnable)

	for {
		select {}
	}
}
