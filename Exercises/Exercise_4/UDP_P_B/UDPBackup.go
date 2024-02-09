package UDPBackup
 
import (
	"net"
	"fmt"
	"strconv"
	"os"
)


const (
	host = "localhost"  // elevio.Init("localhost:15657", numFloors)
	port = 9000
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

// UDP Listen
func UDPListen(listenCh chan byte, quit chan bool, flag int) {
	buffer := make([]byte, 1)
	flag = NotEqual(flag)
	laddr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(port+flag))
	CheckError(err)

	listen, err := net.ListenUDP("udp4", laddr)
	CheckError(err)
	defer listen.Close()
	for {
		select {
		case <-quit:
			return
		default:
			_, _, err := listen.ReadFromUDP(buffer)
			CheckError(err)
			listenCh <- buffer[0]
		}

	}
}
// UDP send
func UDPSend(counter byte, flag int){
	baddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(host, strconv.Itoa(port+flag)))
	CheckError(err)

	conn, err := net.DialUDP("udp4", nil, baddr)
	CheckError(err)
	defer conn.Close()

	buf := make([]byte, 1)
	buf[0] = counter

	_, err = conn.Write(buf)
	CheckError(err)
}

func NotEqual(num int) int {
	if num == 0 {
		return 1
	} else if num == 1 {
		return 2
	} else {
		return 0
	}
}