package udp_broadcast

import (
	"fmt"
	"log"
	"net"
    "time"
)

const broadcastAddress = "255.255.255.255"
const primaryPort = "30006"
//const backupPort = "1002"



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

func BroadcastMessageLoop(port string, message string) {
    fmt.Printf("Broadcasting on port %s, with message: %s\n", port, message)
    broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastAddress+":"+port)
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
    for {
        broadcastMessage := []byte(message)
        _, err = conn.Write(broadcastMessage)
        if err != nil {
            fmt.Println("Could not write to broadcast: ", err)
            return
        } else {
            //fmt.Println(message)
        }
    }
    
}

func ListenToBroadcastUntillTimeout(port string, messageCh chan string) {
    broadcastAddr, err := net.ResolveUDPAddr("udp4", broadcastAddress+":"+port)
    if err != nil {
        fmt.Println("Could not resolve broadcastAddr: ", err)
    }

    conn, err := net.ListenUDP("udp4", broadcastAddr)
    if err != nil {
		fmt.Println("Could not listen to broadcast connetion: ", err)
    }
    defer conn.Close()

    buffer := make([]byte, 1024)
    
    for {
        
        conn.SetReadDeadline(time.Now().Add(5 * time.Second))
        n, addr, err := conn.ReadFromUDP(buffer)
        if err != nil {
			fmt.Println("Could not read from broadcast connetion: ", err)
            break
        }

        message := string(buffer[:n])
        log.Printf("Received broadcast message from %s: %s", addr.String(), message)
        messageCh <- addr.String()
    }
}

func ProcessPairInit(){
    primaryIPCh := make(chan string)
    ListenToBroadcastUntillTimeout(primaryPort, primaryIPCh)
    //localIP_string := IPToString(GetLocalIP())
    BroadcastMessageLoop(primaryPort, "I'm primary and i'm alive")
}

