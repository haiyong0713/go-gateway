package account

// DecoCards 装扮卡片资源
type DecoCards struct {
	ID           int64       `json:"id"`
	ItemID       int         `json:"item_id"`
	ItemType     int         `json:"item_type"`
	CardURL      string      `json:"card_url"`
	ImageEnhance string      `json:"image_enhance"`
	BigCardURL   string      `json:"big_card_url"`
	CardType     int         `json:"card_type"`
	Name         string      `json:"name"`
	ExpireTime   int         `json:"expire_time"`
	CardTypeName string      `json:"card_type_name"`
	JumpURL      string      `json:"jump_url"`
	Fan          DecorateFan `json:"fan"`
}

type DecorateFan struct {
	IsFan   int    `json:"is_fan"`
	Number  int    `json:"number"`
	Color   string `json:"color"`
	Name    string `json:"name"`
	NumDesc string `json:"num_desc"`
}
