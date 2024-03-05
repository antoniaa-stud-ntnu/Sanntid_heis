package dummy

import (
	"Project/network/tcp"
	"Project/network/udpBroadcast"
	"Project/singleElevator/elevator"
	"fmt"
	"os"
)

const MasterPort = 27300
const BackupPort = 27301
const DummyPort = 27302

// Dummy init
	// Elevator init
	// Listen to primary broadcast (UDP)
	// If primary is dead, become primary
	// If primary is alive, save primary IP
	// Open TCP connection to primary, by choosing random port number, encapsulating its own IP and sending it to primary
	// Send full state to primary
	// go sendAliveMessage()


// Send Im alive message to primary (TCP)




func OnDummy() {
	singleElevatorProcess()
	//Send full state to master
	
	fmt.Println("Dummy elevator is running")
}


func MBD_FSM (MBDCh chan string, PrimaryIPCh chan string) {
	PrimaryIP := <- PrimaryIPCh
	println("Primary IP is:", PrimaryIP)
	MBD := <- MBDCh
	for {
		switch MBD {
		case "Master":
			//Do master stuff
			tcp.TCP_server("localhost", MasterPort)
		case "Backup":
			//Do backup stuff
			tcp.TCP_client
		case "Dummy":

			//Do dummy stuff
			//Send state update to master
			//Listen to master

		}
	}
}