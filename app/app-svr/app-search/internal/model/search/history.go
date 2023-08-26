package search

// LiveParam statue param
type LiveParam struct {
	RoomIDs    string `form:"room_ids"`
	Uid        int64
	Platform   string
	ReqBiz     string
	DeviceName string
	NetWork    string
	Build      int64
}
