package websocket

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"chinese-chess-backend/dto"
	"chinese-chess-backend/utils"
)

const (
	HeartbeatInterval = 5 * time.Second  // 发送心跳的间隔
	HeartbeatTimeout  = 30 * time.Second // 心跳超时时间
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求，生产环境应该限制
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type clientStatus int

const (
	Online clientStatus = iota + 1
	Playing
	Matching
)

type Client struct {
	Conn     *websocket.Conn
	Id       int
	Status   clientStatus
	RoomId   int
	LastPong time.Time // 上次收到PONG的时间
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
	Rooms      map[int](*ChessRoom)
	Clients    map[int]*Client
	NextId     int
	commands   chan hubCommand
	spareRooms []int // 有空位的房间id
	mu         sync.Mutex
	pool       *utils.WorkerPool
}

func NewChessHub() *ChessHub {
	pool := utils.NewWorkerPool()
	hub := &ChessHub{
		Rooms:    make(map[int](*ChessRoom)),
		Clients:  make(map[int]*Client),
		NextId:   0,
		commands: make(chan hubCommand),
		spareRooms: make([]int, 0),
		mu:       sync.Mutex{},
		pool:     pool,
	}
	pool.Start()

	return hub
}

func (ch *ChessHub) Run() {
	go func() {
		for err := range ch.pool.ErrChan {
			log.Printf("Worker pool error: %v\n", err)
		}
	}()
	for cmd := range ch.commands {
		ch.pool.Process(context.Background(), func() error {
			ch.mu.Lock()
			defer ch.mu.Unlock()
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
					return nil
				}
				// 有空闲房间，加入到空闲房间中
				roomId := ch.spareRooms[0]
				ch.spareRooms = ch.spareRooms[1:]
				room := ch.Rooms[roomId]
				if room == nil {
					ch.sendMessageInternal(client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间不存在",
					})
					return nil
				}
				if room.Current == nil {
					room.Current = client
					client.RoomId = roomId
				} else if room.Next == nil {
					room.Next = client
					client.RoomId = roomId
				} else {
					ch.sendMessageInternal(client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间已满",
					})
					return nil
				}
				// 发送消息给两个客户端，通知他们开始游戏
				go func() {
					ch.commands <- hubCommand{
						commandType: start,
						client:      client,
					}
				}()
			case move:
				req := cmd.payload.(moveRequest)
				room := ch.Rooms[req.from.RoomId]
				if room == nil {
					ch.sendMessageInternal(req.from, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间不存在",
					})
					return nil
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
					return fmt.Errorf("发送消息失败: %v", err)
				}
	
			case start:
				room := ch.Rooms[cmd.client.RoomId]
				if room == nil {
					cmd.client.RoomId = -1
					cmd.client.Status = Online
					ch.sendMessageInternal(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "请进行匹配",
					})
					return nil
				}
				if room.Current == nil || room.Next == nil {
					ch.sendMessageInternal(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间未满员，无法开始游戏",
					})
					return nil
				}
				room.Current.Status = Playing
				room.Next.Status = Playing
				cur := startMessage{BaseMessage: BaseMessage{Type: Start}, Role: "red"}
				next := startMessage{BaseMessage: BaseMessage{Type: Start}, Role: "black"}
				ch.sendMessageInternal(room.Current, cur)
				ch.sendMessageInternal(room.Next, next)
			case end:
				room := ch.Rooms[cmd.client.RoomId]
				if room == nil {
					ch.sendMessageInternal(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间不存在",
					})
					return nil
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
			case heartbeat:
				// 更新客户端的最后一次心跳时间
				client := cmd.client
				client.LastPong = time.Now()
			}
			return nil
		})
	}
}

func (ch *ChessHub) HandleConnection(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("用户未登录"))
		return
	}

	id, ok := userId.(int)
	if !ok {
		dto.ErrorResponse(c, dto.WithMessage("用户ID转换失败"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("websocket upgrade error"))
		return
	}
	defer conn.Close()

	// 创建一个新的客户端
	client := &Client{
		Conn:   conn,
		Id:     id,
		Status: Online,
		RoomId: -1,
	}

	conn.SetReadLimit(1024 * 1024)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
		client.LastPong = time.Now()
		return nil
	})
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Printf("WebSocket connection closed with code %d: %s\n", code, text)
		return nil
	})

	conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))

	go func() {
		ticker := time.NewTicker(HeartbeatInterval)
		defer ticker.Stop()

		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("发送 ping 失败: %v\n", err)
				return
			}
		}
	}()

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

	ch.sendMessage(client, NormalMessage{
		BaseMessage: BaseMessage{Type: Normal},
		Message:     "连接成功",
	})

	for {
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
