package msg_queue

const Service_Exchange = "service_exchange"

type ServiceType string

const (
	MsgService      ServiceType = "msg_service"
	GroupService    ServiceType = "group_service"
	UserService     ServiceType = "user_service"
	RelationService ServiceType = "relation_service"
)

type ServiceActionType uint

const (
	Notice = iota
	SendMessage
)

type ServiceQueueMsg struct {
	Form   ServiceType       `json:"form"`
	Action ServiceActionType `json:"action"`
	Data   interface{}       `json:"data"`
}
