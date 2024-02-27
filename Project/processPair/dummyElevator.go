package dummyElevator

include(
	"Project/network/udp_broadcast"
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