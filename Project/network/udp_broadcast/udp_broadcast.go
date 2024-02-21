package udp_broadcast

import (
	"log"
	"net"
)

const broadcastAddress = "255.255.255.255"
const broadcasrPort = "20006"

func GetLocalIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddress := conn.LocalAddr().(*net.UDPAddr)

    return localAddress.IP
}

