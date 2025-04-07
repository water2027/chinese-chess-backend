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

	"chinese-chess-backend/database"
	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/room"
	"chinese-chess-backend/dto/user"
	"chinese-chess-backend/utils"
	"slices"
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

type ChessHub struct {
	Rooms      map[int](*ChessRoom)
	Clients    map[int]*Client
	commands   chan hubCommand
	spareRooms []room.RoomInfo // 有空位的房间id
	mu         sync.Mutex
	pool       *utils.WorkerPool
	matchPool  [](*Client)
}

func NewChessHub() *ChessHub {
	pool := utils.NewWorkerPool()
	hub := &ChessHub{
		Rooms:      make(map[int](*ChessRoom)),
		Clients:    make(map[int]*Client),
		commands:   make(chan hubCommand),
		spareRooms: make([]room.RoomInfo, 0),
		mu:         sync.Mutex{},
		pool:       pool,
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
			switch cmd.commandType {
			case register:
				client := cmd.client
				ch.mu.Lock()
				ch.Clients[client.Id] = client
				ch.mu.Unlock()
				// 在线用户
				database.SetValue(fmt.Sprint(client.Id), "a", 0)
			case unregister:
				client := cmd.client
				roomId := client.RoomId
				ch.mu.Lock()
				room, ok := ch.Rooms[roomId]
				ch.mu.Unlock()
				if ok {
					var target *Client
					if room.Current == client {
						target = room.Next
					} else {
						target = room.Current
					}
					if target != nil {
						ch.sendMessage(target, NormalMessage{
							BaseMessage: BaseMessage{Type: Normal},
							Message:     "对方已断开连接",
						})
					}
					room.clear()
					ch.mu.Lock()
					delete(ch.Rooms, roomId)
					// 如果房间原本只有一个人，那么删除房间
					for i, r := range ch.spareRooms {
						if r.Id == roomId {
							ch.spareRooms = slices.Delete(ch.spareRooms, i, i+1)
							break
						}
					}
					ch.mu.Unlock()
				}
				ch.mu.Lock()
				if _, ok := ch.Clients[client.Id]; ok {
					delete(ch.Clients, client.Id)
					client.Conn.Close()
				}
				ch.mu.Unlock()
				database.DeleteValue(fmt.Sprint(client.Id))
			case match: 
				client := cmd.client
				ch.mu.Lock()
				ch.matchPool = append(ch.matchPool, client)
				fmt.Println(ch.matchPool)
				if len(ch.matchPool) < 2 {
					ch.sendMessageInternal(client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "正在匹配，请稍等",
					})
					ch.mu.Unlock()
					return nil
				}
				// 匹配成功，创建房间
				room := NewChessRoom()
				room.join(ch.matchPool[0])
				room.join(ch.matchPool[1])
				ch.matchPool = ch.matchPool[2:]
				ch.Rooms[room.Id] = room
				ch.mu.Unlock()
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

				if !room.isFull() {
					ch.sendMessageInternal(req.from, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "游戏未开始",
					})
					return nil
				}

				if room.Current != req.from {
					// 如果不是当前玩家，则不允许移动
					ch.sendMessageInternal(req.from, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "请等待对方移动",
					})
					return nil
				}

				target := room.Next

				ch.sendMessageInternal(target, req.move)

				// 交换当前玩家和下一个玩家
				room.exchange()
			case sendMessage:
				req := cmd.payload.(sendMessageRequest)
				err := req.target.sendMessage(req.message)
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
				if !room.isFull() {
					ch.sendMessageInternal(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间未满员，无法开始游戏",
					})
					return nil
				}
				room.Current.startPlay(Red)
				room.Next.startPlay(Black)
				cur := startMessage{BaseMessage: BaseMessage{Type: Start}, Role: "red"}
				next := startMessage{BaseMessage: BaseMessage{Type: Start}, Role: "black"}
				ch.sendMessageInternal(room.Current, cur)
				ch.sendMessageInternal(room.Next, next)
				// 移除空余房间
				ch.mu.Lock()
				for i, r := range ch.spareRooms {
					if r.Id == room.Id {
						ch.spareRooms = slices.Delete(ch.spareRooms, i, i+1)
						break
					}
				}
				ch.mu.Unlock()
			case end:
				room := ch.Rooms[cmd.client.RoomId]
				if room == nil {
					ch.sendMessageInternal(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间不存在",
					})
					return nil
				}
				winner := room.Next.Role
				// 发送消息给两个客户端，通知他们结束游戏
				endMsg := endMessage{
					BaseMessage: BaseMessage{Type: End},
					Winner:      winner,
				}
				ch.sendMessageInternal(room.Current, endMsg)
				ch.sendMessageInternal(room.Next, endMsg)
				room.clear()
				delete(ch.Rooms, cmd.client.RoomId)
			case heartbeat:
				// 更新客户端的最后一次心跳时间
				client := cmd.client
				client.LastPong = time.Now()
			case join:
				joinMsg := cmd.payload.(joinMessage)
				ch.mu.Lock()
				room := ch.Rooms[joinMsg.RoomId]
				if room == nil {
					ch.sendMessage(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     "房间不存在",
					})
					ch.mu.Unlock()
					return nil
				}
				err := room.join(cmd.client)
				if err != nil {
					ch.sendMessageInternal(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: Normal},
						Message:     err.Error(),
					})
					ch.mu.Unlock()
					return nil
				}
				ch.mu.Unlock()
				// 发送消息给两个客户端，通知他们开始游戏
				go func() {
					ch.commands <- hubCommand{
						commandType: start,
						client:      cmd.client,
					}
				}()
			case create:
				// 创建房间
				client := cmd.client
				r := NewChessRoom()
				r.join(client)
				ch.Rooms[r.Id] = r
				roomInfo := room.RoomInfo{
					Id: client.RoomId,
					Current: user.UserInfo{
						ID: uint(client.Id),
					},
				}
				ch.mu.Lock()
				ch.spareRooms = append(ch.spareRooms, roomInfo)
				ch.mu.Unlock()
				// 发送消息给客户端，通知他们创建房间成功
				ch.sendMessage(client, NormalMessage{
					BaseMessage: BaseMessage{Type: Create},
				})
				return nil
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
	client := NewClient(conn, id)

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

func (ch *ChessHub) GetSpareRooms(c *gin.Context) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	c.Set("rooms", ch.spareRooms)
	c.Next()
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
				Message:     "您已在游戏中",
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
	case Join:
		// 用户加入房间
		if client.Status == Playing {
			// 如果用户已经在游戏中，则不允许加入房间
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: Normal},
				Message:     "您已在游戏中",
			}
			ch.sendMessage(client, msg)
			return nil
		}
		var joinMsg joinMessage
		err := json.Unmarshal(rawMessage, &joinMsg)
		if err != nil {
			fmt.Printf("解析加入房间消息失败: %v\n", err)
			return nil
		}
		ch.commands <- hubCommand{
			commandType: join,
			client:      client,
			payload:     joinMsg,
		}
	case Create:
		// 用户创建房间
		if client.Status == Playing {
			// 如果用户已经在游戏中，则不允许创建房间
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: Normal},
				Message:     "您已在游戏中",
			}
			ch.sendMessage(client, msg)
			return nil
		}
		ch.commands <- hubCommand{
			commandType: create,
			client:      client,
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
