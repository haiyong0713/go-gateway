package bcg

const (
	PlatformAndroid   = "android"
	PlatformIDAndroid = 1
	PlatformIOS       = "iphone"
	PlatformIDIOS     = 2
	PlatformWeb       = "web"
	PlatformIDWeb     = 3

	GoodsLocTypeDesc = 1
	GoodsLocTypeCard = 2
)

func TranPlatformID(platform string) int {
	switch platform {
	case PlatformAndroid:
		return PlatformIDAndroid
	case PlatformIOS:
		return PlatformIDIOS
	case PlatformWeb:
		return PlatformIDWeb
	default:
		return 0
	}
}

type GoodsCtx struct {
	AdExtra    string `json:"ad_extra"`
	Build      int64  `json:"build"`
	Buvid      string `json:"buvid"`
	City       string `json:"city"`
	Country    string `json:"country"`
	IP         string `json:"ip"`
	Network    string `json:"network"`
	PlatformID int    `json:"platform_id"`
	Province   string `json:"province"`
}

type GoodsParams struct {
	Uid         int64  `json:"uid"`
	UpUid       int64  `json:"up_uid"`
	DynamicID   int64  `json:"dynamic_id"`
	Ctx         string `json:"ctx"`
	InputExtend string `json:"InputExtend"`
}

type GoodsRes struct {
	OutputExtend *GoodsOutput `json:"output_extend"`
}

type GoodsOutput struct {
	List []*GoodsItem `json:"list"`
}

type GoodsItem struct {
	ItemsID           int64    `json:"itemsId"`
	ItemIdStr         string   `json:"itemsIdStr"`
	Name              string   `json:"name"`
	Brief             string   `json:"brief"`
	Img               string   `json:"img"`
	Price             float32  `json:"price"`
	PriceStr          string   `json:"priceStr"`
	IconName          string   `json:"iconName"`
	IconURL           string   `json:"iconUrl"`
	JumpLink          string   `json:"jumpLink"`
	JumpLinkDesc      string   `json:"jumpLinkDesc"`
	WordJumpLinkDesc  string   `json:"wordJumpLinkDesc"`
	Type              int      `json:"type"` // 来源 1-淘宝、2-会员购  3-京东
	SourceType        int      `json:"sourceType"`
	AdMark            string   `json:"adMark"`
	SchemaURL         string   `json:"schemaUrl"`
	SchemaPackageName string   `json:"schemaPackageName"`
	UserAdWebV2       bool     `json:"userAdWebV2"`
	MSource           string   `json:"msource"`
	OpenWhiteList     []string `json:"openWhiteList"`
	ShopGoodType      int      `json:"shopGoodType"`
	AppName           string   `json:"appName"`
	OuterApp          int      `json:"outerApp"` // 是否外部app 1-外部  0-内部
	SourceDesc        string   `json:"sourceDesc"`
}

// 广告device特殊处理
func SetAdDevice(mobiApp, metaDevice string) string {
	if mobiApp == "iphone" && metaDevice == "pad" {
		return "ipad"
	} else {
		return mobiApp
	}
}
