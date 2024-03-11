package roleDistributor

import (
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"bytes"
	"fmt"
	"net"
	"sort"
)

func RoleDistributor(peerUpdateToRoleDistributorCh chan peers.PeerUpdate, MBDCh chan<- string, SortedAliveElevIPsCh chan<- []net.IP) {

	localIPstr, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Could not get local ip: %v\n", err)
	}
	localIP := net.ParseIP(localIPstr) //lokal IP i samme format som sortedIPs
	fmt.Println("RoleDistributor started")

	localElevInPeers := false
	fmt.Println(localIP, localElevInPeers)

	var oldMasterIP net.IP
	var oldBackupIP net.IP
	
	//time.Sleep(1 * time.Second)
	for {
		p := <-peerUpdateToRoleDistributorCh
		fmt.Println("Peer Update to role, peers: ", p.Peers)

		//fmt.Printf("Peer update in role distributor:\n")

		// Extracting IP adresses from peers and finding out if local IP is within
		sortedIPs := make([]net.IP, 0, len(p.Peers))
		for _, ip := range p.Peers {
			peerIP := net.ParseIP(ip)
			sortedIPs = append(sortedIPs, peerIP)
			if peerIP.Equal(localIP) {
				localElevInPeers = true
			}
		}

		// Exiting iteration if peers doesn't include the local elevator
		if !localElevInPeers {
			break
		}

		// Sorting IP addresses to determin master and backup IP
		sort.Slice(sortedIPs, func(i, j int) bool {
			return bytes.Compare(sortedIPs[i], sortedIPs[j]) < 0
		})
		masterIP := sortedIPs[0]
		//fmt.Println("Master Ip := <-peerUpdateToRoleDistributorChP: ", masterIP.String())
		//fmt.Println("2")
		backupIP := net.IP{} //Backup is empty IP if there is only one peer
		if len(sortedIPs) > 1 {
			backupIP = sortedIPs[1]
		}

		//fmt.Println("Before sending")
		for i, ip := range sortedIPs {
			fmt.Printf("Index %d: IP Address: %s\n", i, ip.String())
		}

		//SortedAliveElevIPsCh <- sortedIPs //Sendes masters IP adress on channel, to be used in MBD_FSM

		changeNodeRole := func(nodeID net.IP, role string) {
			fmt.Printf("nodeID-%s == localIP-%s \n", nodeID.String(), localIP.String())
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
		//fmt.Println("Now checking if lost peer")
		if len(p.Lost) > 0 {
			//lostID, _ := strconv.Atoi(p.Lost[0]) //p.lost kan teknisk sett være flere, men i praksis vil to lost samtidig ende opp som en om gangen rett etter hverandre
			lostIP := net.ParseIP(p.Lost[0])
			fmt.Println("I am inside len(p.lost) > 0, lostIP :", lostIP)
			fmt.Printf("Inside len(p.lost) > 0, lostIP-%d, oldMasterIP-%d, oldBackupIP-%d ", lostIP, oldMasterIP, oldBackupIP)
			if lostIP.Equal(oldMasterIP) { //Master lost, backup take over
				fmt.Printf("I am inside len(p.lost) > 0, lost mindre enn master\n")
				changeNodeRole(masterIP, "Master")
				changeNodeRole(backupIP, "Backup")
				setDummies(sortedIPs) //Not tested, Need to ensure that the other elevators are dummys
			} else if lostIP.Equal(oldBackupIP) { //Master intact, but backup lost
				fmt.Printf("I am inside len(p.lost) > 0, lost mindre enn backup\n")
				fmt.Println(len(sortedIPs))
				if len(sortedIPs) > 1 {
					changeNodeRole(backupIP, "Backup")
					
				}
				if len(sortedIPs) > 2 {
					setDummies(sortedIPs) //Not tested, Need to ensure that the other elevators are dummys
				}
			}
		}

		if p.New != "" {

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
		SortedAliveElevIPsCh <- sortedIPs //Sendes masters IP adress on channel, to be used in MBD_FSM
		oldMasterIP = masterIP
		oldBackupIP = backupIP
	}
}

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
