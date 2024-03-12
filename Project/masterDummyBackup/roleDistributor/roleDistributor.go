package roleDistributor

import (
	"Project/network/udpBroadcast/udpNetwork/localip"
	"Project/network/udpBroadcast/udpNetwork/peers"
	"bytes"
	"fmt"
	"net"
	"sort"
	//"time"
)
type RoleAndSortedAliveElevs struct {
	Role 				string
	SortedAliveElevs 	[]net.IP
}


type Role int

const (
    Master Role = iota // 0
    Backup             // 1
    Dummy              // 2
)

func (r Role) String() string {
    switch r {
    case Master:
        return "Master"
    case Backup:
        return "Backup"
    default:
        return "Dummy"
    }
}



func RoleDistributor(peerUpdateToRoleDistributorCh chan peers.PeerUpdate, roleAndSortedAliveElevs chan<- RoleAndSortedAliveElevs) {

	localIPstr, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Could not get local ip: %v\n", err)
	}
	localIP := net.ParseIP(localIPstr) //lokal IP i samme format som sortedIPs
	fmt.Println("RoleDistributor started")

	localElevInPeers := false
	fmt.Println(localIP, localElevInPeers)

	
	//time.Sleep(1 * time.Second)
	for {
		p := <-peerUpdateToRoleDistributorCh
		fmt.Println("Peer Update to role, peers: ", p.Peers)

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

		// Sorting IP addresses to find out which role the local elevator should have
		sort.Slice(sortedIPs, func(i, j int) bool {
			return bytes.Compare(sortedIPs[i], sortedIPs[j]) < 0
		})


		 checkRoles := func(sortedIPs []net.IP) string {
			for i, ip := range sortedIPs {
				var expectedRole Role
				switch i {
				case 0:
					expectedRole = Master
				case 1:
					expectedRole = Backup
				default:
					expectedRole = Dummy
				}
				
				if ip.Equal(localIP) {
					return expectedRole.String()
				}
				
			}
			return ""
		}
			
		
		
		newRole := ""
		
		if len(p.Lost) > 0 {
			newRole = checkRoles(sortedIPs)
		}

		if p.New != "" {
			newRole = checkRoles(sortedIPs)
		}
		//time.Sleep(3*time.Second)
		
		roleAndSortedAliveElevs <- RoleAndSortedAliveElevs{newRole, sortedIPs} 
		fmt.Println("Sent updated role and sorted alive elevs to MBD_FSM")
		
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
