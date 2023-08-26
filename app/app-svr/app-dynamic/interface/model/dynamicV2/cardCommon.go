package dynamicV2

import "encoding/json"

const (
	// 跳转类型
	AttachButtonTypeLink = 1
	// 开关类型
	AttachButtonTypeSwitch = 2
	// 未点
	AttachButtonStatusUncheck = 1
	// 已点
	AttachButtonStatusCheck = 2
)

type DynamicCommon struct {
	RID  int64  `json:"rid"`
	Card string `json:"card"`
}

type DynamicCommonCard struct {
	RID  int64 `json:"rid"`
	User *struct {
		UID   int64  `json:"uid"`
		UName string `json:"uname"`
		Face  string `json:"face"`
	} `json:"user"`
	Vest *struct {
		UID     int64  `json:"uid"`
		Content string `json:"content"`
		Ctrl    string `json:"ctrl"` // 返回string "[]"
	} `json:"vest"`
	Sketch *struct {
		Title    string            `json:"title"`
		DescText string            `json:"desc_text"`
		CoverURL string            `json:"cover_url"`
		TagURL   string            `json:"target_url"`
		SketchID int64             `json:"sketch_id"`
		BizType  int               `json:"biz_type"`
		Tags     json.RawMessage   `json:"tags"` // 返回[]
		Text     string            `json:"text"`
		BizID    int64             `json:"biz_id"`
		Button   *AttachCardButton `json:"button"`
	} `json:"sketch"`
}

type AttachCardButton struct {
	Type      int          `json:"type"`
	JumpStyle *ButtonStyle `json:"jump_style"`
	JumpURL   string       `json:"jump_url"`
	Uncheck   *ButtonStyle `json:"uncheck"`
	Check     *ButtonStyle `json:"check"`
	Status    int          `json:"status"`
}

type ButtonStyle struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type DynamicCommonCardTags struct {
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
