package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/api"
	"go-gateway/app/app-svr/app-dynamic/interface/model/pgc"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatIPhone is int8 for iphone.
	PlatIPhone = int8(1)
	// PlatIPad is int8 for ipad.
	PlatIPad = int8(2)
	// PlatWPhone is int8 for wphone.
	PlatWPhone = int8(3)
	// PlatAndroidG is int8 for Android Global.
	PlatAndroidG = int8(4)
	// PlatIPhoneI is int8 for Iphone Global.
	PlatIPhoneI = int8(5)
	// PlatIPadI is int8 for IPAD Global.
	PlatIPadI = int8(6)
	// PlatAndroidTV is int8 for AndroidTV Global.
	PlatAndroidTV = int8(7)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// PlatAndroidB is int8 for android_b
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for iphone_b
	PlatIPhoneB = int8(10)
	// PlatIPadHD is int8 for ipadHD.
	PlatIPadHD = int8(20)
	// PlatAndroidHD is int8 for android_hd
	PlatAndroidHD = int8(90)

	GotoAv                     = "av"
	GotoSpaceDyn               = "space_dyn" // 空间动态tab
	GotoLBS                    = "lbs"
	GotoDyn                    = "dynamic"
	GotoLive                   = "live"
	GotoTag                    = "tag"
	GotoChannel                = "channel"
	GotoActivity               = "activity"
	GotoTopic                  = "topic"
	GotoArticle                = "article"
	GotoClip                   = "clip"
	GOtoMedialist              = "medialist"
	GotoURL                    = "url"
	GotoSpace                  = "space" // 空间
	GotoFeedSchool             = "feed_school"
	GotoChannelSearch          = "channel_search"    // 频道垂搜
	GotoTopicSearch            = "topic_search"      // 话题垂搜
	GotoOfficialAccount        = "official_account"  // 校园 - 官方账号
	GotoOfficialDynamic        = "official_dynamic"  // 校园 - 入校必看
	GotoSchoolBillboard        = "school_billboard"  // 校园 - 校园十大
	GotoSchoolTopicHome        = "school_topic_home" // 校园 - 话题讨论
	GotoSchoolTopicList        = "school_topic_list" // 校园话题列表
	GotoStory                  = "story"
	GotoDynPublishWithNewTopic = "dyn_publish_with_new_topic"

	// bvid开关
	BvOpen       = 1
	InvalidTitle = "内容已失效"

	// Icon
	IconShare = int32(1)
	IconReply = int32(2)
	IconLike  = int32(3)

	// ModuleType
	ModuleTypeAuthor = "author"
	ModuleTypePlayer = "player"
	ModuleTypeDesc   = "desc"
	ModuleTypeStat   = "stat"

	// CardType
	CardTypeAv = "av"

	// teenagers
	TeenagersClose = 0
	TeenagersOpen  = 1

	// badge type
	BgStyleFill              = 1
	BgStyleStroke            = 2
	BgStyleFillAndStroke     = 3
	BgStyleNoFillAndNoStroke = 4

	// 频道类型
	OldChanne  = 1
	NewChannel = 2

	// 投票跳转地址
	VoteURI = "https://t.bilibili.com/vote/h5/index/#/result?vote_id=%v&dynamic_id=%v"
	LottURI = "https://t.bilibili.com/lottery/h5/index/#/result?business_id=%d&business_type=%d&lottery_id=%v&dynamic_id=%v"
	LBSURI  = "bilibili://following/dynamic_location?poi=%v&type=%v&lat=%v&lng=%v&title=%v&address=%v"

	// abtest
	_abMiss        = "miss"
	_abUnloginFeed = "dt_unlogin_region_ups"
	_abDynAll      = "dyn_tab_all"
	_abDynVideo    = "dyn_tab_video"

	AbDynVdTabPureNew = "dt_filter_v2_new"
	AbDynVdTabOld     = "dt_filter_v2_old"
	AbDynVdTabOld2New = "dt_filter_v2_oldtonew"

	// 动态详情页底栏样式升级
	AbDynDetailBar = "dt_detail_page_bottom_bar_sty"
	// 综合页话题广场样式实验
	AbDynAllTopicSquare = "topic_square_new_style2"
)

