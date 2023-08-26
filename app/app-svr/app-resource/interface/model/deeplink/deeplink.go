package deeplink

type HWDeeplinkReq struct {
	Oaid   string `form:"oaid" json:"oaid" validate:"required"`
	Bundle string `form:"bundle" json:"bundle" validate:"required"`
}

type HWDeeplinkMeta struct {
	Deeplink string `json:"deeplink"`
}

type BackButtonReq struct {
	Schema string `form:"schema" validate:"required"`
}

type BackButtonMeta struct {
	BtnChannel string `json:"btn_channel,omitempty"`
	BackURL    string `json:"back_url,omitempty"`
	BackName   string `json:"back_name,omitempty"`
	Color      string `json:"color,omitempty"`
	NoClose    bool   `json:"no_close,omitempty"`
	Passed     bool   `json:"passed,omitempty"`
	DirectBack bool   `json:"direct_back,omitempty"`
	BtnSize    int64  `json:"btn_size,omitempty"` // 1 优选按钮 2 备选按钮
}

type AiDeeplinkReq struct {
	OriginLink string `form:"origin_link" validate:"required"`
}

type AiDeeplink struct {
	Deeplink string `json:"deeplink"`
}

type AiDeeplinkGroupRsp struct {
	AbId string `json:"ab_id"`
	Ext  string `json:"ext"`
}

type AiDeeplinkMaterial struct {
	InnerId    string
	InnerType  int64
	SourceName string
	AbId       string
	AccountId  string
}
