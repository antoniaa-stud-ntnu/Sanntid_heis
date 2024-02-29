package dummy

import (
	"Project/network/tcp"
	"Project/network/udpBroadcast"
	"Project/singleElevator/elevator"
	"fmt"
	"os"
)

const masterPort = 27300
const backupPort = 27301
const dummyPort = 27302

// Dummy init
	// Elevator init
	// Listen to primary broadcast (UDP)
	// If primary is dead, become primary
	// If primary is alive, save primary IP
	// Open TCP connection to primary, by choosing random port number, encapsulating its own IP and sending it to primary
	// Send full state to primary
	// go sendAliveMessage()


// Send Im alive message to primary (TCP)


// Do as primary says




func MBD_FSM (MBDCh chan int, PrimaryIPCh chan string) {
	PrimaryIP := <- PrimaryIPCh
	MBD := <- MBDCh
	for {
		switch MBD {
		case elevator.MasterBackupDummy.Master:
			//Do master stuff
			tcp.TCP_server("localhost", masterPort)
		case elevator.MasterBackupDummy.Backup:
			//Do backup stuff
		case elevator.MasterBackupDummy.Dummy:

			//Do dummy stuff
			//Send state update to master

			//Listen to master

		}
	}
}