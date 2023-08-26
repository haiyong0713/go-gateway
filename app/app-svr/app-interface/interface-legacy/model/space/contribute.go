package space

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/audio"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/bplus"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/comic"
	"go-gateway/app/app-svr/archive/service/api"

	article "git.bilibili.co/bapis/bapis-go/article/model"
)

const (
	_gotoAv      = 0
	_gotoArticle = 1
	_gotoClip    = 2
	_gotoAlbum   = 3
	_gotoAudio   = 4
	_gotoComic   = 5

	ComicStatusNotStart      = -1
	ComicStatusSerialization = 0
	ComicStatusFinished      = 1
	//主页tab
	HomeTab        = "home"
	DyanmicTab     = "dynamic"
	ContributeTab  = "contribute"
	AllTab         = "all"
	VideoTab       = "video"
	SeasonTab      = "season"
	SeasonVideoTab = "season_video"
	ArticleTab     = "article"
	AudiosTab      = "audio"
	ComicTab       = "comic"
	AlbumTab       = "album"
	ClipTab        = "clip"
	ShopTab        = "shop"
	FavoriteTab    = "favorite"
	BangumiTab     = "bangumi"
	CheeseTab      = "cheese"
	SeriesTab      = "series"
	// 投稿页排序
	ArchiveNew  = "pubdate"
	ArchivePlay = "click"
	ActivityTab = "activity"
)

// Contributes struct
type Contributes struct {
	Tab   *Tab    `json:"tab,omitempty"`
	Items []*Item `json:"items,omitempty"`
	Links *Links  `json:"links,omitempty"`
}

// Tab struct
type Tab struct {
	Archive   bool `json:"archive"`
	Article   bool `json:"article"`
	Clip      bool `json:"clip"`
	Album     bool `json:"album"`
	Favorite  bool `json:"favorite"`
	Bangumi   bool `json:"bangumi"`
	Coin      bool `json:"coin"`
	Like      bool `json:"like"`
	Community bool `json:"community"`
	Dynamic   bool `json:"dynamic"`
	Audios    bool `json:"audios"`
	Shop      bool `json:"shop"`
	Mall      bool `json:"mall"`
	UGCSeason bool `json:"ugc_season"`
	Comic     bool `json:"comic"`
	Cheese    bool `json:"cheese"`
	SubComic  bool `json:"sub_comic"`
	Activity  bool `json:"activity"`
	Series    bool `json:"series"`
}

// Item struct
type Item struct {
	ID            int64                `json:"id,omitempty"`
	TypeName      string               `json:"tname,omitempty"`
	Category      *article.Category    `json:"category,omitempty"`
	Title         string               `json:"title,omitempty"`
	Cover         string               `json:"cover,omitempty"`
	Tag           string               `json:"tag,omitempty"`
	Tags          []*article.Tag       `json:"tags,omitempty"`
	Desc          string               `json:"description"`
	URI           string               `json:"uri,omitempty"`
	Param         string               `json:"param,omitempty"`
	Goto          string               `json:"goto,omitempty"`
	Length        string               `json:"length,omitempty"`
	Duration      int64                `json:"duration,omitempty"`
	Banner        string               `json:"banner,omitempty"`
	Play          int                  `json:"play,omitempty"`
	Comment       int                  `json:"comment,omitempty"`
	Danmaku       int                  `json:"danmaku,omitempty"`
	Count         int                  `json:"count,omitempty"`
	Reply         int                  `json:"reply,omitempty"`
	CTime         xtime.Time           `json:"ctime,omitempty"`
	MTime         xtime.Time           `json:"mtime,omitempty"`
	ImageURLs     []string             `json:"image_urls,omitempty"`
	Pictures      []*bplus.Pictures    `json:"pictures,omitempty"`
	Words         int64                `json:"words,omitempty"`
	Stats         *article.Stats       `json:"stats,omitempty"`
	AuthType      int                  `json:"authType,omitempty"`
	Member        int64                `json:"member,omitempty"`
	Badges        []*model.ReasonStyle `json:"badges,omitempty"`
	Styles        string               `json:"styles,omitempty"`
	Label         string               `json:"label,omitempty"`
	IsUGCPay      bool                 `json:"is_ugcpay"`
	IsCooperation bool                 `json:"is_cooperation"`
	IsSteins      bool                 `json:"is_steins"`
	IsPopular     bool                 `json:"is_popular"`
}

