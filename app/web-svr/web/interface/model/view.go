package model

// dmVoteReq struct
type DmVoteReq struct {
	AID      int64  `form:"aid" validate:"min=1"`
	CID      int64  `form:"cid" validate:"min=1"`
	Vote     int32  `form:"vote" validate:"min=1"`
	VoteID   int64  `form:"vote_id" validate:"min=1"`
	Progress int32  `form:"progress" validate:"min=1"`
	Mid      int64  `form:"-"`
	Buvid    string `form:"-"`
}

type DmVoteReply struct {
	Vote *VoteReply `json:"vote,omitempty"`
	Dm   *DmReply   `json:"dm,omitempty"`
}

type VoteReply struct {
	UID  int64 `json:"uid"`
	Type int32 `json:"type"`
	Vote int32 `json:"vote"`
}

type DmReply struct {
	DmID      int64  `json:"dm_id"`
	DmIDStr   string `json:"dm_id_str"`
	Visible   bool   `json:"visible"`
	Action    string `json:"action"`
	Animation string `json:"animation"`
}
