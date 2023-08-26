package note

type ReplyAddRes struct {
	Code int64            `json:"code"`
	Data *ReplyAddResData `json:"data"`
}

type ReplyAddResData struct {
	Rpid int64 `json:"rpid"`
}
