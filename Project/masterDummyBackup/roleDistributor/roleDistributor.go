package roleDistributor

import (
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"Project/singleElevator/elevator"
	"bytes"
	"fmt"
	"net"
	"sort"
)

func RoleDistributor(peerUpdateToRoleDistributorCh chan peers.PeerUpdate, MBDCh chan string, PrimaryIPCh chan net.IP) {

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
		case p := <-peerUpdateToRoleDistributorCh:
			sortedIPs := make([]net.IP, 0, len(p.Peers))
			for _, ip := range p.Peers {
				sortedIPs = append(sortedIPs, net.ParseIP(ip))
			} //Lager liste med alle heisene i IP-format, sortedIPs
			sort.Slice(sortedIPs, func(i, j int) bool {
				return bytes.Compare(sortedIPs[i], sortedIPs[j]) < 0
			}) //Sorterer IP-ene (sortedIPs) i stigende rekkefølge

			localIPstr, err := localip.LocalIP()
			if err != nil {
				fmt.Printf("Could not get local ip: %v\n", err)
			}
			localIP := net.ParseIP(localIPstr) //lokal IP i samme format som sortedIPs

			masterIP := sortedIPs[0]
			backupIP := net.IP{} //Backup is empty IP if there is only one peer
			if len(sortedIPs) > 1 {
				backupIP = sortedIPs[1]
			}

			PrimaryIPCh <- masterIP //Sendes masters IP adress on channel, to be used in MBD_FSM

			changeNodeRole := func(nodeID net.IP, role string) {
				if nodeID.Equal(localIP) {
					fmt.Printf("I am now changing role to %v\n", role)
					MBDCh <- role
				}
			}

			setDummies := func(sortedIPs []net.IP) {
				for dummy := 2; dummy < len(sortedIPs); dummy++ {
					changeNodeRole(sortedIPs[dummy], "Dummy")
				}
			}

			if len(p.Lost) > 0 {
				//lostID, _ := strconv.Atoi(p.Lost[0]) //p.lost kan teknisk sett være flere, men i praksis vil to lost samtidig ende opp som en om gangen rett etter hverandre
				lostIP := net.ParseIP(p.Lost[0])
				fmt.Printf("I am inside len(p.lost) > 0 \n")
				if bytes.Compare(lostIP, masterIP) == -1 { //Master lost, backup take over
					fmt.Printf("I am inside len(p.lost) > 0, lost mindre enn master\n")
					changeNodeRole(masterIP, "Master")
					changeNodeRole(backupIP, "Backup")
					setDummies(sortedIPs) //Not tested, Need to ensure that the other elevators are dummys
				} else if bytes.Compare(lostIP, backupIP) == -1 { //Master intact, but backup lost
					fmt.Printf("I am inside len(p.lost) > 0, lost mindre enn backup\n")
					changeNodeRole(backupIP, "Backup")
					setDummies(sortedIPs) //Not tested, Need to ensure that the other elevators are dummys
				}
			}
			if p.New != "" {
				fmt.Println("I am inside new peer handler")
				newID := net.ParseIP(p.New)
				if newID.Equal(masterIP) { //New master
					changeNodeRole(masterIP, "Master")
					changeNodeRole(backupIP, "Backup")
				} else if newID.Equal(backupIP) { //New backup
					changeNodeRole(backupIP, "Backup")
				}

				if !newID.Equal(masterIP) && !newID.Equal(backupIP) { //New dummy
					//changeNodeRole(newID, 2)
					setDummies(sortedIPs)
				}
			}

		}
	}
}
