package model

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	GotoAv            = "av"
	GotoSpaceDyn      = "space_dyn" // 空间动态tab
	GotoDyn           = "dynamic"
	GotoLive          = "live"
	GotoArticle       = "article"
	GotoClip          = "clip"
	GotoURL           = "url"
	GotoSpace         = "space"          // 空间
	GotoChannelSearch = "channel_search" // 频道垂搜
	GotoTopicSearch   = "topic_search"   // 话题垂搜
	GotoStory         = "story"
	GotoWebAv         = "web_av"
	GotoWebSpace      = "web_space"

	// 投票跳转地址
	VoteURI = "https://t.bilibili.com/vote/h5/index/#/result?vote_id=%v&dynamic_id=%v"
	LottURI = "https://t.bilibili.com/lottery/h5/index/#/result?business_id=%d&business_type=%d&lottery_id=%v&dynamic_id=%v"
	LBSURI  = "bilibili://following/dynamic_location?poi=%v&type=%v&lat=%v&lng=%v&title=%v&address=%v"

	// 商品类型
	GoodsTypeTaoBao  = 1
	GoodsLocTypeCard = 2

	// 时间类型
	PerSecond = 1
	PerMinute = PerSecond * 60
	PerHour   = PerMinute * 60

	// desc解析正则
	EmojiRex = `[[][^\[\]]+[]]`
	MailRex  = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	WebRex   = `(http(s)?://)?([a-z0-9A-Z-]+\.)?(bilibili\.(com|tv|cn)|biligame\.(com|cn|net)|(bilibiliyoo|im9)\.com|biliapi\.net|b23\.tv|bili22\.cn|bili33\.cn|bili23\.cn|bili2233\.cn|(sugs\.suning\.com)|kaola\.com|bigfun\.cn|mcbbs\.net|mp\.weixin\.qq\.com|static\.cdsb\.com|bjnews\.com\.cn|720yun\.com|cctv\.com|jueze2021\.peopleapp\.com)($|/)([/.$*?~=#!%@&A-Za-z0-9_-]*)`
	AvRex    = `(AV|av|Av|aV)[0-9]+`
	BvRex    = `(BV|bv|Bv|bV)1[1-9A-NP-Za-km-z]{9}`
	CvRex    = `((CV|cv|Cv|cV)[0-9]+|(mobile/[0-9]+))`
	VcRex    = `(VC|vc|Vc|vC)[0-9]+`
	TopicRex = `#[^#@\r\n]{1,32}#`

	// base构造来源区分
	BaseSourceWeb = "base_source_web"

	// 预约卡按钮状态
	UpbuttonReservation       = 0 // 预约
	UpbuttonReservationOk     = 1 // 已预约
	UpbuttonCancel            = 2 // 取消预约
	UpbuttonCancelOk          = 3 // 已取消
	UpbuttonWatch             = 4 // 去观看
	UpbuttonReplay            = 5 // 回放
	UpbuttonEnd               = 6 // 已结束
	UpbuttonCancelLotteryCron = 7 // 取消预约抽奖

	// 动态的评论与点赞对应id info：https://info.bilibili.co/pages/viewpage.action?pageId=59401964
	ForwardCommentType = 17
	VideoCommentType   = 1
	DrawCommentType    = 11
	WordCommentType    = 17
	ArticleCommentType = 12
	CommonCommentType  = 17
	PgcCommentType     = 1

	// 隐式关联显示文案
	TopicHiddenAttatchedText = "编辑收录"
)

