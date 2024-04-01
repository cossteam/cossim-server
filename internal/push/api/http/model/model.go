package model

type OnlineEventData struct {
	DriverType string `json:"driver_type"`
}

type FriendOnlineStatusMsg struct {
	UserId string `json:"user_id"`
	Status int32  `json:"status"`
}
