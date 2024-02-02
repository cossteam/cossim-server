package model

import (
	"encoding/json"
	"fmt"
)

type RoomInfo struct {
	SenderID        string                        `json:"sender_id"`
	RecipientID     string                        `json:"recipient_id"`
	Participants    map[string]*ActiveParticipant `json:"participants"`
	NumParticipants uint32                        `json:"num_participants,omitempty"`
	MaxParticipants uint32                        `json:"max_participants"`
}

type ActiveParticipant struct {
	Connecting bool
}

func (r *RoomInfo) ToMap() (map[string]string, error) {
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

func (r *RoomInfo) FromMap(data map[string]string) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, r)
}
