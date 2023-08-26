package channel

import (
	"strconv"
	"time"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

// Tab is
type Tab struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	IsAtten    int32      `json:"is_atten,omitempty"`
	Atten      int64      `json:"atten,omitempty"`
	Cover      string     `json:"cover,omitempty"`
	HeadCover  string     `json:"head_cover,omitempty"`
	Content    string     `json:"content,omitempty"`
	URI        string     `json:"uri,omitempty"`
	Activity   int32      `json:"activity,omitempty"`
	SimilarTag []*Tab     `json:"similar_tag,omitempty"`
	TabList    []*TabList `json:"tab,omitempty"`
}

type TabList struct {
	TabID string `json:"tab_id,omitempty"`
	Name  string `json:"name,omitempty"`
	URI   string `json:"uri,omitempty,"`
}

// Tag is
type Tag struct {
	ID      int64  `json:"tag_id,omitempty"`
	Name    string `json:"tag_name,omitempty"`
	IsAtten int8   `json:"is_atten,omitempty"`
	Count   *struct {
		Atten int `json:"atten,omitempty"`
	} `json:"count,omitempty"`
}

// Param is
type Param struct {
	MobiApp   string `form:"mobi_app"`
	Device    string `form:"device"`
	AccessKey string `form:"access_key"`
	Build     int    `form:"build"`
	Ver       string `form:"ver"`
	Lang      string `form:"lang"`
	ID        int64  `form:"id"`
	MID       int64  `form:"mid"`
	Channel   string `form:"channel"`
}

// List is
type List struct {
	RegionTop    []*Region  `json:"region_top,omitempty"`
	RegionBottom []*Region  `json:"region_bottom,omitempty"`
	AttenChannel []*Channel `json:"atten_channel,omitempty"`
	RecChannel   []*Channel `json:"rec_channel,omitempty"`
	Ver          string     `json:"ver"`
}

// Region is
type Region struct {
	ID       int64  `json:"-"`
	RID      int    `json:"tid"`
	ReID     int    `json:"reid"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	GoTo     string `json:"goto"`
	Param    string `json:"param"`
	Type     int8   `json:"type"`
	URI      string `json:"uri,omitempty"`
	Area     string `json:"-"`
	Language string `json:"-"`
	Plat     int8   `json:"-"`
}

// Channel is
type Channel struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsAtten int32  `json:"is_atten,omitempty"`
	Cover   string `json:"cover,omitempty"`
	Atten   int64  `json:"atten,omitempty"`
	Content string `json:"content,omitempty"`
}

// Category is
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// RegionLimit
type RegionLimit struct {
	ID        int64  `json:"-"`
	Rid       int64  `json:"-"`
	Build     int    `json:"-"`
	Condition string `json:"-"`
}

// RegionConfig
type RegionConfig struct {
	ID       int64 `json:"-"`
	Rid      int64 `json:"-"`
	ScenesID int   `json:"-"`
}

// ParamSquare
type ParamSquare struct {
	MobiApp       string    `form:"mobi_app"`
	Device        string    `form:"device"`
	AccessKey     string    `form:"access_key"`
	Build         int       `form:"build"`
	Lang          string    `form:"lang"`
	MID           int64     `form:"mid"`
	LoginEvent    int32     `form:"login_event"`
	Time          time.Time `form:"time"`
	TeenagersMode int       `form:"teenagers_mode"`
	Channel       string    `form:"channel"`
}

// Square
type Square struct {
	Region []*Region      `json:"region,omitempty"`
	Square []card.Handler `json:"square,omitempty"`
}

// Mysub
type Mysub struct {
	List         []*Channel `json:"list,omitempty"`
	DisplayCount int        `json:"display_count,omitempty"`
}

type ChanOids struct {
	Oid      int64  `json:"-"`
	FromType string `json:"-"`
}

func (t *Tab) SimilarTagChange(tc *tag.ChannelReply, channels map[int64]*channelgrpc.Channel) {
	t.ID = tc.Channel.Id
	t.Name = tc.Channel.Name
	t.IsAtten = tc.Channel.Attention
	t.Atten = tc.Channel.Sub
	t.Content = tc.Channel.Content
	t.Cover = tc.Channel.Cover
	if t.Cover == "" {
		t.Cover = "http://i0.hdslb.com/bfs/archive/33dc521a84fb608e07770b3fdc347104aa6e9911.png"
	}
	t.HeadCover = tc.Channel.HeadCover
	if t.HeadCover == "" {
		t.HeadCover = "https://i0.hdslb.com/bfs/tag/30619e3f1cb06223a91f8eb3917876dad5da30f5.png"
	}
	t.Activity = tc.Channel.Activity
	for _, s := range tc.Synonyms {
		ct := &Tab{
			ID:   s.Id,
			Name: s.Name,
		}
		if channels != nil {
			if channel, ok := channels[s.Id]; ok && channel != nil {
				// 新频道
				if channel.CType == model.NewChannel {
					ct.URI = model.FillURI(model.GotoChannelNew, strconv.FormatInt(s.Id, 10), 0, 0, 0, model.ChannelHandler("tab=all&sort=hot"))
				}
			}
		}
		if ct.URI == "" {
			ct.URI = model.FillURI(model.GotoTag, strconv.FormatInt(s.Id, 10), 0, 0, 0, nil)
		}
		t.SimilarTag = append(t.SimilarTag, ct)
	}
}

func (l *TabList) TabListChange(m *tab.Menu) {
	l.TabID = strconv.FormatInt(m.TabID, 10)
	l.Name = m.Name
	l.URI = model.FillURI(model.GotoPegasusTab, strconv.FormatInt(m.TabID, 10), 0, 0, 0, model.PegasusHandler(m))
}
