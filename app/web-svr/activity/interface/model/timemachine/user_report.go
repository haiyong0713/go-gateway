package timemachine

import (
	xtime "go-common/library/time"
)

// bcurd -dsn='main_lottery:RWXcNcw1X43S1K1nvB51x7iNjMMjP0ba@tcp(10.221.34.182:4000)/main_lottery?parseTime=true'  -schema=main_lottery -table=ads_user_year_report_2020_1y_y_20201127 -tmpl=bilibili_log.tmpl > user_report.go

// UserYearReport2020 represents a row from 'ads_user_year_report_2020_1y_y_20201127'.
type UserYearReport2020 struct {
	Mid                  int64      `json:"mid"`                     // 2020年有播放行为的用户id
	VisitDays            int64      `json:"visit_days"`              // 用户使用B站的天数
	PlayVideos           int64      `json:"play_videos"`             // 2020年度该uid累计观看的正常浏览状态视频（OGV+UGC）的个数
	PlayMinutesReal      int64      `json:"play_minutes_real"`       // 2020年度该uid累观看分钟数（取整，超过30秒算1分钟进位）,真实观看时长
	PlayMinutes          int64      `json:"play_minutes"`            // 2020年度该uid累观看分钟数,时长超过365天会处理成365天
	HourVisitDays        string     `json:"hour_visit_days"`         // 6时段的访问天数 1:10,2:34,3:11,...6:45，如果没有则用6时段的播放天数替换
	FavType              int64      `json:"fav_type"`                // 1为tag，2为分区
	FavTag               string     `json:"fav_tag"`                 // 最常看的视频类型（tag/二级分区）
	Top6TidScore         string     `json:"top6_tid_score"`          // tid1:score,tid2:score...tid6:score
	IsShowP4             int64      `json:"is_show_p4"`              // 0为不展示，1展示
	LatestPlayTime       int64      `json:"latest_play_time"`        // 最晚观看时间
	LatestPlayUp         int64      `json:"latest_play_up"`          // 最晚观看视频所属up的mid
	LatestPlayAvid       int64      `json:"latest_play_avid"`        // 最晚观看视频avid
	IsShowP5             int64      `json:"is_show_p5"`              // 0为不展示，1展示
	LongestPlayDay       string     `json:"longest_play_day"`        // 年内单日播放视频时间>=3小时且播放时长最长的那天
	LongestPlayHoursReal int64      `json:"longest_play_hours_real"` // 最长单日播放视频小时数,真实时长
	LongestPlayHours     int64      `json:"longest_play_hours"`      // 最长单日播放视频小时数，超过24小时会处理为24小时
	LongestPlayTag       string     `json:"longest_play_tag"`        // tag多与XX，XX和XX有关,多个用，分割,数据格式为rank:tag,rank:tag,rank:tag
	LongestPlaySubtid    int64      `json:"longest_play_subtid"`     // 播放最多的二级分区id，-1为3个都为tag
	IsShowP6             int64      `json:"is_show_p6"`              // 0为不展示，1展示
	MaxVv                int64      `json:"max_vv"`                  // 取年内累计播放单个正常浏览状态UGC视频中循环次数最多的次数
	MaxVvUp              int64      `json:"max_vv_up"`               // 播放次数最多的视频的up主mid
	MaxVvAvid            int64      `json:"max_vv_avid"`             // 播放次数最多视频的avid
	IsShowP7             int64      `json:"is_show_p7"`              // 0为不展示，1展示
	SumLike              int64      `json:"sum_like"`                // 今年总的点赞次数
	SumCoin              int64      `json:"sum_coin"`                // 今年总的投币次数
	SumFav               int64      `json:"sum_fav"`                 // 今年总的收藏次数
	CoinTime             int64      `json:"coin_time"`               // 投过币的且硬币量最高的视频的最早投币时间 eg:20201112
	CoinUsers            int64      `json:"coin_users"`              // 除了自己其他的投币用户数
	CoinAvid             int64      `json:"coin_avid"`               // 投币视频avid
	IsShowP8             int64      `json:"is_show_p8"`              // 0为不展示，1展示
	RecommandAvid        string     `json:"recommand_avid"`          // 1:avid1,2:avid2,3:avid3,....6:avid6
	IsShowP9             int64      `json:"is_show_p9"`              // 0为不展示，1展示
	FavUpType            int64      `json:"fav_up_type"`             // 0:视频up主，1:专栏up主
	FavUp                int64      `json:"fav_up"`                  // 最喜爱的up主mid
	FavUpVv              int64      `json:"fav_up_vv"`               // 观看最喜欢的up主次数
	FavUpOid             int64      `json:"fav_up_oid"`              // 最喜欢的up主的代表作,avid或者专栏id
	IsShowP10            int64      `json:"is_show_p10"`             // 0为不展示，1展示
	CreateAvs            int64      `json:"create_avs"`              // 20年投稿视频数
	CreateReads          int64      `json:"create_reads"`            // 20年投稿专栏数
	AvVv                 int64      `json:"av_vv"`                   // up主所有视频稿件在20年的播放量
	ReadVv               int64      `json:"read_vv"`                 // up主所有专栏在20年的阅读量
	BestCreateType       int64      `json:"best_create_type"`        // 0:视频，1:专栏
	BestCreate           int64      `json:"best_create"`             // up主2019年的一个代表作，包含专栏和视频
	IsShowP11            int64      `json:"is_show_p11"`             // 0为不展示，1展示
	PlayComic            int64      `json:"play_comic"`              // 播放过的动画数量
	PlayMovie            int64      `json:"play_movie"`              // 播放过的电影数量
	PlayDrama            int64      `json:"play_drama"`              // 播放电视剧数量
	PlayDocumentary      int64      `json:"play_documentary"`        // 播放纪录片数量
	PlayVariety          int64      `json:"play_variety"`            // 播放综艺的数量
	FavSeasonID          int32      `json:"fav_season_id"`           // 最喜欢看的ogv的seasonid
	FavSeasonType        string     `json:"fav_season_type"`         // 最喜欢看的ogv的分类
	IsShowP12            int64      `json:"is_show_p12"`             // 0为不展示，1展示
	LiveHours            int64      `json:"live_hours"`              // 观看直播的小时数
	LiveBeyondPercent    int64      `json:"live_beyond_percent"`     // 观看时长超过百分之多少的观众 eg:99%
	FavLiveUp            int64      `json:"fav_live_up"`             // 最喜欢观看的主播
	FavLiveUpPlay        int64      `json:"fav_live_up_play"`        // 观看该主播时长（分钟）
	MaxVvHighlight       string     `json:"max_vv_highlight"`        // 最多观看次数视频高光时刻avid,cid,beginsecond,endsecond，'为视频不符合条件
	CoinHighlight        string     `json:"coin_highlight"`          // 投币最多视频高光时刻avid,cid,beginsecond,endsecond，'为视频不符合条件
	Ctime                xtime.Time `json:"ctime"`                   // 资源创建时间
	Mtime                xtime.Time `json:"mtime"`                   // 资源修改时间
	VipDays              int64      `json:"vip_days"`                // 成为大会员的天数
	VipAvCount           int64      `json:"vip_av_count"`            // 观看大会员ogv部数
	VipAvPlay            int64      `json:"vip_av_play"`             // 观看大会员专属内容分钟数
	IsShowP13            int64      `json:"is_show_p13"`             // 0为不展示，1展示
	LatestPlayHighlight  string     `json:"latest_play_highlight"`   // 最晚观看视频高光
	FavUpHighlight       string     `json:"fav_up_highlight"`        // 最喜欢up代表作高光时刻
}
