package websocket

type MessageType int

// 信息类型
const (
	messageNormal MessageType = iota + 1 // 普通消息
	messageMatch                         // 匹配消息
	messageMove                          // 移动消息
	messageStart                         // 开始消息
	messageEnd                           // 结束消息
	messageJoin                          // 加入消息
	messageCreate                        // 创建房间消息
	messageGiveUp                        // 放弃消息
	messageError  = 10
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