var (
	LBSHandler = func(poi string, poiType int64, lat, lng float64, title, address string) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%v?poi=%v&type=%v&lat=%v&lng=%v&title=%v&address=%v", uri, poi, poiType, lat, lng, title, address)
		}
	}
	ChannelHandler = func(tab string) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?%s", uri, tab)
		}
	}
	SuffixHandler = func(name string) func(uri string) string {
		return func(uri string) string {
			if !strings.Contains(uri, "?") {
				uri = fmt.Sprintf("%s?%s", uri, name)
			} else {
				uri = fmt.Sprintf("%s&%s", uri, name)
			}
			return uri
		}
	}
	// URI里面拼接动态ID
	DynamicIDHandler = func(dynamicID int64) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("DynamicIDHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("DynamicIDHandler url.ParseQuery error(%v)", err)
				return uri
			}
			if dynamicID > 0 {
				params.Set("dynamic_id", strconv.FormatInt(dynamicID, 10))
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	// abtest
	UnloginAbtestFlag  = ab.String(_abUnloginFeed, "initTag", _abMiss)
	TabAbtestAllflag   = ab.String(_abDynAll, "initTag", _abMiss)
	TabAbtestVedioflag = ab.String(_abDynVideo, "initTag", _abMiss)

	DynVideoTabPureNew = ab.String(AbDynVdTabPureNew, "动态筛选器二期分流-新用户", "1")
	DynVideoTabOld     = ab.String(AbDynVdTabOld, "动态筛选器二期分流-老用户", "1")
	DynVideoTabOld2New = ab.String(AbDynVdTabOld2New, "动态筛选器二期分流-老转新用户", "1")

	DynDetailBar      = ab.String(AbDynDetailBar, "【动态详情页&图文进公域】底栏升级", "0") // 0 对照组
	DynAllTopicSquare = ab.String(AbDynAllTopicSquare, "综合页话题广场新样式实验", "1") // 1对照组
)

