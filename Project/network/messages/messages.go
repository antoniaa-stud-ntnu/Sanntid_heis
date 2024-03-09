package messages

import (
	"Project/singleElevator/elevio"
	"encoding/json"
	"fmt"
)

const MsgHRAInput = "HRAInput"               // To backup --> to MBDfsm
const MsgElevState = "ElevState"             // To master --> to MBDfsm
const MsgHallReq = "HallReq"                 // To master --> to MBDfsm
const MsgHallLigths = "HallLights"           // To elevators --> fsm
const MsgAssignedHallReq = "AssignedHallReq" // To elevators --> to fsm

type HRAElevState struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

type ElevStateMsg struct {
	IpAddr    string       `json:"ipAdress"`
	ElevState HRAElevState `json:"elevState"`
}

type HallReqMsg struct {
	TAddFRemove bool              `json:"tAdd_fRemove"`
	Floor       int               `json:"floor"`
	Button      elevio.ButtonType `json:"button"`
}

type dataWithType struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func ToBytes(structType string, msg interface{}) []byte {
	msgJsonBytes, _ := json.Marshal(msg)

	dataToSend := dataWithType{
		Type: structType,
		Data: msgJsonBytes,
	}

	finalJSONBytes, _ := json.Marshal(dataToSend)
	return finalJSONBytes
}

func FromBytes(jsonBytes []byte) (string, interface{}) {
	var DataWithType dataWithType
	err := json.Unmarshal(jsonBytes, &DataWithType)
	if err != nil {
		// Handle error
		return DataWithType.Type, nil
	}
	switch DataWithType.Type {
	case MsgHRAInput:
		var HRAInputData HRAInput
		err = json.Unmarshal(DataWithType.Data, &HRAInputData)
		return DataWithType.Type, HRAInputData
	case MsgElevState:
		var MsgElevStateData ElevStateMsg
		err = json.Unmarshal(DataWithType.Data, &MsgElevStateData)
		return DataWithType.Type, MsgElevStateData
	case MsgHallReq:
		var MsgHallReqData HallReqMsg
		err = json.Unmarshal(DataWithType.Data, &MsgHallReqData)
		return DataWithType.Type, MsgHallReqData
	case MsgHallLigths:
		var MsgHallLightsData HallReqMsg
		err = json.Unmarshal(DataWithType.Data, &MsgHallLightsData)
		return DataWithType.Type, MsgHallLightsData
	case MsgAssignedHallReq:
		var MsgAssignedHallReq [][2]bool
		err = json.Unmarshal(DataWithType.Data, &MsgAssignedHallReq)
		return DataWithType.Type, MsgAssignedHallReq
	default:
		return DataWithType.Type, nil
	}
}

func DistributeMessages(jsonMessageCh chan []byte, toFSMCh chan []byte, toMbdFSMCh chan []byte) {
	var dataWithType dataWithType

	for {
		select {
		case jsonMsgReceived := <-jsonMessageCh:
			err := json.Unmarshal(jsonMsgReceived, &dataWithType)
			if err != nil {
				fmt.Println("Error decoding json:", err)
			}
			switch dataWithType.Type {
			case MsgHRAInput, MsgElevState, MsgHallReq: // sende til mbdFSM
				toMbdFSMCh <- jsonMsgReceived
			case MsgHallLigths, MsgAssignedHallReq: // sende til fsm
				toFSMCh <- jsonMsgReceived
			}
		}
	}
}
