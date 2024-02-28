package primary

import (
	"Project/network/Network-go/network/peers"
	"Project/singleElevator/elevator"
	"command-line-arguments/home/student/Heis2/Sanntid_heis/Project/network/Network-go-master/network/peers/peers.go"
	"strconv"
)

// Connection with elevators
// Needs to update state variable when new information is available - retry if failed
// Sends the information to backup
// Recieves either ACK or NACK from backup that the information is stored
// If NACK, it needs to resend the information
// Run hall_request_assigner
// Send hall_request_assigner output to elevators

// Alive elevators variable, contains ip adresses and port to all alive elevators

// Update all states variable

// Run hall_request_assigner

// Primary init
// Initialize all states
// go sendIPtoPrimaryBroadcast() //Im alive message
// Choose a backup randomly from the alive elevators --> spawn backup.
//
func PrimaryHandler(peerUpdateCh chan peers.PeerUpdate, MBDCh chan elevator.MasterBackupDummy) {
	
	// Hver gang det skjer en endring av antall heiser kalles denne

	// If one connection lost:
		// If master is lost
			// Backup take over
		// else (backup lost)
			// make a new backup
		// else (dummy elevator)
			// don't care
	// If new connection:
		// If the new elevator is master (and there are two masters now):
			// deside which elevator is master
		// else
			// master is starting a TCP conntection to the dummy elevator
	for {
		select {
		case p := <- peerUpdateCh:
			peerIDs = peerIDfromString(p.Peers)
			sort.Ints(peerIDs) //The lowest ID is always the primary and backup is next
			masterID = peerIDs[0]
			backupID = peerIDs[1]
			if len(p.Lost) == 0 {
				lostID := strconv.Atoi(p.Lost)
				if lostID < masterID { //Master lost, backup take over
					if getThisID() == masterID {
						MBDCh <- elevator.MasterBackupDummy.Master
					} else if getThisID() == backupID {
						MBDCh <- elevator.MasterBackupDummy.Backup
					} 
				} else if lostID < backupID { //Master intact, but backup lost
					if getThisID() == backupID {
						MBDCh <- elevator.MasterBackupDummy.Backup
					} 
				}
			}
			
		}
	}
}

func peerIDfromString (Peers []string) PeerIDs []int {
	var t2 = []int{}

    for _, i := range t {
        j, err := strconv.Atoi(i)
        if err != nil {
            panic(err)
        }
        t2 = append(t2, j)
    }
	return t2
}

func getThisID () thisID int {
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	id := fmt.Sprintf("%s-%d", localIP, os.Getpid())
	return strconv.Atoi(id) //Returning id as int
}