// FillURI deal app schema.
func FillURI(gt, param string, f func(uri string) string) (uri string) {
	if param == "" {
		return
	}
	switch gt {
	case GotoAv, "":
		uri = "bilibili://video/" + param
	case GotoSpaceDyn:
		uri = "bilibili://space/" + param + "?defaultTab=dynamic"
	case GotoLBS:
		uri = "bilibili://following/dynamic_location"
	case GotoDyn:
		uri = "bilibili://following/detail/" + param
	case GotoLive:
		uri = "bilibili://live/" + param
	case GotoTag:
		uri = "bilibili://pegasus/channel/" + param
	case GotoChannel:
		uri = "bilibili://pegasus/channel/v2/" + param
	case GotoActivity:
		uri = "bilibili://following/activity_landing/" + param
	case GotoTopic:
		uri = "bilibili://pegasus/channel/" + param + "/"
	case GotoArticle:
		uri = "bilibili://article/" + param + "?from=5"
	case GotoClip:
		uri = "bilibili://clip/detail/" + param + "/0"
	case GOtoMedialist:
		uri = "bilibili://music/playlist/playpage/" + param + "?from=dt_playlist"
	case GotoSpace:
		uri = "bilibili://space/" + param
	case GotoFeedSchool:
		uri = "bilibili://campus/moment/" + param
	case GotoChannelSearch:
		uri = "bilibili://pegasus/channel/search?query=" + param
	case GotoTopicSearch:
		uri = "activity://following/topic_search?search_name=" + param + "&only_search=false&hotTopic=true"
	case GotoOfficialDynamic:
		uri = "bilibili://campus/read/" + param
	case GotoOfficialAccount:
		uri = "bilibili://campus/official/" + param
	case GotoSchoolBillboard:
		uri = "bilibili://campus/billboard/" + param
	case GotoSchoolTopicHome:
		uri = "bilibili://campus/topic_home/" + param
	case GotoSchoolTopicList:
		uri = "bilibili://campus/topic/" + param
	case GotoStory:
		uri = "bilibili://story/" + param
	case GotoURL:
		uri = param
	case GotoDynPublishWithNewTopic:
		uri = "bilibili://following/publish?" + param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

// a wrapper for url.QueryEscape but replace "+" with "%20"
func QueryEscape(pram string) string {
	return strings.Replace(url.QueryEscape(pram), "+", "%20", -1)
}

var (
	AvPlayHandlerGRPC = func(a *arcgrpc.Arc, ap *arcgrpc.BvcVideoItem, his *arcgrpc.History) func(uri string) string {
		var player string
		if ap != nil {
			bs, _ := json.Marshal(ap)
			player = url.QueryEscape(string(bs))
			if strings.IndexByte(player, '+') > -1 {
				player = strings.Replace(player, "+", "%20", -1)
			}
		}
		return func(uri string) string {
			var uriStr string
			if player != "" && (a.Dimension.Height != 0 || a.Dimension.Width != 0) {
				uriStr = fmt.Sprintf("%s?page=1&player_preload=%s&player_width=%d&player_height=%d&player_rotate=%d", uri, player, a.Dimension.Width, a.Dimension.Height, a.Dimension.Rotate)
			} else if player != "" {
				uriStr = fmt.Sprintf("%s?page=1&player_preload=%s", uri, player)
			} else if a.Dimension.Height != 0 || a.Dimension.Width != 0 {
				uriStr = fmt.Sprintf("%s?player_width=%d&player_height=%d&player_rotate=%d", uri, a.Dimension.Width, a.Dimension.Height, a.Dimension.Rotate)
			}
			if his != nil && his.Cid == a.FirstCid { //由于秒开目前仅有第一p，所以播放进度目前只有在第1p时返回
				if uriStr == "" {
					uriStr = fmt.Sprintf("%s?history_progress=%d", uri, his.Progress)
				} else {
					uriStr = fmt.Sprintf("%s&history_progress=%d", uriStr, his.Progress)
				}
			}
			if uriStr != "" {
				uri = uriStr
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			// 拜年祭活动合集
			if a.AttrValV2(arcgrpc.AttrBitV2ActSeason) == arcgrpc.AttrYes && a.SeasonTheme != nil {
				params.Set("is_festival", "1")
				params.Set("bg_color", a.SeasonTheme.BgColor)
				params.Set("selected_bg_color", a.SeasonTheme.SelectedBgColor)
				params.Set("text_color", a.SeasonTheme.TextColor)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	AvPlayHandlerGRPCV2 = func(ap *arcgrpc.ArcPlayer, cid int64, showPlayerURL bool) func(uri string) string {
		var (
			a                               = ap.Arc
			playerInfo                      *arcgrpc.PlayerInfo
			ok                              bool
			player                          string
			height, width, rotate, progress int64
		)
		if playerInfo, ok = ap.PlayerInfo[cid]; !ok {
			if playerInfo, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; !ok {
				playerInfo = ap.PlayerInfo[a.FirstCid]
			}
		}
		if playerInfo != nil {
			// 秒开部分
			if playerInfo.Playurl != nil && showPlayerURL {
				bs, _ := json.Marshal(playerInfo.Playurl)
				player = string(bs)
			}
			// 扩展部分
			if playerInfo.PlayerExtra != nil {
				cid = playerInfo.PlayerExtra.Cid
				// 宽高
				if playerInfo.PlayerExtra.Dimension != nil {
					height = playerInfo.PlayerExtra.Dimension.Height
					width = playerInfo.PlayerExtra.Dimension.Width
					rotate = playerInfo.PlayerExtra.Dimension.Rotate
				}
				// 历史
				progress = playerInfo.PlayerExtra.Progress
			}
		}
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			// 秒开处理
			if player != "" {
				params.Set("player_preload", player)
			}
			if cid > 0 {
				params.Set("cid", strconv.FormatInt(cid, 10))
			}
			params.Set("history_progress", strconv.FormatInt(progress, 10))
			if height != 0 || width != 0 {
				params.Set("player_height", strconv.FormatInt(height, 10))
				params.Set("player_width", strconv.FormatInt(width, 10))
				params.Set("player_rotate", strconv.FormatInt(rotate, 10))
			}
			// 拜年祭活动合集
			if a.AttrValV2(arcgrpc.AttrBitV2ActSeason) == arcgrpc.AttrYes && a.SeasonTheme != nil {
				params.Set("is_festival", "1")
				params.Set("bg_color", a.SeasonTheme.BgColor)
				params.Set("selected_bg_color", a.SeasonTheme.SelectedBgColor)
				params.Set("text_color", a.SeasonTheme.TextColor)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}

	BatchPlayHandler = func(batch *pgc.PGCBatch) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			// 秒开处理
			if batch.InlineVideo.Url != "" {
				bs, _ := json.Marshal(batch.InlineVideo)
				params.Set("player_preload", string(bs))
			}
			if batch.InlineVideo.Cid > 0 {
				params.Set("cid", strconv.FormatInt(batch.InlineVideo.Cid, 10))
			}
			if batch.ID > 0 {
				params.Set("season_id", strconv.Itoa(int(batch.SeasonID)))
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}
	SeasonPlayHandler = func(batch *pgc.PGCSeason) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			// 秒开处理
			if batch.InlineVideo.Url != "" {
				bs, _ := json.Marshal(batch.InlineVideo)
				params.Set("player_preload", string(bs))
			}
			if batch.InlineVideo.Cid > 0 {
				params.Set("cid", strconv.FormatInt(batch.InlineVideo.Cid, 10))
			}
			if batch.InlineVideo.Cid > 0 {
				params.Set("aid", strconv.FormatInt(batch.InlineVideo.Aid, 10))
			}
			if batch.InlineVideo.Cid > 0 {
				params.Set("ep_id", strconv.FormatInt(batch.InlineVideo.Epid, 10))
			}
			if batch.InlineVideo.Cid > 0 {
				params.Set("duration", strconv.FormatInt(batch.InlineVideo.Duration, 10))
			}
			if batch.ID > 0 {
				params.Set("season_id", strconv.Itoa(batch.ID))
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}
	EncodePlusAs20 = func(param string) string {
		// 重新encode的时候空格变成了+号问题修复
		if strings.IndexByte(param, '+') > -1 {
			return strings.Replace(param, "+", "%20", -1)
		}
		return param
	}
	StoryHandler = func(ap *arcgrpc.ArcPlayer, cid int64, isIOS bool, scene string, vmid, offset int64, storyParam string) func(uri string) string {
		type playerArg struct {
			Duration int64  `json:"duration"`
			Aid      int64  `json:"aid"`
			Type     string `json:"type"`
			Cid      int64  `json:"cid"`
		}
		// 截止双端6.84版本 对于story的秒开处理存在差异
		// 这里是一个兼容结构保证双端从动态跳转到story播放器时可以正常秒开
		type storyItem struct {
			Uri        string            `json:"uri"`
			FfCover    string            `json:"ff_cover"`
			PlayerArgs playerArg         `json:"player_args"`
			Dimension  arcgrpc.Dimension `json:"dimension"`
		}

		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			params.Set("scene", scene)
			params.Set("vmid", strconv.FormatInt(vmid, 10))
			params.Set("offset", strconv.FormatInt(offset, 10))
			if len(storyParam) > 0 {
				params.Set("story_param", storyParam)
			}
			u.RawQuery = EncodePlusAs20(params.Encode())
			uri = u.String()

			// 兼容双端story秒开问题
			// TODO: 端上改好之后删掉这块逻辑
			item := storyItem{
				Uri:     uri,
				FfCover: ap.GetArc().GetFirstFrame(),
				PlayerArgs: playerArg{
					Aid:      ap.GetArc().GetAid(),
					Cid:      cid,
					Type:     "av",
					Duration: ap.GetArc().GetDuration(),
				},
				Dimension: ap.GetArc().GetDimension(),
			}
			itemMarshaled, _ := json.Marshal(item)
			// 双端这个参数名称都不一样
			if isIOS {
				params.Set("original_json", string(itemMarshaled))
			} else {
				params.Set("story_item", string(itemMarshaled))
			}
			u.RawQuery = EncodePlusAs20(params.Encode())
			return u.String()
		}
	}
)

// nolint:gomnd
func StatString(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// DurationString duration to string
func DurationString(second int64) (s string) {
	var hour, min, sec int
	if second < 1 {
		return
	}
	d, err := time.ParseDuration(strconv.FormatInt(second, 10) + "s")
	if err != nil {
		log.Error("%v", err)
		return
	}
	r := strings.NewReplacer("h", ":", "m", ":", "s", ":")
	ts := strings.Split(strings.TrimSuffix(r.Replace(d.String()), ":"), ":")
	// nolint:gomnd
	if len(ts) == 1 {
		sec, _ = strconv.Atoi(ts[0])
	} else if len(ts) == 2 {
		min, _ = strconv.Atoi(ts[0])
		sec, _ = strconv.Atoi(ts[1])
	} else if len(ts) == 3 {
		hour, _ = strconv.Atoi(ts[0])
		min, _ = strconv.Atoi(ts[1])
		sec, _ = strconv.Atoi(ts[2])
	}
	if hour == 0 {
		s = fmt.Sprintf("%d:%02d", min, sec)
		return
	}
	s = fmt.Sprintf("%d:%02d:%02d", hour, min, sec)
	return
}

const (
	PerHour   = 3600
	PerMinute = 60
)

func VideoDuration(du int64) string {
	hour := du / PerHour
	du = du % PerHour
	minute := du / PerMinute
	second := du % PerMinute
	if hour != 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
	}
	return fmt.Sprintf("%02d:%02d", minute, second)
}

func FillBadge(title string, bgColor, bgColorNight, borderColor, borderColorNight string, bType int32) *api.VideoBadge {
	badge := &api.VideoBadge{
		Text:             title,
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BA833F",
		BorderColor:      "#FAAB4B",
		BorderColorNight: "#BA833F",
		BgStyle:          bType,
	}
	if bgColor != "" {
		badge.BgColor = bgColor
	}
	if bgColorNight != "" {
		badge.BgColorNight = bgColorNight
	}
	if borderColor != "" {
		badge.BorderColor = borderColor
	}
	if borderColorNight != "" {
		badge.BorderColorNight = borderColorNight
	}
	return badge
}

func FormMatchTime(stime int64) (label string) {
	ls := time.Unix(stime, 0)
	lt := time.Now()
	if lt.Year() == ls.Year() {
		if lt.YearDay()-ls.YearDay() == 1 {
			label = fmt.Sprintf("昨天 %v", ls.Format("15:04"))
			return
		} else if lt.YearDay()-ls.YearDay() == 0 {
			label = fmt.Sprintf("今天 %v", ls.Format("15:04"))
			return
		} else if lt.YearDay()-ls.YearDay() == -1 {
			label = fmt.Sprintf("明天 %v", ls.Format("15:04"))
			return
		} else {
			label = ls.Format("01-02 15:04")
		}
	} else {
		label = ls.Format("01-02 15:04")
	}
	return
}

func FillReplyURL(uri string, surfix string) string {
	if !strings.Contains(uri, "?") {
		return fmt.Sprintf("%s?%s", uri, surfix)
	} else {
		return fmt.Sprintf("%s&%s", uri, surfix)
	}
}

// UpPubDataString is.
func UpPubDataString(t time.Time) string {
	now := time.Now()
	if now.Year() == t.Year() {
		if now.Month() == t.Month() && now.Day() == t.Day() {
			return "今天 " + t.Format("15:04")
		}
		if now.Month() == t.Month() && now.Day() < t.Day() && (t.Day()-now.Day() == 1) {
			return "明天 " + t.Format("15:04")
		}
		return t.Format("01-02 15:04")
	}
	return t.Format("2006-01-02 15:04")
}

// UpPubShareDataString is.
func UpPubShareDataString(t time.Time) string {
	now := time.Now()
	if now.Year() == t.Year() {
		return t.Format("01月02日 15:04")
	}
	return t.Format("2006年01月02日 15:04")
}

// nolint:gomnd
func UpStatString(number int64, suffix string) (s string) {
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// IsPremiereBefore 稿件首映前
func IsPremiereBefore(a *arcgrpc.Arc) bool {
	const premiereState = -40
	return (a != nil && a.Premiere != nil && a.Premiere.State == arcgrpc.PremiereState_premiere_before) && a.State == premiereState
}

// Plat return plat by platStr or mobiApp
func Plat(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
		return PlatIPhone
	case "white":
		return PlatIPhone
	case "ipad":
		return PlatIPadHD
	case "android":
		return PlatAndroid
	case "android_b":
		return PlatAndroidB
	case "win":
		return PlatWPhone
	case "android_G":
		return PlatAndroidG
	case "android_i":
		return PlatAndroidI
	case "iphone_i":
		if device == "pad" {
			return PlatIPadI
		}
		return PlatIPhoneI
	case "ipad_i":
		return PlatIPadI
	case "android_tv":
		return PlatAndroidTV
	case "iphone_b":
		return PlatIPhoneB
	case "android_hd":
		return PlatAndroidHD
	}
	return PlatIPhone
}
