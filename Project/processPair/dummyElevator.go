package dummyElevator

include(
	"Project/network/udp_broadcast"
	"Project/netwprk/Network-go/network/bcast"
	"Project/netwprk/Network-go/network/localip"
	"Project/netwprk/Network-go/network/peers"
)

primaryIP := ""
// Dummy init
	// Elevator init
	// Listen to primary broadcast (UDP)
	// If primary is dead, become primary
	// If primary is alive, save primary IP
	// Open TCP connection to primary, by choosing random port number, encapsulating its own IP and sending it to primary
	// Send full state to primary
	// go sendAliveMessage()

func ProcessPairInit(){
    primaryIPCh := make(chan string)
    udp_broadcast.ListenToBroadcastUntillTimeout(udp_broadcast.PrimaryPort, primaryIPCh)
	switch primaryIPCh {
	case "No primary":
		udp_broadcast.BroadcastMessageLoop(udp_broadcast.PrimaryPort, "I'm primary and i'm alive")
	default:
		primaryIP <- primaryIPCh

	}


}

// Send Im alive message to primary (TCP)

// Listen to primary (TCP)
func DummyToPrimary() {
	port := 30007
	conn, err := connectToServer(primaryIP, port)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to the Primary!")

	// Send a constant message
	localIP := udp_broadcast.GetLocalIP()
	constantMessage := "Connect to " + udp_broadcast.IPToString(localIP) + ":" + port + "\x00" //Husk å endre ip-adresse og port

	_, err = conn.Write([]byte(constantMessage))
	if err != nil {
		fmt.Println("Error sending constant message:", err)
		os.Exit(1)
	}
		fmt.Println("Message sent:", constantMessage)
}

// Do as primary says


func dummyElevatorInit() {
	//Hardware init

	//Start broadcasting


	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Println(err)
		localIP = "DISCONNECTED"
	}
	id = fmt.Sprintf("%s-%d", localIP, os.Getpid())

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			if len(p.Peers) == 1{
				//Become master

			} else {
				//Wait till master initiates contact
			}
		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)
		}
	}
}