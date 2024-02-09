package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"
)

func main() {

	var counter byte = 0;

	listenCh := make(chan byte)
	quit     := make(chan bool)

	flagPtr := flag.Int("Flag", 0, "wæææ")
	flag.Parse()

	go UDPBackup.UDPListen(listenCh, quit, *flagPtr)

Backup:
	for {
		select {
		case <-time.After(2 * time.Second):
			break Backup
		case counter = <-listenCh:
			break
		}
	}

	// Spawn backup process
	close(quit)
	launchBackupProcess(*flagPtr)

	fmt.Println("Active Mode")
	for {
		fmt.Println(counter)
		counter++
		UDPBackup.UDPSend(counter, *flagPtr)
		time.Sleep(500 * time.Millisecond)
	}
}


func launchBackupProcess(flag int) {
	flag = UDPBackup.NotEqual(flag)
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run processPairs.go -flag="+strconv.Itoa(flag))
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}