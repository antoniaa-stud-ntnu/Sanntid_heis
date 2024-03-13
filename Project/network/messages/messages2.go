package messages

import (
	"Project/network/tcp"
	"Project/singleElevator/elevio"
	"encoding/json"
	"fmt"

)

const HRAInputType = "HRAInput"
const MsgElevStateType = "msgElevState"
const MsgHallReqType = "msgHallReq"


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

type msgElevState struct {
	IpAddr    string       `json:"ipAdress"`
	ElevState HRAElevState `json:"elevState"`
}

type msgHallReq struct {
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
	case "HRAInput":
		var HRAInputData HRAInput
		err = json.Unmarshal(DataWithType.Data, &HRAInputData)
		return DataWithType.Type, HRAInputData
	case "msgElevState":
		var MsgElevStateData msgElevState
		err = json.Unmarshal(DataWithType.Data, &MsgElevStateData)
		return DataWithType.Type, MsgElevStateData
	case "msgHallReq":
		var MsgHallReqData msgHallReq
		err = json.Unmarshal(DataWithType.Data, &MsgHallReqData)
		return DataWithType.Type, MsgHallReqData
	default:
		return DataWithType.Type, nil
	}
}