// Links struct
type Links struct {
	Previous int64 `json:"previous,omitempty"`
	Next     int64 `json:"next,omitempty"`
}

// Link func
func (l *Links) Link(sinceID, untilID int64) {
	if sinceID < 0 || untilID < 0 {
		return
	}
	l.Previous = sinceID
	l.Next = untilID
}

// Items struct
type Items []*Item

// Len()
func (is Items) Len() int { return len(is) }

// Less()
func (is Items) Less(i, j int) bool {
	var it, jt xtime.Time
	if is[i] != nil {
		it = is[i].CTime
	}
	if is[j] != nil {
		jt = is[j].CTime
	}
	return it > jt
}

// Swap()
func (is Items) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

// Clip struct
type Clip struct {
	ID       int64      `json:"id"`
	Duration int64      `json:"duration"`
	CTime    xtime.Time `json:"ctime"`
	View     int        `json:"view"`
	Damaku   int        `json:"damaku"`
	Title    string     `json:"title"`
	Cover    string     `json:"cover"`
	Tag      string     `json:"tag"`
}

// Album struct
type Album struct {
	ID       int64       `json:"doc_id"`
	CTime    xtime.Time  `json:"ctime"`
	Count    int         `json:"count"`
	View     int         `json:"view"`
	Comment  int         `json:"comment"`
	Title    string      `json:"title"`
	Desc     string      `json:"description"`
	Pictures []*Pictures `json:"pictures"`
}

// Pictures struct
type Pictures struct {
	ImgSrc    string `json:"img_src"`
	ImgWidth  string `json:"img_width"`
	ImgHeight string `json:"img_height"`
}

// Tag tag.
type Tag struct {
	Tid  int64  `json:"tid"`
	Name string `json:"name"`
}

// FromArc3 func
func (i *Item) FromArc3(a *api.Arc, popularAIDs map[int64]struct{}) {
	i.ID = a.Aid
	i.Title = a.Title
	i.Cover = a.Pic
	i.TypeName = a.TypeName
	i.Param = strconv.FormatInt(a.Aid, 10)
	i.Goto = model.GotoAv
	if a.AttrVal(api.AttrBitIsPGC) == api.AttrYes && a.RedirectURL != "" {
		i.URI = a.RedirectURL
	} else {
		i.URI = model.FillURI(i.Goto, i.Param, model.AvHandler(a))
	}
	i.Danmaku = int(a.Stat.Danmaku)
	i.Duration = a.Duration
	i.CTime = a.PubDate
	i.Play = int(a.Stat.View)
	if a.Rights.UGCPay == 1 {
		i.IsUGCPay = true
		i.Badges = append(i.Badges, model.PayBadge)
	}
	if a.Rights.IsCooperation == 1 {
		i.IsCooperation = true
		i.Badges = append(i.Badges, model.CooperationBadge)
	}
	if a.AttrVal(api.AttrBitSteinsGate) == api.AttrYes {
		i.IsSteins = true
		i.Badges = append(i.Badges, model.SteinsBadge)
	}
	if _, ok := popularAIDs[a.Aid]; ok {
		i.IsPopular = true
		i.Badges = append(i.Badges, model.PopularBadge)
	}
	// 最多展示两个角标
	//nolint:gomnd
	if len(i.Badges) > 2 {
		i.Badges = i.Badges[:2]
	}
}

// FromArticle func
func (i *Item) FromArticle(a *article.Meta) {
	i.ID = a.ID
	i.Title = a.Title
	i.Category = a.Category
	i.Desc = a.Summary
	i.ImageURLs = a.ImageURLs
	i.CTime = a.PublishTime
	i.Tags = a.Tags
	i.Banner = a.BannerURL
	i.Param = strconv.FormatInt(a.ID, 10)
	i.Goto = model.GotoArticle
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Stats = a.Stats
}

