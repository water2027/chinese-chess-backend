package room

import (
	userModel "chinese-chess-backend/model/user"
)

type RoomInfo struct {
	Id      int  `json:"id"`
	Current userModel.User `json:"current"`
	Next    userModel.User `json:"next"`
}