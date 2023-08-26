package model

type MemberPrivacyMsg struct {
	Action string         `json:"action"`
	Table  string         `json:"table"`
	New    *MemberPrivacy `json:"new"`
	Old    *MemberPrivacy `json:"old"`
}

type MemberPrivacy struct {
	// Comment: 主键
	ID int64 `json:"id"`
	// Comment: 用户id
	Mid int64 `json:"mid"`
	// Comment: 隐私设置
	Privacy string `json:"privacy"`
	// Comment: 状态： 0：隐藏 1：展示
	Status int64 `json:"status"`
	// Comment: 修改时间
	// Default: CURRENT_TIMESTAMP
	ModifyTime string `json:"modify_time"`
	NewUser    bool   `json:"new_user"`
}