// FromClip func
func (i *Item) FromClip(c *bplus.Clip) {
	i.ID = c.ID
	i.Duration = c.Duration
	i.CTime = c.CTime
	i.Play = c.View
	i.Danmaku = c.Damaku
	i.Param = strconv.FormatInt(c.ID, 10)
	i.Goto = model.GotoClip
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Title = c.Title
	i.Cover = c.Cover
	i.Tag = c.Tag
}

// FromAlbum func
func (i *Item) FromAlbum(a *bplus.Album) {
	i.ID = a.ID
	i.CTime = a.CTime
	i.Count = a.Count
	i.Play = a.View
	i.Comment = a.Comment
	i.Param = strconv.FormatInt(a.ID, 10)
	i.Goto = model.GotoAlbum
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Title = a.Title
	i.Desc = a.Desc
	i.Pictures = a.Pictures
}

// FromAlbum func
func (i *Item) FromComic(c *comic.Comic) {
	i.Title = c.Title
	i.Cover = c.VerticalCover
	i.Param = strconv.FormatInt(c.ID, 10)
	i.URI = c.URL
	i.Goto = model.GotoComic
	i.Count = c.Total
	var styles []string
	for _, style := range c.Styles {
		styles = append(styles, style.Name)
	}
	if len(styles) > 0 {
		i.Styles = strings.Join(styles, " ")
	}
	update, _ := strconv.ParseInt(c.LastUpdateTime, 10, 64)
	switch c.IsFinish {
	case ComicStatusSerialization:
		if update != 0 || c.LastShortTitle != "" {
			i.Label = fmt.Sprintf("%v更新至%v", time.Unix(update, 0).Format("01-02"), c.LastShortTitle)
		}
	case ComicStatusFinished:
		i.Label = fmt.Sprintf("全%v话", c.Total)
	}
	i.CTime = xtime.Time(update)
}

// FromAudio func
func (i *Item) FromAudio(a *audio.Audio) {
	i.ID = a.ID
	i.CTime = a.CTime
	i.Play = a.Play
	i.Reply = a.Reply
	i.Param = strconv.FormatInt(a.ID, 10)
	i.Goto = model.GotoAudio
	i.URI = a.Schema
	i.Cover = a.Cover
	i.Title = a.Title
	i.AuthType = a.AuthType
	i.Duration = a.Duration
}

// FormatKey func
func (i *Item) FormatKey() {
	switch i.Goto {
	case model.GotoAv:
		//nolint:gomnd
		i.Member = i.ID<<6 | _gotoAv
	case model.GotoArticle:
		//nolint:gomnd
		i.Member = i.ID<<6 | _gotoArticle
	case model.GotoClip:
		//nolint:gomnd
		i.Member = i.ID<<6 | _gotoClip
	case model.GotoAlbum:
		//nolint:gomnd
		i.Member = i.ID<<6 | _gotoAlbum
	case model.GotoAudio:
		//nolint:gomnd
		i.Member = i.ID<<6 | _gotoAudio
	case model.GotoComic:
		//nolint:gomnd
		i.Member = i.ID<<6 | _gotoComic
	default:
		i.Member = i.ID
	}
}

// ParseKey func
func (i *Item) ParseKey() {
	//nolint:gomnd
	i.ID = i.Member >> 6
	switch int(i.Member & 0x3f) {
	case _gotoAv:
		i.Goto = model.GotoAv
	case _gotoArticle:
		i.Goto = model.GotoArticle
	case _gotoClip:
		i.Goto = model.GotoClip
	case _gotoAlbum:
		i.Goto = model.GotoAlbum
	case _gotoAudio:
		i.Goto = model.GotoAudio
	case _gotoComic:
		i.Goto = model.GotoComic
	}
}

// Attrs struct
type Attrs struct {
	Archive bool `json:"archive,omitempty"`
	Article bool `json:"article,omitempty"`
	Clip    bool `json:"clip,omitempty"`
	Album   bool `json:"album,omitempty"`
	Audio   bool `json:"audio,omitempty"`
	Comic   bool `json:"comic,omitempty"`
}
