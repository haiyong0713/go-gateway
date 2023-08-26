package mall

const (
	TabNo  = 0
	TabYes = 1
)

// Mall struct.
type Mall struct {
	ShopID   int64  `json:"shopId,omitempty"`
	VAPPID   string `json:"vAppId,omitempty"`
	APPID    string `json:"appId,omitempty"`
	Name     string `json:"name,omitempty"`
	URL      string `json:"jumpUrl,omitempty"`
	Logo     string `json:"logo,omitempty"`
	TabState int8   `json:"showItemsTab,omitempty"`
}
