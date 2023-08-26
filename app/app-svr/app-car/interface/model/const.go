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
	"go-gateway/pkg/idsafe/bvid"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
)

const (
	// 历史记录
	HistoryMax = 1200
	HistoryArc = "archive"
	HistoryPGC = "pgc"
)

const (
	// car
	AndroidBilithings = "android_bilithings"
	// PlatCar is int8 for car.
	PlatCar = int8(30)
	// PlatH5 is int8 for car h5
	PlatH5 = int8(31)
	// device
	DeviceCar   = "android_car"
	DeviceThing = "android_thing"
	// 风控相关
	SilverSourceLike         = "1"
	SilverSceneLike          = "thumbup_video"
	SilverActionLike         = "like"
	SilverGaiaCommonActivity = "gaia_common_activity"
	// 大会员
	VipBatchNotEnoughErr       = 69006  // 资源池数量不足
	CustomModuleRid51          = 100000 // 五一特辑
	CustomModuleRid61Childhood = 100001 // 61 童年回来了
	CustomModuleRid61Eden      = 100002 // 61 小朋友乐园
	CustomModuleRidDW          = 100003 // 端午 "粽”有陪伴
	FnvalDolby                 = 256    //请求杜比音轨

	ClosePersonalAi = 1 // 关闭个性化ai
)

type DeviceInfo struct {
	AccessKey string `form:"access_key"`
	AppKey    string `form:"appkey"`
	MobiApp   string `form:"mobi_app"`
	Device    string `form:"device"`
	Platform  string `form:"platform"`
	Channel   string `form:"channel"`
	Build     int    `form:"build"`
	Model     string `form:"model"` // 车型
}

// CardGt is
type CardGt string

// CardType is
type CardType string

// Icon is
type Icon string

const (
	// card goto
	CardGotoAv              = CardGt("av")
	CardGotoPGC             = CardGt("pgc")
	CardGotoDefalutFavorite = CardGt("default_favorite")
	CardGotoUserFavorite    = CardGt("user_favorite")
	CardGotoTopView         = CardGt("to_view")
	// goto
	GotoAv       = "av"
	GotoAvHis    = "av_his"
	GotoAvView   = "av_view"
	GotoPGC      = "pgc"
	GotoPGCEp    = "pgc_ep"
	GotoPGCEpHis = "pgc_ep_his"
	GotoFavorite = "favorite"
	GotoTopView  = "to_view"
	GotoSpace    = "space"
	GotoUp       = "up"
	GotoWebBV    = "web_bv"
	GotoWebPGC   = "web_pgc"

	SmallCoverV1    = CardType("small_cover_v1")
	SmallCoverV2    = CardType("small_cover_v2")
	SmallCoverV3    = CardType("small_cover_v3")
	VerticalCoverV1 = CardType("vertical_cover_v1")
	SmallCoverV4    = CardType("small_cover_v4")
	BannerV1        = CardType("banner_v1")
	FmV1            = CardType("fm_v1")

	IconPlay         = Icon("play")
	IconShow         = Icon("show")
	IconDanmaku      = Icon("danmaku")
	IconFavoritePlay = Icon("favorite_play")
	IconUp           = Icon("up")
	IconTop          = Icon("top")

	// 推荐角标填充样式
	BgStyleFill              = "fill"
	BgStyleStroke            = "stroke"
	BgStyleFillAndStroke     = "fill_stroke"
	BgStyleNoFillAndNoStroke = "no_fill_stroke"

	// 角标颜色
	BgColorRed    = "red"
	BgColorYellow = "yellow"
	BgColorBlue   = "blue"

	// 业务来源是列表、详情页、语音、小窗、音频
	FromList  = "list"
	FromView  = "view"
	FromVoice = "voice"
	FromAudio = "audio"

	// 客户端根据source_type类型也请求不同业务方接口
	EntrancePopular         = "popular"
	EntranceMyAnmie         = "my_anmie"
	EntrancePgcList         = "pgc_list"
	EntrancePgcRcmdList     = "pgc_rcmd_list"
	EntranceCommonSearch    = "common_search"
	EntranceVoiceSearch     = "voice_search"
	EntranceDynamicVideo    = "dynamic_video"
	EntranceDynamicVideoNew = "dynamic_video_new"
	EntranceHistoryRecord   = "history_record"
	EntranceRelate          = "relate"
	EntranceSpace           = "space"
	EntranceMyFavorite      = "my_favorite" // 我的收藏
	EntranceUpFavorite      = "up_favorite" // up主收藏
	EntranceMediaList       = "media_list"  // 播单
	EntranceToView          = "to_view"     // 稍后再看
	EntranceRegion          = "region_list" // 分区列表
	EntranceAudioFeed       = "audio_feed"
	EntranceAudioChannel    = "audio_channel"

	// AttrNo attribute no
	AttrNo = int32(0)
	// AttrYes attribute yes
	AttrYes = int32(1)

	// 过滤原因：互动视频
	FilterAttrBitSteinsGate = "AttrBitSteinsGate"

	// PGC视频类型
	// 番剧
	PGCTypeBangumi = 1
	// 电影
	PGCTypeMovie = 2
	// 纪录片
	PGCTypeDocumentary = 3
	// 国漫
	PGCTypeGc = 4
	// 电视剧
	PGCTypeTv = 5

	// 主站历史business
	PgcBusinesses = "pgc"
	UgcBusinesses = "archive"
)

