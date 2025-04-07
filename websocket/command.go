package websocket

type CommendType int

const (
	commandRegister    CommendType = iota + 1 // 注册
	commandUnregister                         // 注销
	commandMatch                              // 匹配
	commandMove                               // 移动
	commandSendMessage                        // 发送消息
	commandStart                              // 开始游戏
	commandEnd                                // 结束游戏
	commandJoin                               // 加入房间
	commandCreate                             // 创建房间
	commandHeartbeat                          // 心跳
)

type moveRequest struct {
	from *Client
	move MoveMessage
}

type sendMessageRequest struct {
	target  *Client
	message any
}

type hubCommand struct {
	commandType CommendType
	client      *Client
	payload     any
}
