package websocket

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"

	"chinese-chess-backend/dto"
	"chinese-chess-backend/utils"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求，生产环境应该限制
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type websocketQuery struct {
	Token string `form:"token"`
}

type clientStatus int

const (
	Online clientStatus = iota + 1
	Playing
	Matching
)

type Client struct {
	Conn   *websocket.Conn
	Id     int
	Status clientStatus
	RoomId int
}

type ChessRoom struct {
	Current *Client // 先进入房间的作为先手，默认为当前玩家
	Next    *Client // 后进入房间的作为后手，默认为下一个玩家
}

func NewChessRoom() *ChessRoom {
	return &ChessRoom{
		Current: nil,
		Next:    nil,
	}
}

type ChessHub struct {
	Rooms     map[int](*ChessRoom)
	Clients   map[int]*Client
	NextId    int
	commands  chan hubCommand
	spareRooms []int // 有空位的房间id
}

func NewChessHub() *ChessHub {
	return &ChessHub{
		Rooms:    make(map[int](*ChessRoom)),
		Clients:  make(map[int]*Client),
		NextId:   0,
		commands: make(chan hubCommand),
	}
}

func (ch *ChessHub) Run() {
	for cmd := range ch.commands {
		switch cmd.commandType {
		case register:
			client := cmd.client
			ch.Clients[client.Id] = client
		case unregister:
			client := cmd.client
			roomId := client.RoomId
			if room, ok := ch.Rooms[roomId]; ok {
				room.Current = nil
				room.Next = nil
				delete(ch.Rooms, roomId)
			}
			if _, ok := ch.Clients[client.Id]; ok {
				delete(ch.Clients, client.Id)
				client.Conn.Close()
			}
		case match:
			client := cmd.client
			if len(ch.spareRooms) == 0 {
				// 没有空闲房间，创建一个新的房间
				room := NewChessRoom()
				room.Current = client
				client.RoomId = ch.NextId
				ch.Rooms[ch.NextId] = room
				ch.NextId++
				ch.spareRooms = append(ch.spareRooms, client.RoomId)
				continue
			}
			// 有空闲房间，加入到空闲房间中
			roomId := ch.spareRooms[0]
			ch.spareRooms = ch.spareRooms[1:]
			room := ch.Rooms[roomId]
			if room == nil {
				fmt.Println("房间不存在")
				continue
			}
			if room.Current == nil {
				room.Current = client
				client.RoomId = roomId
			} else if room.Next == nil {
				room.Next = client
				client.RoomId = roomId
			} else {
				fmt.Println("房间已满")
				continue
			}
		case move:
			req := cmd.payload.(moveRequest)
			room := ch.Rooms[req.from.RoomId]
			if room == nil {
				fmt.Println("房间不存在")
				continue
			}

			var target *Client
			if room.Current == req.from {
				target = room.Next
			} else {
				target = room.Current
			}

			ch.sendMessageInternal(target, req.move)

			// 交换当前玩家和下一个玩家
			if room.Current == req.from {
				room.Current = room.Next
				room.Next = req.from
			} else {
				room.Next = room.Current
				room.Current = req.from
			}
		case sendMessage:
			req := cmd.payload.(sendMessageRequest)
			err := req.target.Conn.WriteJSON(req.message)
			if err != nil {
				fmt.Printf("发送消息失败: %v\n", err)
			}

		case start:
			room := ch.Rooms[cmd.client.RoomId]
			if room == nil {
				fmt.Println("房间不存在")
				continue
			}
			if room.Current == nil || room.Next == nil {
				fmt.Println("房间未满员，无法开始游戏")
				continue
			}
			room.Current.Status = Playing
			room.Next.Status = Playing
			startMessage := BaseMessage{Type: Start}
			ch.sendMessageInternal(room.Current, startMessage)
			ch.sendMessageInternal(room.Next, startMessage)
		case end:
			room := ch.Rooms[cmd.client.RoomId]
			if room == nil {
				fmt.Println("房间不存在")
				continue
			}
			room.Current.Status = Online
			room.Next.Status = Online
			room.Current.RoomId = -1
			room.Next.RoomId = -1
			// 发送消息给两个客户端，通知他们结束游戏
			endMessage := BaseMessage{Type: End}
			ch.sendMessageInternal(room.Current, endMessage)
			ch.sendMessageInternal(room.Next, endMessage)
			room.Current = nil
			room.Next = nil
			delete(ch.Rooms, cmd.client.RoomId)
		}

	}
}

func (ch *ChessHub) HandleConnection(c *gin.Context) {
	var query websocketQuery
	err := c.ShouldBindQuery(&query)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("未登录或token错误"))
		return
	}
	fmt.Println(1)
	
	if query.Token == "" {
		dto.ErrorResponse(c, dto.WithMessage("未登录或token错误"))
		return
	}
	fmt.Println(2)
	
	// 验证token是否有效
	id := utils.ParseToken(query.Token)
	fmt.Println("id", id)
	if id <= 0 {
		dto.ErrorResponse(c, dto.WithMessage("未登录或token错误"))
		return
	}
	fmt.Println(3)
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	fmt.Println(4)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("websocket upgrade error"))
		return
	}
	defer conn.Close()
	fmt.Println(5)

	// 创建一个新的客户端
	client := &Client{
		Conn:   conn,
		Id:     id,
		Status: Online,
		RoomId: -1,
	}

	ch.commands <- hubCommand{
		commandType: register,
		client:      client,
	}
	defer func() {
		ch.commands <- hubCommand{
			commandType: unregister,
			client:      client,
		}
	}()

	for {
		fmt.Println("等待消息...")
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("读取消息失败: %v\n", err)
			break
		}

		err = ch.handleMessage(client, message)
		if err != nil {
			fmt.Printf("处理消息失败: %v\n", err)
			return
		}
	}
	fmt.Println("客户端断开连接")
}

func (ch *ChessHub) handleMessage(client *Client, rawMessage []byte) error {
	var base BaseMessage
	err := json.Unmarshal(rawMessage, &base)
	if err != nil {
		return fmt.Errorf("解析消息失败: %v", err)
	}
	fmt.Println("解析消息成功", base)

	switch base.Type {
	case Match:
		switch client.Status {
		case Online:
			client.Status = Matching
			ch.commands <- hubCommand{
				commandType: match,
				client:      client,
			}
		case Matching:
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: Normal},
				Message:     "您已在匹配队列中，请耐心等待",
			}
			ch.sendMessage(client, msg)
		case Playing:
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: Normal},
				Message:     "您已在游戏中，请耐心等待",
			}
			ch.sendMessage(client, msg)
		}
	case Move:
		if client.Status == Playing {
			var moveMsg MoveMessage
			err := json.Unmarshal(rawMessage, &moveMsg)
			if err != nil {
				fmt.Printf("解析移动消息失败: %v\n", err)
				return err
			}

			ch.commands <- hubCommand{
				commandType: move,
				client:      client,
				payload: moveRequest{
					from: client,
					move: moveMsg,
				},
			}
		} else {
			return fmt.Errorf("玩家不在游戏中")
		}
	case End:
		if client.Status == Playing {
			ch.commands <- hubCommand{
				commandType: end,
				client:      client,
			}
		}
	}

	return nil
}

func (ch *ChessHub) sendMessage(client *Client, message any) {
	ch.commands <- hubCommand{
		commandType: sendMessage,
		payload: sendMessageRequest{
			target:  client,
			message: message,
		},
	}
}

func (ch *ChessHub) sendMessageInternal(client *Client, message any) {
	err := client.Conn.WriteJSON(message)
	if err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
	}
}