var (
	// PGC角标颜色对应的车载服务的角标颜色
	PGCBageType = map[int32]string{
		0: BgColorRed,
		1: BgColorBlue,
		2: BgColorYellow,
	}

	ButtonText = map[string]string{
		GotoUp:  "关注",
		GotoPGC: "追番",
	}

	ArcPlayHandler = func(ap *arcgrpc.PlayerInfo) func(uri string) string {
		return func(uri string) string {
			if ap == nil || ap.Playurl == nil || ap.PlayerExtra == nil {
				return uri
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
			var bs []byte
			bs, _ = json.Marshal(ap.Playurl)
			player := string(bs)
			if player == "" {
				return uri
			}
			params.Set("player_preload", player)
			if ap.PlayerExtra.Dimension != nil && (ap.PlayerExtra.Dimension.Height != 0 || ap.PlayerExtra.Dimension.Width != 0) {
				params.Set("player_width", strconv.FormatInt(ap.PlayerExtra.Dimension.Width, 10))
				params.Set("player_height", strconv.FormatInt(ap.PlayerExtra.Dimension.Height, 10))
				params.Set("player_rotate", strconv.FormatInt(ap.PlayerExtra.Dimension.Rotate, 10))
			}
			params.Set("history_progress", strconv.FormatInt(ap.PlayerExtra.Progress, 10))
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}
	PGCPlayHandler = func(e *pgcinline.EpisodeCard) func(uri string) string {
		return func(uri string) string {
			if e.PlayerInfo == nil {
				return uri
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
			var bs []byte
			bs, _ = json.Marshal(e.PlayerInfo)
			player := string(bs)
			if player == "" {
				return uri
			}
			params.Set("player_preload", player)
			if e.Dimension.Height != 0 || e.Dimension.Width != 0 {
				params.Set("player_width", strconv.Itoa(int(e.Dimension.Width)))
				params.Set("player_height", strconv.Itoa(int(e.Dimension.Height)))
				params.Set("player_rotate", strconv.Itoa(int(e.Dimension.Rotate)))
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

	ParamHandler = func(main interface{}, cid, rid int64, entrance, followtype, keyword string) func(uri string) string {
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
			// 特殊参数用于进入接口列表做插入逻辑使用
			if main != nil {
				b, _ := json.Marshal(main)
				params.Set("param", string(b))
			}
			if cid > 0 {
				params.Set("cid", strconv.FormatInt(cid, 10))
			}
			if rid > 0 {
				params.Set("rid", strconv.FormatInt(rid, 10))
			}
			params.Set("sourceType", entrance)
			if followtype != "" {
				params.Set("followType", followtype)
			}
			if keyword != "" {
				params.Set("keyword", keyword)
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
	SpaceHandler = func(vmid int64) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("SpaceHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("SpaceHandler url.ParseQuery error(%v)", err)
				return uri
			}
			if vmid > 0 {
				params.Set("vmid", strconv.FormatInt(vmid, 10))
			}
			u.RawQuery = params.Encode()
			return u.String()
		}
	}
	FavHandler = func(favid, vmid int64) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("FavHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("FavHandler url.ParseQuery error(%v)", err)
				return uri
			}
			if favid > 0 {
				params.Set("fav_id", strconv.FormatInt(favid, 10))
			}
			if vmid > 0 {
				params.Set("vmid", strconv.FormatInt(vmid, 10))
			}
			u.RawQuery = params.Encode()
			return u.String()
		}
	}
	MediaHandler = func(vmid int64) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("MediaHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("MediaHandler url.ParseQuery error(%v)", err)
				return uri
			}
			if vmid > 0 {
				params.Set("vmid", strconv.FormatInt(vmid, 10))
			}
			u.RawQuery = params.Encode()
			return u.String()
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
	AvPlayHandlerGRPC = func(ap *arcgrpc.ArcPlayer, cid int64, showPlayerURL bool) func(uri string) string {
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
)

// FillURI deal app schema.
func FillURI(gt string, plat int8, build int, param string, f func(uri string) string) (uri string) {
	switch gt {
	case GotoAv, GotoPGC:
		if param != "" {
			uri = fmt.Sprintf("bilithings://player?goto=%s&aid=%s", gt, param)
		}
	case GotoFavorite:
		if param != "" {
			uri = fmt.Sprintf("bilithings://player?sourceType=%s&fav_id=%s", EntranceMediaList, param)
		}
	case GotoTopView:
		uri = fmt.Sprintf("bilithings://player?sourceType=%s", EntranceToView)
	case GotoSpace:
		if param != "" {
			uri = fmt.Sprintf("bilithings://user?vmid=%s", param)
		}
	case GotoWebBV:
		if param != "" {
			uri = fmt.Sprintf("https://www.bilibili.com/video/%s?share_source=copy_web", param)
		}
	case GotoWebPGC:
		if param != "" {
			uri = fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%s?share_source=copy_web", param)
		}
	default:
		uri = param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

// StatString Stat to string
func StatString(number int32, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	// nolint:gomnd
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	// nolint:gomnd
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// StatIntString Stat to string
func StatIntString(number int32, suffix string, isInt bool) (s string) {
	if number == 0 {
		s = "0" + suffix
		return
	}
	// nolint:gomnd
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	// nolint:gomnd
	if number < 100000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	// nolint:gomnd
	if number < 100000000 {
		if isInt {
			s = strconv.FormatInt(int64(number)/10000, 10)
			return s + "万" + suffix
		}
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return s + "亿" + suffix
}

// StatString64 Stat to string
func StatString64(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	// nolint:gomnd
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	// nolint:gomnd
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// BangumiFavString BangumiFav to string
func BangumiFavString(number, tp int32) string {
	var suffix = "追番"
	switch tp {
	// nolint:gomnd
	case 2, 3, 5, 7:
		suffix = "追剧"
	}
	return StatString(number, suffix)
}

func BangumiTotalCountString(value string, IsFinish int) string {
	if IsFinish == 1 {
		return fmt.Sprintf("已完结，全%s话", value)
	}
	return fmt.Sprintf("更新至%s话", value)
}

func FavoriteCountString(number int32) string {
	return fmt.Sprintf("%d个内容", number)
}

// DurationString duration to string
func DurationString(second int64) (s string) {
	var hour, min, sec int
	if second < 1 {
		return
	}
	d, err := time.ParseDuration(strconv.FormatInt(second, 10) + "s")
	if err != nil {
		log.Error("%+v", err)
		return
	}
	r := strings.NewReplacer("h", ":", "m", ":", "s", ":")
	ts := strings.Split(strings.TrimSuffix(r.Replace(d.String()), ":"), ":")
	if len(ts) == 1 {
		sec, _ = strconv.Atoi(ts[0])
	} else if len(ts) == 2 { // nolint:gomnd
		min, _ = strconv.Atoi(ts[0])
		sec, _ = strconv.Atoi(ts[1])
	} else if len(ts) == 3 { // nolint:gomnd
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

// HisDurationString duration to string
func HisDurationString(second int64) string {
	if second < 1 {
		return "0:00"
	}
	return DurationString(second)
}

// Plat return plat by platStr or mobiApp
func Plat(mobiApp, device string) int8 {
	return PlatCar
}

// HisPubDataString is.
func HisPubDataString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	if now.Year() == t.Year() {
		if now.YearDay()-t.YearDay() == 1 {
			return fmt.Sprintf("昨天 %v", t.Format("15:04"))
		}
		if now.YearDay()-t.YearDay() == 0 {
			return fmt.Sprintf("今天 %v", t.Format("15:04"))
		}
		return t.Format("01-02 15:04")
	}
	return t.Format("2006-01-02 15:04")
}

func PubDataString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	if now.Year() == t.Year() {
		return t.Format("01-02")
	}
	return t.Format("2006-01-02")
}

func ReplyDataString(second int64) string {
	const (
		_week = 7
	)
	// 同步平移时间
	now := time.Now()
	t := time.Unix(second, 0)
	sub := now.Sub(t)
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
	if now.Year() == t.Year() {
		if now.YearDay()-t.YearDay() == 1 {
			return "昨天"
		}
		if day := now.YearDay() - t.YearDay(); day <= _week {
			return fmt.Sprintf("%d天前", day)
		}
		return t.Format("01-02")
	}
	return t.Format("2006-01-02")
}

// HighLightString is.
func HighLightString(value string) string {
	return fmt.Sprintf("<em class=\"keyword\">%s</em>", value)
}

// FanString fan to string
func FanString(number int32) string {
	const _suffix = "粉丝"
	return StatString(number, _suffix)
}

// FanIntString fan to string
func FanIntString(number int32) string {
	const _suffix = "粉丝"
	return StatIntString(number, _suffix, true)
}

// VedioString fan to string
func VedioString(number int32) string {
	const _suffix = "视频"
	return StatString(number, _suffix)
}

// VedioString fan to string
func VedioIntString(number int32) string {
	const _suffix = "视频"
	return StatIntString(number, _suffix, false)
}

// AttrVal get attribute value
func AttrVal(attr int32, bit uint32) (v int32) {
	v = (attr >> bit) & int32(1)
	return
}

type Relation struct {
	Status     int8 `json:"status,omitempty"`
	IsFollow   int8 `json:"is_follow,omitempty"`
	IsFollowed int8 `json:"is_followed,omitempty"`
}

// 互相关注关系转换
func RelationChange(upMid int64, relations map[int64]*relationgrpc.InterrelationReply) (r *Relation) {
	const (
		// state使用
		_statenofollow      = 1
		_statefollow        = 2
		_statefollowed      = 3
		_statemutualConcern = 4
		// 关注关系
		_follow = 1
	)
	r = &Relation{
		Status: _statenofollow,
	}
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	switch rel.Attribute {
	// nolint:gomnd
	case 2, 6: // 用户关注UP主
		r.Status = _statefollow
		r.IsFollow = _follow
	}
	if rel.IsFollowed { // UP主关注用户
		r.Status = _statefollowed
		r.IsFollowed = _follow
	}
	if r.IsFollow == _follow && r.IsFollowed == _follow { // 用户和UP主互相关注
		r.Status = _statemutualConcern
	}
	return
}

func AvIsNormalGRPC(a *arcgrpc.ArcWithPlayurl) bool {
	if a == nil || a.Arc == nil || a.Arc.FirstCid == 0 {
		return false
	}
	return a.IsNormal()
}

func AvIsNormal(a *arcgrpc.Arc) bool {
	if a == nil || a.FirstCid == 0 {
		return false
	}
	return a.IsNormal()
}

func AvIsNormalView(a *arcgrpc.ViewReply) bool {
	if a == nil || a.FirstCid == 0 {
		return false
	}
	return a.IsNormal()
}

func GetBvID(input int64) (bid string, err error) {
	if bid, err = bvid.AvToBv(input); err != nil {
		return "", fmt.Errorf("视频ID非法！")
	}
	return
}

func TimeToUnix(timeStr string) (time.Time, error) {
	timeTemplate1 := "2006-01-02 15:04:05"                                 //常规类型
	stamp, err := time.ParseInLocation(timeTemplate1, timeStr, time.Local) //使用parseInLocation将字符串格式化返回本地时区时间
	return stamp, err
}

func PGCTypeValue(pgcType int) (string, error) {
	switch pgcType {
	case PGCTypeBangumi:
		return "番剧", nil
	case PGCTypeMovie:
		return "电影", nil
	case PGCTypeDocumentary:
		return "纪录片", nil
	case PGCTypeGc:
		return "国漫", nil
	case PGCTypeTv:
		return "电视剧", nil
	default:
		return "", fmt.Errorf("PGC类型不存在 type(%d)", pgcType)
	}
}

func ViewInfo(info, addValue, splice string) string {
	if addValue == "" {
		return info
	}
	if info == "" {
		return addValue
	}
	return info + splice + addValue
}

func BangumiRating(rating float64, suffix string) string {
	return fmt.Sprintf("%.1f%s", rating, suffix)
}

func CheckMidMaxInt32(mid int64, build int) bool {
	if mid > math.MaxInt32 && (build >= 1000000 && build < 1060000) {
		return true
	}
	return false
}

type Prune struct {
	// PGC
	SeasonID int64 `json:"season_id,omitempty"`
	// His
	Business string `json:"business,omitempty"`
	Oid      int64  `json:"oid,omitempty"`
	Cid      int64  `json:"cid,omitempty"`
	Epid     int64  `json:"epid,omitempty"`
	// popular、search
	Goto    string `json:"goto,omitempty"`
	ID      int64  `json:"id,omitempty"`
	ChildID int64  `json:"child_id,omitempty"`
	// dynamic
	DynamicID int64 `json:"dynamic_id,omitempty"`
	Dtype     int64 `json:"dtype,omitempty"`
	Drid      int64 `json:"drid,omitempty"`
}
