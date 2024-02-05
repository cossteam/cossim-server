package model

import (
	"encoding/json"
	"fmt"
)

type UserRoomInfo struct {
	SenderID        string                        `json:"sender_id"`
	RecipientID     string                        `json:"recipient_id"`
	NumParticipants uint32                        `json:"num_participants"`
	MaxParticipants uint32                        `json:"max_participants"`
	Participants    map[string]*ActiveParticipant `json:"participants"`
}

func (r *UserRoomInfo) ToMap() (map[string]string, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	for key, value := range result {
		if key == "participants" {
			participantsJSON, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			resultMap[key] = string(participantsJSON)
		} else {
			resultMap[key] = fmt.Sprintf("%v", value)
		}
	}

	return resultMap, nil
}

func (r *UserRoomInfo) ToInterface() (map[string]interface{}, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *UserRoomInfo) ToJSONString() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *UserRoomInfo) FromMap(data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, r)
}

type GroupRoomInfo struct {
	GroupID         uint32                        `json:"group_id"`
	SenderID        string                        `json:"sender_id"`
	Participants    map[string]*ActiveParticipant `json:"participants"`
	NumParticipants uint32                        `json:"num_participants"`
	MaxParticipants uint32                        `json:"max_participants"`
}

type ActiveParticipant struct {
	Connecting bool
}

func (r *GroupRoomInfo) ToMap() (map[string]string, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	for key, value := range result {
		if key == "participants" {
			participantsJSON, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			resultMap[key] = string(participantsJSON)
		} else {
			resultMap[key] = fmt.Sprintf("%v", value)
		}
	}

	return resultMap, nil
}

func (r *GroupRoomInfo) ToInterface() (map[string]interface{}, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *GroupRoomInfo) ToJSONString() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *GroupRoomInfo) FromMap(data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, r)
}
