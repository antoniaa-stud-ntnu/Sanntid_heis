package roleDistributor

import (
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevator"
	"fmt"
	"net"
	"sort"
	"bytes"
)

func RoleDistributor(peerUpdateToRoleDistributorCh chan peers.PeerUpdate, MBDCh chan elevator.MasterBackupDummyType, PrimaryIPCh chan net.IP) {
	
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
		case p := <- peerUpdateToRoleDistributorCh:
			fmt.Printf("I am inside case p1, RoleDistributor\n")
			sortedIPs := make([]net.IP, 0, len(p.Peers)) //
			for _, ip := range p.Peers {
				sortedIPs = append(sortedIPs, net.ParseIP(ip))
			}
			sort.Slice(sortedIPs, func(i, j int) bool {
				return bytes.Compare(sortedIPs[i], sortedIPs[j]) < 0
			})
			fmt.Printf("I am inside case p2, RoleDistributor\n")
			localIPstr, err := localip.LocalIP()
			if err != nil {
							fmt.Printf("Could not get local ip: %v\n", err)
						}
			localIP := net.ParseIP(localIPstr) 
			fmt.Printf("I am inside case p3, RoleDistributor\n")
			//peerIDs := peerIDfromString(p.Peers)
			//sort.Ints(peerIDs) //The lowest ID is always the primary and backup is next
			masterIP := sortedIPs[0]
			backupIP := net.IP{} //Backup is empty IP if there is only one peer
			if len(sortedIPs) > 1 {
				backupIP = sortedIPs[1]
				fmt.Printf("I am inside len()>1 RoleDistributor, lost\n")
			} 
			fmt.Println("Before sending PrimaryIPCh")
			PrimaryIPCh <- masterIP //Sendes "IP-ProsessID" on channel, to be used in MBD_FSM
			fmt.Println("After sending PrimaryIPCh")
			changeNodeRole := func(nodeID net.IP, role elevator.MasterBackupDummyType) {
				if nodeID.Equal(localIP) {
					MBDCh <- role
				}
			}
			fmt.Printf("I am inside case p4, RoleDistributor\n")
			if len(p.Lost) > 0 {
				fmt.Printf("I am inside RoleDistributor, lost\n")
				//lostID, _ := strconv.Atoi(p.Lost[0]) //p.lost kan teknisk sett være flere, men i praksis vil to lost samtidig ende opp som en om gangen rett etter hverandre
				lostIP := net.ParseIP(p.Lost[0])
				//if lostIP < masterIP { //Master lost, backup take over
				if bytes.Compare(lostIP, masterIP) == -1 { //Master lost, backup take over
					changeNodeRole(masterIP, 0)
					changeNodeRole(backupIP, 1)
				} else if bytes.Compare(lostIP, backupIP) == -1 { //Master intact, but backup lost
					changeNodeRole(backupIP, 1)
				}
			} 
			if p.New != "" {
				fmt.Printf("I am inside RoleDistributor, new\n")
				newID := net.ParseIP(p.New)
				if newID.Equal(masterIP) { //New master
					changeNodeRole(masterIP, 0)
					changeNodeRole(backupIP, 1)
				} else if newID.Equal(backupIP) { //New backup
					changeNodeRole(backupIP, 1)
				}

				if !newID.Equal(masterIP) && !newID.Equal(backupIP) { //New dummy
					changeNodeRole(newID, 2)
				}
			}
			
		}
	}
}

/*
func peerIDfromString(Peers []string)  ([]int) {
	var t2 = []int{}

    for _, i := range Peers {
        j, err := strconv.Atoi(i)
        if err != nil {
            panic(err)
        }
        t2 = append(t2, j)
    }
	return t2
}*/

