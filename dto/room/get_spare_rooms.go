package room

type GetSpareRoomsRequest struct {
	Infos []RoomInfo `json:"-"`
}

type GetSpareRoomsResponse struct {
	Rooms []RoomInfo `json:"rooms"`
}