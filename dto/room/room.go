package room

import (
	"chinese-chess-backend/dto/user"
)

type RoomInfo struct {
	Id      int           `json:"id"`
	Current user.UserInfo `json:"current"`
	Next    user.UserInfo `json:"next"`
}
