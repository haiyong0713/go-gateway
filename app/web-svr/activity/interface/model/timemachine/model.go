package timemachine

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
)

//go:generate kratos t protoc --grpc timemachine.proto

var DefaultTidScores = []*TidScore{
	{Tname: "生活", Tid: 160, Score: 1},
	{Tname: "影视", Tid: 181, Score: 1},
	{Tname: "科技", Tid: 36, Score: 1},
	{Tname: "娱乐", Tid: 5, Score: 1},
	{Tname: "音乐", Tid: 3, Score: 1},
	{Tname: "游戏", Tid: 4, Score: 1},
}

type Result struct {
	Sid          int64         `json:"sid"`
	PageOne      *PageOne      `json:"page_one,omitempty"`
	PageTwo      *PageTwo      `json:"page_two,omitempty"`
	PageThree    *PageThree    `json:"page_three,omitempty"`
	PageFour     *PageFour     `json:"page_four,omitempty"`
	PageFive     *PageFive     `json:"page_five,omitempty"`
	PageSix      *PageSix      `json:"page_six,omitempty"`
	PageSeven    *PageSeven    `json:"page_seven,omitempty"`
	PageEight    *PageEight    `json:"page_eight,omitempty"`
	PageNine     *PageNine     `json:"page_nine,omitempty"`
	PageTen      *PageTen      `json:"page_ten,omitempty"`
	PageEleven   *PageEleven   `json:"page_eleven,omitempty"`
	PageTwelve   *PageTwelve   `json:"page_twelve,omitempty"`
	PageThirteen *PageThirteen `json:"page_thirteen,omitempty"`
	PageFourteen *PageFourteen `json:"page_fourteen"`
}

type PageOne struct {
	Mid              int64            `json:"mid"`
	Name             string           `json:"name"`
	Face             string           `json:"face"`
	VisitDays        int64            `json:"visit_days"`
	HourVisitDays    map[string]int64 `json:"hour_visit_days"`
	MaxVisitDaysHour int64            `json:"max_visit_days_hour"`
}

type PageTwo struct {
	Vv             int64       `json:"vv"`
	MaxVvTid       int32       `json:"max_vv_tid"`
	MaxVvTname     string      `json:"max_vv_tname"`
	Top6VvTidScore []*TidScore `json:"top6_vv_tid_score"`
}

type PageThree struct {
	MaxVvSubtid int32  `json:"max_vv_subtid"`
	Top10VvTag  string `json:"top10_vv_tag"`
	TagName     string `json:"tag_name"`
	TagDescOne  string `json:"tag_desc_one"`
	TagDescTwo  string `json:"tag_desc_two"`
	TagPic      string `json:"tag_pic"`
}

type PageFour struct {
	CoinTime  string `json:"coin_time"`
	CoinUsers int64  `json:"coin_users"`
	Arc       *Arc   `json:"arc"`
}

type PageFive struct {
	PlayBangumi    int64   `json:"play_bangumi"`
	BestLikeSeason *Season `json:"best_like_season"`
}

type PageSix struct {
	PlayMovies       int64   `json:"play_movies"`
	PlayDramas       int64   `json:"play_dramas"`
	PlayDocumentarys int64   `json:"play_documentarys"`
	PlayZongyi       int64   `json:"play_zongyi"`
	BestLikeYinshi   *Season `json:"best_like_yinshi"`
}

type PageSeven struct {
	ViewTime   string `json:"view_time"`
	EventID    int64  `json:"event_id"`
	EventTitle string `json:"event_title"`
	EventDesc  string `json:"event_desc"`
}

type PageEight struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Arc  *Arc   `json:"arc"`
}

type PageNine struct {
	Mid      int64  `json:"mid"`
	Name     string `json:"name"`
	Face     string `json:"face"`
	Duration int64  `json:"duration"`
}

type PageTen struct {
	CreateAvs   int64 `json:"create_avs"`
	CreateReads int64 `json:"create_reads"`
	AvVv        int64 `json:"av_vv"`
	ReadVv      int64 `json:"read_vv"`
	Type        int   `json:"type"`
	Arc         *Arc  `json:"arc"`
}

type PageEleven struct {
	Mid       int64  `json:"mid"`
	Name      string `json:"name"`
	Face      string `json:"face"`
	BestFanVv int64  `json:"best_fan_vv"`
}

type PageTwelve struct {
	LiveDays         int64   `json:"live_days"`
	Ratio            float64 `json:"ratio"`
	MaxOnlineNumTime string  `json:"max_online_num_time"`
	MaxOnlineNum     int64   `json:"max_online_num"`
}

type PageThirteen struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type PageFourteen struct {
	RegionDesc string  `json:"region_desc"`
	Flags      []*Flag `json:"flags"`
}

