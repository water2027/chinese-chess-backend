package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/room"
	"chinese-chess-backend/service"
)

type RoomController struct {
	roomService *service.RoomService
}

func NewRoomController(roomService *service.RoomService) *RoomController {
	return &RoomController{
		roomService: roomService,
	}
}

func (rc *RoomController) GetSpareRooms(c *gin.Context) {
	info, exists := c.Get("rooms")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("room not found"))
		return
	}
	infos, ok := info.([]room.RoomInfo)
	if !ok {
		dto.ErrorResponse(c, dto.WithMessage("room not found"))
		return
	}
	fmt.Println(infos)
	resp, err := rc.roomService.GetSpareRooms(room.GetSpareRoomsRequest{
		Infos: infos})
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}
