package tcp

import (
	
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

const MasterPort = "20019"

type EditConnMap struct {
	Insert 		bool
	ClientIP 	string
	Conn       	net.Conn
}



func SetUpMaster(isMasterCh chan bool, masterPort string, editMastersConnMapCh chan EditConnMap, incommingNetworkMsgCh chan []byte) {
	quitRecieverCh := make(chan bool)
	iPToConnMap := make(map[string]net.Conn)
	var wasMaster bool
	
	for {
		isMaster := <- isMasterCh
		if isMaster {
			wasMaster = true
			//go ListenForConnectionsAndHandle(masterPort, &iPToConnMap, editMastersConnMapCh, incommingNetworkMsgCh, quitRecieverCh)
			go func() {
				tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+masterPort)
				if err != nil {
					fmt.Printf("Could not resolve address: %s\n", err)
					os.Exit(1)
				}
				listener, err := net.ListenTCP("tcp", tcpAddr)
				if err != nil {
					fmt.Println("Could not open listener: ", err)
					return
				}
				defer listener.Close()
			
				fmt.Printf("Server is listening on port %s\n", masterPort)
			
				for {
					conn, err := listener.Accept()
					if err != nil {
						fmt.Printf("Could not accept connection: %s\n", err)
						continue
					}
					//fmt.Println("Master accepted new connection: ", conn)
			
					connectionsIP := ((conn.RemoteAddr().(*net.TCPAddr)).IP).String()
					//fmt.Printf("Master accepted new connection-%d from IP-%s \n", conn, connectionsIP)
					(iPToConnMap)[connectionsIP] = conn
					//fmt.Printf("iPToConnMap is updated to: IP-%s, Conn-%d", connectionsIP, conn)
					editMastersConnMapCh <- EditConnMap{true, connectionsIP, conn}
			
					// Handle client connection in a goroutine
					go recieveMessage(conn, incommingNetworkMsgCh, quitRecieverCh)
				}
			}()
			go func() {
				for ip, conn := range iPToConnMap {
					_, err := conn.Read(make([]byte, 1024))
					if err != nil {
						fmt.Println("Deleting a conn: ", (iPToConnMap))
						delete((iPToConnMap), ip)
					}
				}
			}()
		} else {
			if wasMaster {
				quitRecieverCh <- false
				wasMaster = false
			}
			
		}
	}
}

/*
// Master opening a listening server and saving+handling incomming connections from all the elevators
func ListenForConnectionsAndHandle(masterPort string, iPToConnMap *map[string]net.Conn, editConnMapCopyCh chan EditConnMap, incommingNetworkMsgCh chan []byte, quitCh chan bool) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+masterPort)
	if err != nil {
		fmt.Printf("Could not resolve address: %s\n", err)
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Could not open listener: ", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server is listening on port %s\n", masterPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Could not accept connection: %s\n", err)
			continue
		}
		//fmt.Println("Master accepted new connection: ", conn)

		connectionsIP := ((conn.RemoteAddr().(*net.TCPAddr)).IP).String()
		//fmt.Printf("Master accepted new connection-%d from IP-%s \n", conn, connectionsIP)
		(*iPToConnMap)[connectionsIP] = conn
		//fmt.Printf("iPToConnMap is updated to: IP-%s, Conn-%d", connectionsIP, conn)
		editConnMapCopyCh <- EditConnMap{true, connectionsIP, conn}

		// Handle client connection in a goroutine
		go recieveMessage(conn, incommingNetworkMsgCh, quitCh)
	}
}

func LookForClosedConns(iPToConnMap *map[string]net.Conn) {
	for ip, conn := range *iPToConnMap {

		_, err := conn.Read(make([]byte, 1024))
		if err != nil {
			
			fmt.Println("Deleting a conn: ", (*iPToConnMap))
			delete((*iPToConnMap), ip)
		}
	}
}
*/

// Recieving messages and sending them on a channel for to be handeled else where

func EstablishConnectionAndListen(ipCh chan net.IP, port string, connCh chan net.Conn, incommingNetworkMsgCh chan []byte) {
	quitOldReciever := make(chan bool)
	var conn net.Conn
	for {
		IP := <-ipCh
		tcpAddr, err := net.ResolveTCPAddr("tcp4", IP.String()+":"+port)
		if err != nil {
			fmt.Printf("Could not resolve address: %s\n", err)
		}

		if conn != nil {
			conn.Close()
			quitOldReciever <- true
		}


		for {
			conn, err = net.Dial("tcp", tcpAddr.String())
			if err != nil {
				fmt.Println("Could not connect to server: ", err)
				time.Sleep(50*time.Millisecond)
			}else {
				break
			}
		}
		fmt.Println("Connected to master")
		connCh <- conn

		go recieveMessage(conn, incommingNetworkMsgCh, quitOldReciever)
	}
}

func recieveMessage(conn net.Conn, incommingMsgCh chan<- []byte, quitCh <-chan bool) {
	defer conn.Close()
	fmt.Println("In TCP recieve message")

	buffer := make([]byte, 65536)

	for {
		select {
		case <- quitCh:
			return
		default:
			data, err := conn.Read(buffer)
			if err != nil {
				conn.Close()
				if err == io.EOF {
					fmt.Println("Client closed the connection.")
				} else {
					fmt.Println("Error:", err)
				}
				return
			}

		msg := make([]byte, data)
		copy(msg, buffer[:data])
		incommingMsgCh <- msg
		}
	}
}

type SendNetworkMsg struct {
	RecieverConn net.Conn
	Message 	[]byte
}

func SendMessage(sendNetworkMsgCh chan SendNetworkMsg) { // (to master, masterconn ch)
	for {
		fmt.Println("Sending msg")
		sendNetworkMsg := <-sendNetworkMsgCh
		conn := sendNetworkMsg.RecieverConn
		message := sendNetworkMsg.Message
		// Send data to the other side of the connection
		_, err := conn.Write(message)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by the server.")
			} else {
				fmt.Println("Error sending data to server:", err)
			}
			return
		}
	}
	
}

/*
tcpFSM
for {
	select{
		case <- sendNetworkMsgCh
		case <- reciev
		case <- master	
	}
}*/