type TidScore struct {
	Tid   int64  `json:"tid"`
	Tname string `json:"tname"`
	Score int64  `json:"score"`
}

type Arc struct {
	Aid      int64      `json:"aid"`
	Title    string     `json:"title"`
	Pic      string     `json:"pic"`
	Duration int64      `json:"duration"`
	Owner    api.Author `json:"owner,omitempty"`
}

type Season struct {
	SeasonID       int32  `json:"season_id"`
	Title          string `json:"title"`
	Cover          string `json:"cover"`
	SeasonType     int32  `json:"season_type"`
	SeasonTypeName string `json:"season_type_name"`
}

type Flag struct {
	Lid     int64  `json:"lid"`
	Message string `json:"message"`
}

type Tag struct {
	Name    string `json:"name"`
	DescOne string `json:"desc_one"`
	DescTwo string `json:"desc_two"`
	Pic     string `json:"pic"`
}

type TagScore struct {
	Tid   int64
	Score int64
}

type RegionDesc struct {
	Tid      int32  `json:"tid"`
	Name     string `json:"name"`
	DescOne  string `json:"desc_one"`
	DescTwo  string `json:"desc_two"`
	Pic      string `json:"pic"`
	FlagDesc string `json:"flag_desc"`
}

type Event struct {
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	PreTime string `json:"pre_time"`
}

type ResUserReport2020HourVisitDays struct {
	Name      string `json:"name"`
	VisitDays int64  `json:"visit_days"`
	Desc      string `json:"-"`
}
type ResUserReport2020Top6TidScore struct {
	Tid   int64  `json:"tid,omitempty"`
	TName string `json:"tname,omitempty"`
	Score int64  `json:"score,omitempty"`
}

