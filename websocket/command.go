package websocket

type CommendType int

const (
	register    CommendType = iota + 1 // 注册
	unregister                         // 注销
	match                              // 匹配
	move                               // 移动
	sendMessage                        // 发送消息
	start                              // 开始游戏
	end                                // 结束游戏
	join                               // 加入房间
	create                             // 创建房间
	heartbeat                          // 心跳
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
