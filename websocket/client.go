package websocket

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type clientStatus int

const (
	userOnline clientStatus = iota + 1
	userPlaying
	userMatching
)

type clientRole int

const (
	// 0表示没有角色，1表示红方，2表示黑方
	roleNone clientRole = iota
	roleRed 
	roleBlack
)

type Client struct {
	Conn     *websocket.Conn
	Id       int
	Status   clientStatus
	RoomId   int
	Role     clientRole // 角色
	LastPong time.Time // 上次收到PONG的时间
}

func NewClient(conn *websocket.Conn, id int) *Client {
	return &Client{
		Conn:     conn,
		Id:       id,
		Status:   userOnline,
		RoomId:   -1,
		Role:     roleNone,
		LastPong: time.Now(),
	}
}

func (c *Client) sendMessage(message any) error {
	if c.Conn == nil {
		return fmt.Errorf("client connection is nil")
	}
	err := c.Conn.WriteJSON(message)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) startPlay(role clientRole) {
	c.Role = role
	c.Status = userPlaying
}