type ResUserReport2020VideoInfo struct {
	Oid          int64  `json:"oid,omitempty"`
	Cid          int64  `json:"cid,omitempty"`
	Title        string `json:"title,omitempty"`
	Pic          string `json:"pic,omitempty"`
	Mid          int64  `json:"mid,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	Face         string `json:"face,omitempty"`
	PopularStart int64  `json:"popular_start,omitempty"`
	PopularEnd   int64  `json:"popular_end,omitempty"`
}

type ResUserReport2020 struct {
	User               *ResUserReport2020VideoInfo       `json:"user"`
	Identification     bool                              `json:"identification"`
	Silence            bool                              `json:"silence"`
	TelStatus          bool                              `json:"tel_status"`
	VisitDays          int64                             `json:"visit_days"`   // 用户使用B站的天数
	PlayVideos         int64                             `json:"play_videos"`  // 2020年度该uid累计观看的正常浏览状态视频（OGV+UGC）的个数
	PlayMinutes        int64                             `json:"play_minutes"` // 2020年度该uid累观看分钟数,时长超过365天会处理成365天
	PlayDesc           string                            `json:"play_desc"`
	HourVisitDays      []*ResUserReport2020HourVisitDays `json:"hour_visit_days"` // 6时段的访问天数 1:10,2:34,3:11,...6:45，如果没有则用6时段的播放天数替换
	VisitDesc          string                            `json:"visit_desc"`
	FrequentlyTime     string                            `json:"frequently_time"`
	FavType            int64                             `json:"fav_type"` // 1为tag，2为分区
	FavTag             string                            `json:"fav_tag"`  // 最常看的视频类型（tag/二级分区）
	FavTagDesc         string                            `json:"fav_tag_desc"`
	FavTagPic          string                            `json:"fav_tag_pic"`
	Top6TidScore       []*ResUserReport2020Top6TidScore  `json:"top6_tid_score"`   // tid1:score,tid2:score...tid6:score
	IsShowP4           int64                             `json:"is_show_p4"`       // 0为不展示，1展示
	LatestPlayTime     int64                             `json:"latest_play_time"` // 最晚观看时间
	LatestPlayVideo    *ResUserReport2020VideoInfo       `json:"latest_play_video"`
	IsShowP5           int64                             `json:"is_show_p5"`         // 0为不展示，1展示
	LongestPlayDay     string                            `json:"longest_play_day"`   // 年内单日播放视频时间>=3小时且播放时长最长的那天
	LongestPlayHours   int64                             `json:"longest_play_hours"` // 最长单日播放视频小时数，超过24小时会处理为24小时
	LongestPlayTag     []string                          `json:"longest_play_tag"`   // tag多与XX，XX和XX有关,多个用，分割,数据格式为rank:tag,rank:tag,rank:tag
	LongestPlayDesc    string                            `json:"longest_play_desc"`
	LongestPlayTagDesc string                            `json:"longest_play_tag_desc"`
	LongestPlayTagImg  string                            `json:"longest_play_tag_img"`
	IsShowP6           int64                             `json:"is_show_p6"` // 0为不展示，1展示
	MaxVv              int64                             `json:"max_vv"`     // 取年内累计播放单个正常浏览状态UGC视频中循环次数最多的次数
	MaxVvVideo         *ResUserReport2020VideoInfo       `json:"max_vv_video"`
	MaxVvDesc          string                            `json:"max_vv_desc"`
	IsShowP7           int64                             `json:"is_show_p7"` // 0为不展示，1展示
	SumLike            int64                             `json:"sum_like"`   // 今年总的点赞次数
	SumCoin            int64                             `json:"sum_coin"`   // 今年总的投币次数
	SumFav             int64                             `json:"sum_fav"`    // 今年总的收藏次数
	CoinTime           int64                             `json:"coin_time"`  // 投过币的且硬币量最高的视频的最早投币时间 eg:20201112
	CoinUsers          int64                             `json:"coin_users"` // 除了自己其他的投币用户数
	CoinVideo          *ResUserReport2020VideoInfo       `json:"coin_video"`
	IsShowP8           int64                             `json:"is_show_p8"`       // 0为不展示，1展示
	RecommendVideo     []*ResUserReport2020VideoInfo     `json:"recommend_videos"` // 1:avid1,2:avid2,3:avid3,....6:avid6
	IsShowP9           int64                             `json:"is_show_p9"`       // 0为不展示，1展示
	FavUpType          int64                             `json:"fav_up_type"`      // 0:视频up主，1:专栏up主
	FavUpVv            int64                             `json:"fav_up_vv"`        // 观看最喜欢的up主次数
	FavUpInfo          *ResUserReport2020VideoInfo       `json:"fav_up_info"`
	IsShowP10          int64                             `json:"is_show_p10"`      // 0为不展示，1展示
	CreateAvs          int64                             `json:"create_avs"`       // 20年投稿视频数
	CreateReads        int64                             `json:"create_reads"`     // 20年投稿专栏数
	AvVv               int64                             `json:"av_vv"`            // up主所有视频稿件在20年的播放量
	ReadVv             int64                             `json:"read_vv"`          // up主所有专栏在20年的阅读量
	BestCreateType     int64                             `json:"best_create_type"` // 0:视频，1:专栏
	BestCreateInfo     *ResUserReport2020VideoInfo       `json:"best_create_info"`
	IsShowP11          int64                             `json:"is_show_p11"`      // 0为不展示，1展示
	PlayComic          int64                             `json:"play_comic"`       // 播放过的动画数量
	PlayMovie          int64                             `json:"play_movie"`       // 播放过的电影数量
	PlayDrama          int64                             `json:"play_drama"`       // 播放电视剧数量
	PlayDocumentary    int64                             `json:"play_documentary"` // 播放纪录片数量
	PlayVariety        int64                             `json:"play_variety"`     // 播放综艺的数量
	FavSeasonID        int32                             `json:"fav_season_id"`    // 最喜欢看的ogv的seasonid
	FavSeasonType      string                            `json:"fav_season_type"`  // 最喜欢看的ogv的分类
	FavSeasonInfo      *ResUserReport2020VideoInfo       `json:"fav_season_info"`
	IsShowP12          int64                             `json:"is_show_p12"`  // 0为不展示，1展示
	VipDays            int64                             `json:"vip_days"`     // 成为大会员的天数
	VipAvCount         int64                             `json:"vip_av_count"` // 观看大会员ogv部数
	VipAvPlay          int64                             `json:"vip_av_play"`  // 观看大会员专属内容分钟数
	VipDesc            string                            `json:"vip_desc"`
	IsShowP13          int64                             `json:"is_show_p13"`         // 0为不展示，1展示
	LiveHours          int64                             `json:"live_hours"`          // 观看直播的小时数
	LiveBeyondPercent  int64                             `json:"live_beyond_percent"` // 观看时长超过百分之多少的观众 eg:99%
	FavLiveUp          *ResUserReport2020VideoInfo       `json:"fav_live_up"`         // 最喜欢观看的主播
	FavLiveUpPlay      int64                             `json:"fav_live_up_play"`    // 观看该主播时长（分钟）
	LotteryEnd         bool                              `json:"lottery_end"`
	AID                int64                             `json:"aid"`
	LotteryID          string                            `json:"lottery_id"`
	PublishStatus      int                               `json:"publish_status"` // 0:可投稿，1:冷却中，1小时内不能重复投稿，2:已投稿
	Ctime              xtime.Time                        `json:"ctime"`          // 资源创建时间
	Mtime              xtime.Time                        `json:"mtime"`          // 资源修改时间
}
