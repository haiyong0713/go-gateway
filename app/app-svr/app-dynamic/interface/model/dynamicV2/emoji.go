package dynamicV2

// emoji 资源
type Emoji struct {
	Emote map[string]*EmojiItem `json:"emote"`
}

type EmojiItem struct {
	ID        int `json:"id"`
	PackageID int `json:"package_id"`
	State     int `json:"state"`
	Type      int `json:"type"`
	//Attr      int       `json:"attr"`
	Text  string    `json:"text"`
	URL   string    `json:"url"`
	Meta  EmojiMeta `json:"meta"`
	Mtime int       `json:"mtime"`
}

type EmojiMeta struct {
	Size            int    `json:"size"`
	LabelText       string `json:"label_text"`
	LabelURL        string `json:"label_url"`
	LabelColor      string `json:"label_color"`
	LabelGuideTitle string `json:"label_guide_title"`
	LabelGuideText  string `json:"label_guide_text"`
}
