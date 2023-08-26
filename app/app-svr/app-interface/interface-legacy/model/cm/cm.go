package cm

import "encoding/json"

type Ad struct {
	ShowShopTab   bool            `json:"show_shop_tab,omitempty"`
	ShopTabType   int             `json:"shop_tab_type,omitempty"`
	SourceContent json.RawMessage `json:"source_content,omitempty"`
}

type Topic struct {
	TopicName string `json:"topic_name"`
	TopicID   int64  `json:"topic_id"`
	CoverURL  string `json:"cover_url"`
	CoverMd5  string `json:"cover_md5"`
	View      int64  `json:"view"`
	Discuss   int64  `json:"discuss"`
	JumpUrl   string `json:"jump_url"`
	TopicType int64  `json:"topic_type"`
}

type PickupEntrance struct {
	JumpUrl        string `json:"jump_url"`
	IsShowEntrance bool   `json:"show_entrance"`
}
