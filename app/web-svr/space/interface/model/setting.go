package model

// DefaultPrivacy default privacy.
var (
	IndexOrderAppointment    = 26
	IndexOrderCoinVideo      = 5
	IndexOrderLikeVideo      = 10
	IndexOrderOfficialEvents = 24
	OrderItemCoinVideo       = &IndexOrder{ID: IndexOrderCoinVideo, Name: "最近投币的视频"}
	OrderItemLikeVideo       = &IndexOrder{ID: IndexOrderLikeVideo, Name: "最近点赞的视频"}

	PcyBangumi        = "bangumi"
	PcyTag            = "tags"
	PcyFavVideo       = "fav_video"
	PcyCoinVideo      = "coins_video"
	PcyGroup          = "groups"
	PcyGame           = "played_game"
	PcyChannel        = "channel"
	PcyUserInfo       = "user_info"
	PcyLikeVideo      = "likes_video"
	PcyBbq            = "bbq"
	PcyComic          = "comic"
	PcyDressUp        = "dress_up"
	LivePlayback      = "live_playback"
	DefaultIndexOrder = []*IndexOrder{
		{ID: 1, Name: "我的稿件"},
		{ID: 8, Name: "我的专栏"},
		{ID: 7, Name: "我的视频列表"},
		{ID: 2, Name: "我的收藏夹"},
		{ID: 3, Name: "订阅番剧"},
		{ID: 4, Name: "订阅标签"},
		OrderItemCoinVideo,
		OrderItemLikeVideo,
		{ID: 6, Name: "我的圈子"},
		{ID: 9, Name: "我的相簿"},
		{ID: IndexOrderAppointment, Name: "预约"},
		{ID: 21, Name: "公告"},
		{ID: 22, Name: "直播间"},
		{ID: 23, Name: "个人资料"},
		//{ID: 24, Name: "官方活动"},
		{ID: 25, Name: "最近玩过的游戏"},
	}
	IndexOrderMap = indexOrderMap()
)

// Pcy not in default privacy.
var (
	PcyDisableFollowing  = "disable_following"
	PcyCloseSpaceMedal   = "close_space_medal"
	PcyOnlyShowWearing   = "only_show_wearing"
	PcyDisableShowSchool = "disable_show_school"
	PcyDisableShowNft    = "disable_show_nft"
)

// Setting setting struct.
type Setting struct {
	Privacy       map[string]int `json:"privacy"`
	ShowNftSwitch bool           `json:"show_nft_switch"`
	IndexOrder    []*IndexOrder  `json:"index_order"`
}

type AppSetting struct {
	Privacy       map[string]int `json:"privacy"`
	ShowNftSwitch bool           `json:"show_nft_switch"`
	ExclusiveURL  string         `json:"exclusive_url"` //空间专属页跳转地址
}

// Privacy privacy struct.
type Privacy struct {
	Privacy string `json:"privacy"`
	Status  int    `json:"status"`
}

// IndexOrder index order struct.
type IndexOrder struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func indexOrderMap() map[int]string {
	data := make(map[int]string, len(DefaultIndexOrder))
	for _, v := range DefaultIndexOrder {
		data[v.ID] = v.Name
	}
	return data
}
