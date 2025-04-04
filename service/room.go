package service

import (
	"chinese-chess-backend/database"
	"chinese-chess-backend/dto/room"
	userModel "chinese-chess-backend/model/user"
)

type RoomService struct{}

func NewRoomService() *RoomService {
	return &RoomService{}
}

func (rs *RoomService) GetSpareRooms(req room.GetSpareRoomsRequest) (room.GetSpareRoomsResponse, error) {
    db := database.GetMysqlDb()
	rooms := req.Infos
	var resp room.GetSpareRoomsResponse
    
    // Collect all user IDs that need to be fetched
    var userIDs []uint
    userIDPositionMap := make(map[uint][]struct{
        roomIndex int
        isCurrentUser bool
    })
    
    for i, room := range rooms {
        if room.Current.ID != 0 {
            id := uint(room.Current.ID)
            userIDs = append(userIDs, id)
            userIDPositionMap[id] = append(userIDPositionMap[id], struct{
                roomIndex int
                isCurrentUser bool
            }{i, true})
        }
        
        if room.Next.ID != 0 {
            id := uint(room.Next.ID)
            userIDs = append(userIDs, id)
            userIDPositionMap[id] = append(userIDPositionMap[id], struct{
                roomIndex int
                isCurrentUser bool
            }{i, false})
        }
    }
    
    // If no users to fetch, return the original array
    if len(userIDs) == 0 {
        return resp, nil
    }
    
    // Fetch all users at once with only the needed fields
    var users []userModel.User
    if err := db.Model(&userModel.User{}).
        Select("id, name, exp").
        Where("id IN ?", userIDs).
        Find(&users).Error; err != nil {
        return resp, err
    }
    
    // Create a map for quick lookup
    userMap := make(map[uint]userModel.User)
    for _, user := range users {
        userMap[user.ID] = user
    }
    
    // Update rooms with user information
    for _, user := range users {
        positions := userIDPositionMap[user.ID]
        for _, pos := range positions {
            if pos.isCurrentUser {
                rooms[pos.roomIndex].Current.Name = user.Name
                rooms[pos.roomIndex].Current.Exp = user.Exp
            } else {
                rooms[pos.roomIndex].Next.Name = user.Name
                rooms[pos.roomIndex].Next.Exp = user.Exp
            }
        }
    }

	resp.Rooms = rooms
	return resp, nil
}
