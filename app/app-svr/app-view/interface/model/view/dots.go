package view

type DotsReply struct {
	*InteractionManagement `json:"interaction_management"`
	*NoteManagement        `json:"note_management"`
}

type InteractionManagement struct {
	//互动管理按钮是否展示
	CanShow bool `json:"can_show"`
	//互动管理中各个按钮的状态(按顺序下发)
	InteractionStatus []*InteractionStatus `json:"interaction_status"`
}

type InteractionStatus struct {
	//互动管理中展示的按钮名称 danmuku-弹幕 reply-评论 reply_selection-评论精选
	Name string `json:"name"`
	//按钮状态 0允许开,1允许关
	Status int64 `json:"status"`
}

// 笔记按钮管理
type NoteManagement struct {
	// 记笔记按钮是否展示
	CanShow bool  `json:"can_show"`
	Count   int64 `json:"count"`
}
