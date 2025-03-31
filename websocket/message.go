package websocket

type MessageType int

const (
	Normal MessageType = iota + 1
	Match
	Move
	Start
	End
	Error = 10
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