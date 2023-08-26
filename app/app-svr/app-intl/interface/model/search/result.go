package search

import (
	"bytes"
	"fmt"

	// "hash/crc32"
	"regexp"
	"strconv"
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"
	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	"go-gateway/app/app-svr/app-intl/interface/model"
	bangumimdl "go-gateway/app/app-svr/app-intl/interface/model/bangumi"
	"go-gateway/app/app-svr/archive/service/api"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
)

const (
	_styleGrid       = "grid"       // 默认宫格
	_styleHorizontal = "horizontal" // 分集展示按照横条样式
)

// search const
var (
	getHightLight = regexp.MustCompile(`<em.*?em>`)

	videoStrongStyle = &model.ReasonStyle{
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FAAB4B",
		BgColorNight:     "#BA833F",
		BorderColor:      "#FAAB4B",
		BorderColorNight: "#BA833F",
		BgStyle:          model.BgStyleFill,
	}
	videoWeekStyle = &model.ReasonStyle{
		TextColor:        "#FAAB4B",
		TextColorNight:   "#BA833F",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#FAAB4B",
		BorderColorNight: "#BA833F",
		BgStyle:          model.BgStyleStroke,
	}
)

// Result struct
type Result struct {
	Trackid   string      `json:"trackid,omitempty"`
	Page      int         `json:"page,omitempty"`
	NavInfo   []*NavInfo  `json:"nav,omitempty"`
	Item      []*Item     `json:"item,omitempty"`
	Items     ResultItems `json:"items,omitempty"`
	Array     int         `json:"array,omitempty"`
	Attribute int32       `json:"attribute"`
	EasterEgg *EasterEgg  `json:"easter_egg,omitempty"`
}

// ResultItems struct
type ResultItems struct {
	SuggestKeyWord *Item   `json:"suggest_keyword,omitempty"`
	Operation      []*Item `json:"operation,omitempty"`
	Season2        []*Item `json:"season2,omitempty"`
	Season         []*Item `json:"season,omitempty"`
	Upper          []*Item `json:"upper,omitempty"`
	Movie2         []*Item `json:"movie2,omitempty"`
	Movie          []*Item `json:"movie,omitempty"`
	Archive        []*Item `json:"archive,omitempty"`
	LiveRoom       []*Item `json:"live_room,omitempty"`
	LiveUser       []*Item `json:"live_user,omitempty"`
}

// NavInfo struct
type NavInfo struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
	Pages int    `json:"pages"`
	Type  int    `json:"type"`
	Show  int    `json:"show_more,omitempty"`
}

// TypeSearch struct
type TypeSearch struct {
	TrackID string  `json:"trackid"`
	Pages   int     `json:"pages"`
	Total   int     `json:"total"`
	ExpStr  string  `json:"exp_str"`
	Items   []*Item `json:"items,omitempty"`
}

// Suggestion struct
type Suggestion struct {
	TrackID string      `json:"trackid"`
	UpUser  interface{} `json:"upuser,omitempty"`
	Bangumi interface{} `json:"bangumi,omitempty"`
	Suggest []string    `json:"suggest,omitempty"`
}

// SuggestionResult3 struct
type SuggestionResult3 struct {
	TrackID string  `json:"trackid"`
	List    []*Item `json:"list,omitempty"`
}

// NoResultRcndResult struct
type NoResultRcndResult struct {
	TrackID string  `json:"trackid"`
	Title   string  `json:"title,omitempty"`
	Pages   int     `json:"pages"`
	ExpStr  string  `json:"exp_str"`
	Items   []*Item `json:"items,omitempty"`
}

// EasterEgg struct
type EasterEgg struct {
	ID        int64 `json:"id,omitempty"`
	ShowCount int   `json:"show_count,omitempty"`
}

