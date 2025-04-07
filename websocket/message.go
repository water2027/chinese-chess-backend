package websocket

type MessageType int

// 使用前缀，区分command
const (
	messageNormal MessageType = iota + 1
	messageMatch
	messageMove
	messageStart
	messageEnd
	messageJoin
	messageCreate
	messageError = 10
)

type BaseMessage struct {
	Type MessageType `json:"type"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MoveMessage struct {
	BaseMessage
	From Position `json:"from"`
	To   Position `json:"to"`
}

type NormalMessage struct {
	BaseMessage
	Message string `json:"message"`
}

type startMessage struct {
	BaseMessage
	Role string `json:"role"`
}

type joinMessage struct {
	BaseMessage
	RoomId int `json:"roomId"`
}

type endMessage struct {
	BaseMessage
	Winner clientRole `json:"winner"`
}