var (
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
				log.Error("AvPlayHandlerGRPCV2 ParamHandler url.Parse uri=%+v, error=%+v", uri, err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("AvPlayHandlerGRPCV2 ParamHandler url.ParseQuery uri=%+v, error=%+v", uri, err)
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
)

// nolint:gomnd
func MakeStorySuffixUrl(vmid, dynamicID, topicId, sortBy int64, offset, gt string) string {
	var scene string
	//1.推荐 2.热门 3.最新
	switch sortBy {
	case 1:
		scene = "topic_rcmd"
	case 2:
		scene = "topic_hot"
	case 3:
		scene = "topic_new"
	default:
		log.Error("MakeStorySuffixUrl unexpected sortBy vmid=%d dynamicID=%d topicId=%d sortBy=%d", vmid, dynamicID, topicId, sortBy)
		scene = "topic_new"
	}
	// 0: 动态 1：视频
	var topicType int
	switch gt {
	case GotoStory:
		topicType = 1
	default:
		topicType = 0
	}
	return fmt.Sprintf("scene=%s&vmid=%d&offset=%s&topic_id=%d&topic_rid=%d&topic_type=%d", scene, vmid, offset, topicId, dynamicID, topicType)
}

// FillURI deal app schema.
func FillURI(gt, param string, f func(uri string) string) string {
	if param == "" {
		return ""
	}
	var uri string
	switch gt {
	case GotoAv, "":
		uri = "bilibili://video/" + param
	case GotoSpaceDyn:
		uri = "bilibili://space/" + param + "?defaultTab=dynamic"
	case GotoDyn:
		uri = "bilibili://following/detail/" + param
	case GotoLive:
		uri = "bilibili://live/" + param
	case GotoArticle:
		uri = "bilibili://article/" + param + "?from=5"
	case GotoClip:
		uri = "bilibili://clip/detail/" + param + "/0"
	case GotoSpace:
		uri = "bilibili://space/" + param
	case GotoChannelSearch:
		uri = "bilibili://pegasus/channel/search?query=" + param
	case GotoTopicSearch:
		uri = "activity://following/topic_search?search_name=" + param + "&only_search=false&hotTopic=true"
	case GotoStory:
		uri = "bilibili://story/" + param
	case GotoURL:
		uri = param
	case GotoWebAv:
		uri = "//www.bilibili.com/video/" + param
	case GotoWebSpace:
		uri = "//space.bilibili.com/" + param
	default:
	}
	if f != nil {
		uri = f(uri)
	}
	return uri
}

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

// nolint:gomnd
func StatString(number int64, suffix string, defaultReturnValue string) string {
	if number <= 0 {
		return defaultReturnValue
	}
	if number < 10000 {
		return strconv.FormatInt(number, 10) + suffix
	}
	if number < 100000000 {
		s := strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s := strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

func MakeDynCmtMode(cmtMeta map[int64]*DynCmtMeta, dynamicID int64) (bool, int) {
	const (
		_cmtModeSingleKey = 0
		_cmtModeMultiKey  = 1

		_cmtModeSingleValue = 1
		_cmtModeMultiValue  = 3
	)
	v, ok := cmtMeta[dynamicID]
	if !ok || v == nil || v.CmtShowStat == 0 {
		return false, 0
	}
	switch v.CmtMode {
	case _cmtModeSingleKey:
		return true, _cmtModeSingleValue
	case _cmtModeMultiKey:
		return true, _cmtModeMultiValue
	default:
	}
	return false, 0
}

func ConstructPubTime(localTimeZone int32, timestamp int64) string {
	// 计算时区差值(默认服务端固定东八区)
	// 与客户端约定：东一至东十二区分别1到12; 0时区0; 西一至西十一分别-1到-11
	dd, _ := time.ParseDuration(fmt.Sprintf("%dh", localTimeZone-8))
	t := time.Unix(timestamp, 0)
	// 同步平移时间
	now := time.Now().Add(dd)
	sub := now.Sub(t.Add(dd))
	// 文案格式化
	if sub < time.Minute {
		return "刚刚"
	}
	if sub < time.Hour {
		return fmt.Sprintf("%v分钟前", math.Floor(sub.Minutes()))
	}
	if sub < 24*time.Hour {
		return fmt.Sprintf("%v小时前", math.Floor(sub.Hours()))
	}
	if now.Year() == t.Add(dd).Year() {
		if now.YearDay()-t.Add(dd).YearDay() == 1 {
			return "昨天"
		}
		return t.Add(dd).Format("01-02")
	}
	return t.Add(dd).Format("2006-01-02")
}