// Item struct
type Item struct {
	TrackID        string `json:"trackid,omitempty"`
	LinkType       string `json:"linktype,omitempty"`
	Position       int    `json:"position,omitempty"`
	SuggestKeyword string `json:"suggest_keyword,omitempty"`
	Title          string `json:"title,omitempty"`
	Name           string `json:"name,omitempty"`
	Cover          string `json:"cover,omitempty"`
	URI            string `json:"uri,omitempty"`
	Param          string `json:"param,omitempty"`
	Goto           string `json:"goto,omitempty"`
	// av
	Play       int                  `json:"play,omitempty"`
	Danmaku    int                  `json:"danmaku,omitempty"`
	Author     string               `json:"author,omitempty"`
	ViewType   string               `json:"view_type,omitempty"`
	PTime      xtime.Time           `json:"ptime,omitempty"`
	RecTags    []string             `json:"rec_tags,omitempty"`
	NewRecTags []*model.ReasonStyle `json:"new_rec_tags,omitempty"`
	// bangumi season
	SeasonID       int64   `json:"season_id,omitempty"`
	SeasonType     int     `json:"season_type,omitempty"`
	SeasonTypeName string  `json:"season_type_name,omitempty"`
	Finish         int8    `json:"finish,omitempty"`
	Started        int8    `json:"started,omitempty"`
	Index          string  `json:"index,omitempty"`
	NewestCat      string  `json:"newest_cat,omitempty"`
	NewestSeason   string  `json:"newest_season,omitempty"`
	CatDesc        string  `json:"cat_desc,omitempty"`
	TotalCount     int     `json:"total_count,omitempty"`
	MediaType      int     `json:"media_type,omitempty"`
	PlayState      int     `json:"play_state,omitempty"`
	Style          string  `json:"style,omitempty"`
	Styles         string  `json:"styles,omitempty"`
	CV             string  `json:"cv,omitempty"`
	Rating         float64 `json:"rating,omitempty"`
	Vote           int     `json:"vote,omitempty"`
	RatingCount    int     `json:"rating_count,omitempty"`
	BadgeType      int     `json:"badge_type,omitempty"`
	// upper
	Sign           string          `json:"sign,omitempty"`
	Fans           int             `json:"fans,omitempty"`
	Level          int             `json:"level,omitempty"`
	Desc           string          `json:"desc,omitempty"`
	OfficialVerify *OfficialVerify `json:"official_verify,omitempty"`
	AvItems        []*Item         `json:"av_items,omitempty"`
	Item           []*Item         `json:"item,omitempty"`
	CTime          int64           `json:"ctime,omitempty"`
	IsUp           bool            `json:"is_up,omitempty"`
	LiveURI        string          `json:"live_uri,omitempty"`
	// movie
	ScreenDate string `json:"screen_date,omitempty"`
	Area       string `json:"area,omitempty"`
	CoverMark  string `json:"cover_mark,omitempty"`
	// arc and sp
	Arcs int `json:"archives,omitempty"`
	// arc and movie
	Duration    string `json:"duration,omitempty"`
	DurationInt int64  `json:"duration_int,omitempty"`
	Actors      string `json:"actors,omitempty"`
	Staff       string `json:"staff,omitempty"`
	Length      int    `json:"length,omitempty"`
	Status      int    `json:"status,omitempty"`
	// live
	RoomID      int64  `json:"roomid,omitempty"`
	Mid         int64  `json:"mid,omitempty"`
	Type        string `json:"type,omitempty"`
	Attentions  int    `json:"attentions,omitempty"`
	LiveStatus  int    `json:"live_status,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Region      int    `json:"region,omitempty"`
	Online      int    `json:"online,omitempty"`
	ShortID     int    `json:"short_id,omitempty"`
	CateName    string `json:"area_v2_name,omitempty"`
	IsSelection int    `json:"is_selection,omitempty"`
	// article
	ID         int64    `json:"id,omitempty"`
	TemplateID int      `json:"template_id,omitempty"`
	ImageUrls  []string `json:"image_urls,omitempty"`
	View       int      `json:"view,omitempty"`
	Like       int      `json:"like,omitempty"`
	Reply      int      `json:"reply,omitempty"`
	// special
	Badge      string      `json:"badge,omitempty"`
	RcmdReason *RcmdReason `json:"rcmd_reason,omitempty"`
	// media bangumi and mdeia ft
	Prompt         string        `json:"prompt,omitempty"`
	Episodes       []*Item       `json:"episodes,omitempty"`
	Label          string        `json:"label,omitempty"`
	WatchButton    *WatchButton  `json:"watch_button,omitempty"`
	FollowButton   *FollowButton `json:"follow_button,omitempty"`
	EpisodesNew    []*EpisodeNew `json:"episodes_new,omitempty"`
	SelectionStyle string        `json:"selection_style,omitempty"` // grid || horizontal
	IsOut          int           `json:"is_out,omitempty"`          // is all_net_search
	CheckMore      *CheckMore    `json:"check_more,omitempty"`
	// game
	Reserve string `json:"reserve,omitempty"`
	// user
	Face string `json:"face,omitempty"`
	// suggest
	From      string  `json:"from,omitempty"`
	KeyWord   string  `json:"keyword,omitempty"`
	CoverSize float64 `json:"cover_size,omitempty"`
	SugType   string  `json:"sug_type,omitempty"`
	TermType  int     `json:"term_type,omitempty"`
	// rcmd query
	List       []*Item `json:"list,omitempty"`
	FromSource string  `json:"from_source,omitempty"`
	// live master
	UCover         string `json:"ucover,omitempty"`
	VerifyType     int    `json:"verify_type,omitempty"`
	VerifyDesc     string `json:"verify_desc,omitempty"`
	LevelColor     int64  `json:"level_color,omitempty"`
	IsAttention    int    `json:"is_atten,omitempty"`
	CateParentName string `json:"cate_parent_name,omitempty"`
	CateNameNew    string `json:"cate_name,omitempty"`
	Glory          *Glory `json:"glory_info,omitempty"`
	// twitter
	Covers     []string             `json:"covers,omitempty"`
	CoverCount int                  `json:"cover_count,omitempty"`
	Badges     []*model.ReasonStyle `json:"badges,omitempty"`
	// suggest_keyword
	SugKeyWordType int `json:"sugKeyWord_type,omitempty"`
	// operate
	ContentURI string `json:"content_uri,omitempty"`
	// 回粉
	Relation *cardmdl.Relation `json:"relation"`
}

// EpisodeNew is new structure of episode given by pgc grpc
type EpisodeNew struct {
	Title    string                        `json:"title,omitempty"`
	Uri      string                        `json:"uri,omitempty"`
	Param    string                        `json:"param,omitempty"`
	IsNew    int32                         `json:"is_new"`           // 1=is new, 0=not new
	Badges   []*pgcsearch.SearchBadgeProto `json:"badges,omitempty"` // badges
	Type     int32                         `json:"type,omitempty"`
	Position int                           `json:"position,omitempty"`
}

// CheckMore is displayed only if none of episode has been hit
type CheckMore struct {
	Content string `json:"content"`
	Uri     string `json:"uri"`
}

// WatchButton is the button of watch
type WatchButton struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

// FollowButton is the button of follow
type FollowButton struct {
	Icon         string            `json:"icon"`
	Texts        map[string]string `json:"texts,omitempty"`
	StatusReport string            `json:"status_report"`
}

// FromPGCCard builds the follow button from search card of PGC
func (v *FollowButton) FromPGCCard(card *pgcsearch.SearchFollowProto) {
	if card == nil {
		return
	}
	v.Icon = card.Icon
	v.StatusReport = card.StatusReport
	if len(card.Text) > 0 {
		v.Texts = make(map[string]string, len(card.Text))
		for key, value := range card.Text {
			v.Texts[fmt.Sprintf("%d", key)] = value
		}
	}
}

// Glory live struct
type Glory struct {
	Title string  `json:"title,omitempty"`
	Total int     `json:"total"`
	Items []*Item `json:"items,omitempty"`
}

// RcmdReason struct
type RcmdReason struct {
	Content string `json:"content,omitempty"`
}

// UserResult struct
type UserResult struct {
	Items []*Item `json:"items,omitempty"`
}

// DefaultWords struct
type DefaultWords struct {
	Trackid string `json:"trackid,omitempty"`
	Param   string `json:"param,omitempty"`
	Show    string `json:"show,omitempty"`
	Word    string `json:"word,omitempty"`
}

// FromSeason form func
func (i *Item) FromSeason(b *Bangumi, bangumi string) {
	i.Title = b.Title
	i.Cover = b.Cover
	i.Goto = model.GotoBangumi
	i.Param = strconv.Itoa(int(b.SeasonID))
	i.URI = model.FillURI(bangumi, i.Param, nil)
	i.Finish = int8(b.IsFinish)
	i.Started = int8(b.IsStarted)
	i.Index = b.NewestEpIndex
	i.NewestCat = b.NewestCat
	i.NewestSeason = b.NewestSeason
	i.TotalCount = b.TotalCount
	var buf bytes.Buffer
	if b.CatList.TV != 0 {
		buf.WriteString(`TV(`)
		buf.WriteString(strconv.Itoa(b.CatList.TV))
		buf.WriteString(`) `)
	}
	if b.CatList.Movie != 0 {
		buf.WriteString(`剧场版(`)
		buf.WriteString(strconv.Itoa(b.CatList.Movie))
		buf.WriteString(`) `)
	}
	if b.CatList.Ova != 0 {
		buf.WriteString(`OVA/OAD/SP(`)
		buf.WriteString(strconv.Itoa(b.CatList.Ova))
		buf.WriteString(`)`)
	}
	i.CatDesc = buf.String()
}

// FromUser form func
func (i *Item) FromUser(u *User, as map[int64]*api.ArcPlayer) {
	i.Title = u.Name
	i.Cover = u.Pic
	i.Goto = model.GotoAuthor
	i.OfficialVerify = u.OfficialVerify
	i.Param = strconv.Itoa(int(u.Mid))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Mid = u.Mid
	i.Sign = u.Usign
	i.Fans = u.Fans
	i.Level = u.Level
	i.Arcs = u.Videos
	i.AvItems = make([]*Item, 0, len(u.Res))
	i.RoomID = u.RoomID
	if u.IsUpuser == 1 {
		for _, v := range u.Res {
			vi := &Item{}
			vi.Title = v.Title
			vi.Cover = v.Pic
			vi.Goto = model.GotoAv
			vi.Param = strconv.Itoa(int(v.Aid))
			a, ok := as[v.Aid]
			if ok && a != nil && a.Arc != nil {
				firstPlay := a.PlayerInfo[a.DefaultPlayerCid]
				vi.URI = model.FillURI(vi.Goto, vi.Param, model.AvPlayHandlerGRPC(a.Arc, firstPlay, ""))
				vi.Play = int(a.Arc.Stat.View)
				vi.Danmaku = int(a.Arc.Stat.Danmaku)
			} else {
				switch play := v.Play.(type) {
				case float64:
					vi.Play = int(play)
				case string:
					vi.Play, _ = strconv.Atoi(play)
				}
				vi.Danmaku = v.Danmaku
			}
			vi.CTime = v.Pubdate
			vi.Duration = v.Duration
			i.AvItems = append(i.AvItems, vi)
		}
		i.IsUp = true
	}
}

// FromUpUser form func
func (i *Item) FromUpUser(u *User, as map[int64]*api.ArcPlayer) {
	i.Title = u.Name
	i.Cover = u.Pic
	i.Goto = model.GotoAuthor
	i.OfficialVerify = u.OfficialVerify
	i.Param = strconv.Itoa(int(u.Mid))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Mid = u.Mid
	i.Sign = u.Usign
	i.Fans = u.Fans
	i.Level = u.Level
	i.Arcs = u.Videos
	i.AvItems = make([]*Item, 0, len(u.Res))
	for _, v := range u.Res {
		vi := &Item{}
		vi.Title = v.Title
		vi.Cover = v.Pic
		vi.Goto = model.GotoAv
		vi.Param = strconv.Itoa(int(v.Aid))
		a, ok := as[v.Aid]
		if ok && a != nil && a.Arc != nil {
			firstPlay := a.PlayerInfo[a.DefaultPlayerCid]
			vi.URI = model.FillURI(vi.Goto, vi.Param, model.AvPlayHandlerGRPC(a.Arc, firstPlay, ""))
			vi.Play = int(a.Arc.Stat.View)
			vi.Danmaku = int(a.Arc.Stat.Danmaku)
			if a.Arc.Rights.UGCPay == 1 {
				vi.Badges = append(vi.Badges, model.PayBadge)
			}
			if a.Arc.Rights.IsCooperation == 1 {
				vi.Badges = append(vi.Badges, model.CooperationBadge)
			}
		} else {
			switch play := v.Play.(type) {
			case float64:
				vi.Play = int(play)
			case string:
				vi.Play, _ = strconv.Atoi(play)
			}
			vi.Danmaku = v.Danmaku
		}
		vi.CTime = v.Pubdate
		vi.Duration = v.Duration
		i.AvItems = append(i.AvItems, vi)
	}
	i.RoomID = u.RoomID
	i.IsUp = u.IsUpuser == 1
}

// FromMovie form func
func (i *Item) FromMovie(m *Movie, as map[int64]*api.Arc) {
	i.Title = m.Title
	i.Desc = m.Desc
	if m.Type == "movie" {
		i.Cover = m.Cover
		i.Param = strconv.Itoa(int(m.Aid))
		i.Goto = model.GotoAv
		i.URI = model.FillURI(i.Goto, i.Param, model.AvHandler(as[m.Aid], ""))
		i.CoverMark = model.StatusMark(m.Status)
	} else if m.Type == "special" {
		i.Param = m.SpID
		i.Goto = model.GotoSp
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Cover = m.Pic
	}
	i.Staff = m.Staff
	i.Actors = m.Actors
	i.Area = m.Area
	i.Length = m.Length
	i.Status = m.Status
	i.ScreenDate = m.ScreenDate
}

// FromVideo form func
func (i *Item) FromVideo(v *Video, a *api.ArcPlayer, cooperation bool) {
	i.Title = v.Title
	i.Cover = v.Pic
	i.Author = v.Author
	i.Param = strconv.Itoa(int(v.ID))
	i.Goto = model.GotoAv
	if a != nil && a.Arc != nil {
		i.Face = a.Arc.Author.Face
		firstPlay := a.PlayerInfo[a.DefaultPlayerCid]
		i.URI = model.FillURI(i.Goto, i.Param, model.AvPlayHandlerGRPC(a.Arc, firstPlay, ""))
		i.Play = int(a.Arc.Stat.View)
		i.Danmaku = int(a.Arc.Stat.Danmaku)
		if a.Arc.Rights.UGCPay == 1 {
			i.Badges = append(i.Badges, model.PayBadge)
		}
		if a.Arc.Rights.IsCooperation == 1 {
			i.Badges = append(i.Badges, model.CooperationBadge)
			if i.Author != "" && cooperation {
				i.Author += " 等联合创作"
			}
		}
	} else {
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		switch play := v.Play.(type) {
		case float64:
			i.Play = int(play)
		case string:
			i.Play, _ = strconv.Atoi(play)
		}
		i.Danmaku = v.Danmaku
	}
	i.Desc = v.Desc
	i.Duration = v.Duration
	i.ViewType = v.ViewType
	i.RecTags = v.RecTags
	for _, r := range v.NewRecTags {
		if r.Name != "" {
			switch r.Style {
			case model.BgStyleFill:
				vs := &model.ReasonStyle{}
				*vs = *videoStrongStyle
				vs.Text = r.Name
				i.NewRecTags = append(i.NewRecTags, vs)
			case model.BgStyleStroke:
				vw := &model.ReasonStyle{}
				*vw = *videoWeekStyle
				vw.Text = r.Name
				i.NewRecTags = append(i.NewRecTags, vw)
			}
		}
	}
}

// FromOperate form func
func (i *Item) FromOperate(o *Operate, gt string) {
	i.Title = o.Title
	i.Cover = o.Cover
	i.URI = o.RedirectURL
	i.Param = strconv.FormatInt(o.ID, 10)
	i.Desc = o.Desc
	i.Badge = o.Corner
	i.Goto = gt
	if o.RecReason != "" {
		i.RcmdReason = &RcmdReason{Content: o.RecReason}
	}
}

// FromConverge form func
// nolint:gomnd
func (i *Item) FromConverge(o *Operate, am map[int64]*api.Arc, artm map[int64]*article.Meta) {
	const _convergeMinCount = 2
	cis := make([]*Item, 0, len(o.ContentList))
	for _, c := range o.ContentList {
		ci := &Item{}
		switch c.Type {
		case 0:
			if a, ok := am[c.ID]; ok && a.IsNormal() {
				ci.Title = a.Title
				ci.Cover = a.Pic
				ci.Goto = model.GotoAv
				ci.Param = strconv.FormatInt(a.Aid, 10)
				ci.URI = model.FillURI(ci.Goto, ci.Param, model.AvHandler(a, ""))
				ci.fillArcStat(a)
				cis = append(cis, ci)
			}
		case 2:
			if art, ok := artm[c.ID]; ok {
				ci.Title = art.Title
				ci.Desc = art.Summary
				if len(art.ImageURLs) != 0 {
					ci.Cover = art.ImageURLs[0]
				}
				ci.Goto = model.GotoArticle
				ci.Param = strconv.FormatInt(art.ID, 10)
				ci.URI = model.FillURI(ci.Goto, ci.Param, nil)
				if art.Stats != nil {
					ci.fillArtStat(art)
				}
				ci.Badge = "文章"
				cis = append(cis, ci)
			}
		}
	}
	if len(cis) < _convergeMinCount {
		return
	}
	i.Item = cis
	i.Title = o.Title
	i.Cover = o.Cover
	i.URI = o.RedirectURL
	i.Param = strconv.FormatInt(o.ID, 10)
	if o.CardType == TypeConvergeContent {
		i.ContentURI = model.FillURI(model.GotoConvergeContent, i.Param, nil)
	}
	i.Desc = o.Desc
	i.Badge = o.Corner
	i.Goto = model.GotoConverge
	if o.RecReason != "" {
		i.RcmdReason = &RcmdReason{Content: o.RecReason}
	}
}

// FromMedia form func
func (i *Item) FromMedia(m *Media, prompt string, gt string, bangumis map[string]*bangumimdl.Card, medisas map[int32]*pgcsearch.SearchMediaProto) {
	i.Title = m.Title
	if i.Title == "" {
		i.Title = m.OrgTitle
	}
	i.Cover = m.Cover
	i.Goto = gt
	i.Param = strconv.Itoa(int(m.MediaID))
	i.URI = m.GotoURL
	i.MediaType = m.MediaType
	i.PlayState = m.PlayState
	i.Style = m.Styles
	i.CV = m.CV
	i.Staff = m.Staff
	if m.MediaScore != nil {
		i.Rating = m.MediaScore.Score
		i.Vote = m.MediaScore.UserCount
	}
	i.PTime = m.Pubtime
	areas := strings.Split(m.Areas, "、")
	if len(areas) != 0 {
		i.Area = areas[0]
	}
	i.Prompt = prompt
	var hit string
	for _, v := range m.HitColumns {
		if v == "cv" {
			hit = v
			break
		} else if v == "staff" {
			hit = v
		}
	}
	if hit == "cv" {
		for _, v := range getHightLight.FindAllStringSubmatch(m.CV, -1) {
			if gt == model.GotoBangumi {
				i.Label = fmt.Sprintf("声优: %v...", v[0])
				break
			} else if gt == model.GotoMovie {
				i.Label = fmt.Sprintf("演员: %v...", v[0])
				break
			}
		}
	} else if hit == "staff" {
		for _, v := range getHightLight.FindAllStringSubmatch(m.Staff, -1) {
			i.Label = fmt.Sprintf("制作人员: %v...", v[0])
			break
		}
	}
	// get from PGC API.
	i.SeasonID = m.SeasonID
	ssID := strconv.Itoa(int(m.SeasonID))
	if bgm, ok := bangumis[ssID]; ok {
		i.IsAttention = bgm.IsFollow
		i.IsSelection = bgm.IsSelection
		i.SeasonType = bgm.SeasonType
		i.Badges = bgm.Badges
		for _, v := range bgm.Episodes {
			tmp := &Item{
				Param: strconv.Itoa(int(v.ID)),
				Index: v.Index,
			}
			tmp.URI = model.FillURI(model.GotoEP, tmp.Param, nil)
			i.Episodes = append(i.Episodes, tmp)
		}
	}
}

// FromArticle form func
func (i *Item) FromArticle(a *Article) {
	i.ID = a.ID
	i.Mid = a.Mid
	i.TemplateID = a.TemplateID
	i.Title = a.Title
	i.Desc = a.Desc
	i.ImageUrls = a.ImageUrls
	i.View = a.View
	i.Play = a.View
	i.Like = a.Like
	i.Reply = a.Reply
	i.Goto = model.GotoArticle
	i.Param = strconv.Itoa(int(a.ID))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
}

// FromChannel form func
func (i *Item) FromChannel(c *Channel) {
	i.ID = c.TagID
	i.Title = c.TagName
	i.Cover = c.Cover
	i.Param = strconv.FormatInt(c.TagID, 10)
	i.Goto = model.GotoChannel
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Type = c.Type
	i.Attentions = c.AttenCount
}

// FromQuery form func
func (i *Item) FromQuery(qs []*Query) {
	i.Goto = model.GotoRecommendWord
	for _, q := range qs {
		i.List = append(i.List, &Item{Param: strconv.FormatInt(q.ID, 10), Title: q.Name, Type: q.Type, FromSource: q.FromSource})
	}
}

// FromTwitter form twitter
func (i *Item) FromTwitter(t *Twitter) {
	i.Title = t.Content
	i.Covers = t.Cover
	i.CoverCount = t.CoverCount
	i.Param = strconv.FormatInt(t.ID, 10)
	i.Goto = model.GotoTwitter
	i.URI = model.FillURI(i.Goto, strconv.FormatInt(t.PicID, 10), nil)
}

// fillArcStat fill func
func (i *Item) fillArcStat(a *api.Arc) {
	if a.Access == 0 {
		i.Play = int(a.Stat.View)
	}
	i.Danmaku = int(a.Stat.Danmaku)
	i.Reply = int(a.Stat.Reply)
	i.Like = int(a.Stat.Like)
}

// fillArtStat fill func
func (i *Item) fillArtStat(m *article.Meta) {
	i.Play = int(m.Stats.View)
	i.Reply = int(m.Stats.Reply)
}

// FromSuggest3 form func
func (i *Item) FromSuggest3(st *Sug, as map[int64]*api.Arc) {
	i.From = "search"
	i.Title = st.ShowName
	i.KeyWord = st.Term
	i.Position = st.Pos
	i.Cover = st.Cover
	i.CoverSize = st.CoverSize
	i.SugType = st.SubType
	i.TermType = st.TermType
	if st.TermType == SuggestionJump {
		switch st.SubType {
		case SuggestionAV:
			i.Goto = model.GotoAv
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), model.AvHandler(as[st.Ref], ""))
			i.SugType = "视频"
		case SuggestionArticle:
			i.Goto = model.GotoArticle
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), nil)
			if !strings.Contains(i.URI, "column_from") {
				i.URI += "?column_from=search"
			}
			i.SugType = "专栏"
		}
	} else if st.TermType == SuggestionJumpUser && st.User != nil {
		i.Title = st.User.Name
		i.Cover = st.User.Face
		i.Goto = model.GotoAuthor
		i.OfficialVerify = &OfficialVerify{Type: st.User.OfficialVerifyType}
		i.Param = strconv.Itoa(int(st.User.Mid))
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Mid = st.User.Mid
		i.Fans = st.User.Fans
		i.Level = st.User.Level
		i.Arcs = st.User.Videos
	} else if st.TermType == SuggestionJumpPGC && st.PGC != nil {
		var styles []string
		i.Title = st.PGC.Title
		i.Cover = st.PGC.Cover
		i.PTime = st.PGC.Pubtime
		i.URI = st.PGC.GotoURL
		if i.PTime != 0 {
			if pt := i.PTime.Time().Format("2006"); pt != "" {
				styles = append(styles, pt)
			}
		}
		i.SeasonTypeName = model.FormMediaType(st.PGC.MediaType)
		if i.SeasonTypeName != "" {
			styles = append(styles, i.SeasonTypeName)
		}
		i.Goto = model.GotoPGC
		i.Param = strconv.Itoa(int(st.PGC.MediaID))
		i.Area = st.PGC.Areas
		if i.Area != "" {
			styles = append(styles, i.Area)
		}
		i.Style = st.PGC.Styles
		if len(styles) > 0 {
			i.Styles = strings.Join(styles, " | ")
		}
		i.Label = FormPGCLabel(st.PGC.MediaType, st.PGC.Styles, st.PGC.Staff, st.PGC.CV)
		i.Rating = st.PGC.MediaScore
		i.Vote = st.PGC.MediaUserCount
		i.Badges = st.PGC.Badges
	}
}

// FormPGCLabel from pgc labe.
func FormPGCLabel(mediaType int, styles, staff, cv string) (label string) {
	switch mediaType {
	case model.MediaTypeBangumi: // 番剧
		label = strings.Replace(styles, "\n", "、", -1)
	case model.MediaTypeMovie: // 电影
		if cv != "" {
			label = "演员：" + strings.Replace(cv, "\n", "、", -1)
		}
	case model.MediaTypeDocumentary: // 纪录片
		label = strings.Replace(staff, "\n", "、", -1)
	case model.MediaTypeGuoChuang: // 国创
		label = strings.Replace(styles, "\n", "、", -1)
	case model.MediaTypeTvSeries: // 电视剧
		if cv != "" {
			label = "演员：" + strings.Replace(cv, "\n", "、", -1)
		}
	case model.MediaTypeShow: // 综艺
		label = strings.Replace(cv, "\n", "、", -1)
	//case 123: // 电视剧
	//	label = "演员：" + strings.Replace(cv, "\n", "、", -1)
	//case 124: // 综艺
	//	label = strings.Replace(cv, "\n", "、", -1)
	//case 125: // 纪录片
	//	label = strings.Replace(staff, "\n", "、", -1)
	//case 126: // 电影
	//	label = "演员：" + strings.Replace(cv, "\n", "、", -1)
	//case 127: // 动漫
	//	label = strings.Replace(styles, "\n", "、", -1)
	default:
		label = strings.Replace(cv, "\n", "、", -1)
	}
	return
}

// FromMediaPgcCard def.
func (i *Item) FromMediaPgcCard(m *Media, prompt string, gt string, bangumis map[string]*bangumimdl.Card, seasonEps map[int32]*pgcsearch.SearchCardProto, medisas map[int32]*pgcsearch.SearchMediaProto, cfg *conf.PgcSearchCard, isIpadDirect bool) { // isIpadDirect ipad垂搜
	i.FromMedia(m, prompt, gt, bangumis, medisas)
	i.SelectionStyle = _styleGrid // 默认宫格，当且仅当pgc新接口下发并且为横条才会出横条
	if m.IsAllNet() {             // 全网搜，下发搜索的out_url和立即观看
		i.WatchButton = &WatchButton{
			Title: cfg.OnlineWatch,
			Link:  m.AllNetURL,
		}
		i.IsOut = 1
		return
	}
	i.WatchButton = &WatchButton{ // 默认watch_button，当该season不可播时候用到
		Title: cfg.OfflineWatch,
		Link:  m.GotoURL,
	}
	if m.Canplay() {
		if seasonEp, ok := seasonEps[int32(m.SeasonID)]; ok { // 使用pgc下发的按钮和链接
			i.Styles = seasonEp.Styles
			i.PTime = xtime.Time(seasonEp.PubTime)
			isHorizon := seasonEp.SelectionStyle == _styleHorizontal
			i.WatchButton = &WatchButton{ // pgc下发 立即观看按钮
				Title: seasonEp.ButtonText,
				Link:  seasonEp.Url,
			}
			if seasonEp.Follow != nil { // pgc下发 追番/追剧按钮
				i.FollowButton = new(FollowButton)
				i.FollowButton.FromPGCCard(seasonEp.Follow)
				i.IsAttention = int(seasonEp.IsFollow)
			} else {
				log.Warn("FollowButton Sid %d Missing Follow", m.SeasonID)
			}
			i.SelectionStyle = seasonEp.SelectionStyle
			i.IsSelection = int(seasonEp.IsSelection) // when pgc gives is_selection in new grpc, use it to replace the old http's
			if len(seasonEp.Eps) == 0 {               // 无选集信息，不处理选集和查看更多
				return
			}
			var pos int
			for _, epGrpc := range seasonEp.Eps {
				if isHorizon && ((isIpadDirect && len(i.EpisodesNew) >= cfg.IpadEpSize) || (!isIpadDirect && len(i.EpisodesNew) >= cfg.Epsize)) { // ipad垂搜横条最多3条，ipad综合搜索和手机最多2条
					break
				}
				epNew := new(EpisodeNew)
				if canAppend := epNew.FromPgcRes(epGrpc, isHorizon, cfg.GridBadge); canAppend {
					if epNew.Type == 0 { // 0正常ep 1更多链接
						pos++
						epNew.Position = pos
					}
					i.EpisodesNew = append(i.EpisodesNew, epNew)
				}
			}
			if m.HitEpids == "" && isHorizon && ((isIpadDirect && len(seasonEp.Eps) > cfg.IpadEpSize) || (!isIpadDirect && len(seasonEp.Eps) > cfg.Epsize)) { // 未召回单集 && 横条 && 长度>2(phone), >3(ipad) 展示 "查看全部.."
				if isIpadDirect && len(i.EpisodesNew) > cfg.IpadCheckMoreSize { // ipad垂搜超过3条时候压缩为2条+查看更多
					i.EpisodesNew = i.EpisodesNew[0:cfg.IpadCheckMoreSize]
				}
				i.CheckMore = &CheckMore{
					Content: fmt.Sprintf(cfg.CheckMoreContent, seasonEp.EpSize),
					Uri:     fmt.Sprintf(cfg.CheckMoreSchema, _styleHorizontal, seasonEp.SeasonId), // must be horizontal
				}
			}
		} else { // pgc未下发，使用搜索的goto_url
			i.WatchButton = &WatchButton{
				Title: cfg.OnlineWatch,
				Link:  m.GotoURL,
			}
		}
	}
}

// FromPgcRes builds the episode_new structure
func (v *EpisodeNew) FromPgcRes(ep *pgcsearch.SearchEpProto, isHorizon, gridBadge bool) (canAppend bool) {
	if isHorizon && ep.Title == "" { // 横条且pgc数据为空，认为为非法数据
		return false
	}
	if ep.ReleaseDate == "" { // pgc日期为空时只下发标题
		v.Title = ep.Title
	} else { // 否则日期拼到标题前面
		v.Title = fmt.Sprintf("%s %s", ep.ReleaseDate, ep.Title)
	}
	v.Uri = ep.Url
	v.Param = fmt.Sprintf("%d", ep.Id)
	if isHorizon || gridBadge { // 综合搜索+分类搜索
		v.Badges = ep.Badges
	}
	v.Type = ep.Type
	return true
}
