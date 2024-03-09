package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex // A mutex for synchronization
var _conn net.Conn  // Holds a network connection

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

type MotorDirection int

const (
	Down MotorDirection = -1
	Stop MotorDirection = 0
	Up   MotorDirection = 1
)

type ButtonType int

const (
	HallUp   ButtonType = 0
	HallDown ButtonType = 1
	Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}

	// Attempt to establish a network connection to the specified address
	var err error // outputDevice.MotorDirection(Down)
	// elevio.SetMotorDirection(elevio.Down)
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetButtonLamp(floor int, btn ButtonType, val bool) {
	write([4]byte{2, byte(btn), byte(floor), toByte(val)}) // Other use this
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(val bool) {
	write([4]byte{4, toByte(val), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollRequestButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func getButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

// func GetFloor() int {
// 	a := read([4]byte{7, 0, 0, 0})
// 	if a[1] != 0 {
// 		return int(a[2])
// 	} else {
// 		return -1
// 	}
// }

func GetFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func getObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
