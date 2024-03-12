package messages

import (
	"Project/localElevator/elevio"
	"encoding/json"
	"fmt"
	"strings"
)

const MsgHRAInput = "HRAInput"             // To backup --> to MBDfsm
const MsgElevState = "ElevState"            // To master --> to MBDfsm
const MsgHallReq = "HallReq"              // To master --> to MBDfsm
const MsgHallLigths = "HallLights"           // To elevators --> fsm
const MsgAssignedHallReq = "AssignedHallReq"      // To elevators --> fsm
const MsgRestoreCabReq = "CabReq"               // To elevators --> fsm

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

type dataWithType struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type HallReqMsg struct {
	TAddFRemove bool              `json:"tAdd_fRemove"`
	Floor       int               `json:"floor"`
	Button      elevio.ButtonType `json:"button"`
}


func ToBytes(structType string, msg interface{}) []byte {
	msgJsonBytes, _ := json.Marshal(msg)

	dataToSend := dataWithType{
		Type: structType,
		Data: msgJsonBytes,
	}

	finalJSONBytes, _ := json.Marshal(dataToSend)
	fmt.Println(string(finalJSONBytes), "Structtype: ", structType)
	// Add delimiter to the end of the message
	finalJSONBytes = append(finalJSONBytes, '&')

	return finalJSONBytes
}

func FromBytes(jsonBytes []byte) (string, interface{}) {
	var DataWithType dataWithType
	err := json.Unmarshal(jsonBytes, &DataWithType)
	if err != nil {
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
		var MsgAssignedHallReqData [][2]bool
		err = json.Unmarshal(DataWithType.Data, &MsgAssignedHallReqData)
		return DataWithType.Type, MsgAssignedHallReqData
	case MsgRestoreCabReq:
		var MsgRestoreCabReqData []bool
		err = json.Unmarshal(DataWithType.Data, &MsgRestoreCabReqData)
		return DataWithType.Type, MsgRestoreCabReqData
	default:
		return DataWithType.Type, nil
	}
}

func DistributeMessages(jsonMessageCh chan []byte, toFSMCh chan []byte, toRoleCh chan []byte) {
	var dataWithType dataWithType
	
	for {
		select {
		case jsonMsgReceived := <-jsonMessageCh:
			fmt.Println(string(jsonMsgReceived))
			jsonObjects := strings.Split(string(jsonMsgReceived), "&")

			for _, jsonObject := range jsonObjects {
				if jsonObject != "" {

					err := json.Unmarshal([]byte(jsonObject), &dataWithType)
					if err != nil {
						
						fmt.Println("Error decoding json:", err)
						fmt.Println("jsonObject: ", jsonObject)
						break
					}

					switch dataWithType.Type {
						case MsgHRAInput, MsgElevState, MsgHallReq:
							toRoleCh <- []byte(jsonObject)
							//fmt.Println("Inside DistributeMessages, sent a message to mbdFSM: ", dataWithType.Type)
						case MsgHallLigths, MsgAssignedHallReq, MsgRestoreCabReq:
							toFSMCh <- []byte(jsonObject)
							//fmt.Println("Inside DistributeMessages, sent a message to FSM: ", dataWithType.Type)
					}
				}
				
			}

/*
			fmt.Println("JsonMsgRecieved is: ", string(jsonMsgReceived))
			err := json.Unmarshal(jsonMsgReceived, &dataWithType)
			fmt.Println("Datatypen som unmarshels: ", dataWithType.Type)
			fmt.Println("Dataen som unmarshels: ", string(dataWithType.Data))
			if err != nil {
				
				fmt.Println("Error decoding json:", err)
				os.Exit(55)
			}
			switch dataWithType.Type {
			case MsgHRAInput, MsgElevState, MsgHallReq: // sende til mbdFSM
				toMbdFSMCh <- jsonMsgReceived
				fmt.Println("Inside DistributeMessages, sent a message to mbdFSM: ", dataWithType.Type)
			case MsgHallLigths, MsgAssignedHallReq: // sende til fsm
				toFSMCh <- jsonMsgReceived
				fmt.Println("Inside DistributeMessages, sent a message to FSM: ", dataWithType.Type)
			}*/
		}
	}
}
