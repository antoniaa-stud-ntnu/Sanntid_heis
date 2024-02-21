package udp_broadcast

import (
	"fmt"
	"log"
	"net"
)

const broadcastAddress = "255.255.255.255"
const broadcastPort = "20006"

func GetLocalIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddress := conn.LocalAddr().(*net.UDPAddr)

    return localAddress.IP
}

func IPToString(ip net.IP) string {
    return ip.String()
}

func BroadcastMessage(message string) {
    broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastAddress+":"+broadcastPort)
    if err != nil {
        fmt.Println("Could not resolve broadcastAddr: ", err)
		return
    }

    conn, err := net.DialUDP("udp", nil, broadcastAddr)
    if err != nil {
		fmt.Println("Could not open broadcast connetion: ", err)
		return
    }
    defer conn.Close()

    _, err = conn.Write([]byte(IPToString(net.IP(message))))
    if err != nil {
		fmt.Println("Could not write to broadcast: ", err)
		return
	}
}

func ListenToBroadcast() {
    broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastAddress+":"+broadcastPort)
    if err != nil {
        fmt.Println("Could not resolve broadcastAddr: ", err)
    }

    conn, err := net.ListenUDP("udp", broadcastAddr)
    if err != nil {
		fmt.Println("Could not listen to broadcast connetion: ", err)
    }
    defer conn.Close()

    buffer := make([]byte, 1024)

    for {
        n, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
			fmt.Println("Could not read from broadcast connetion: ", err)
        }

        message := string(buffer[:n])
        log.Printf("Received broadcast message from %s: %s", addr.String(), message)
    }
}

