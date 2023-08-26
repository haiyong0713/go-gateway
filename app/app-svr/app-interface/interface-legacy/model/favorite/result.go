package favorite

import (
	"context"
	"strconv"
	"time"

	xtime "go-common/library/time"
	article "go-gateway/app/app-svr/app-interface/interface-legacy/dao/article/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	artm "go-gateway/app/app-svr/app-interface/interface-legacy/model/article"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/audio"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/bplus"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/sp"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/topic"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
)

// MyFavorite struct
type MyFavorite struct {
	Tab      *Tab     `json:"tab,omitempty"`
	Favorite *FavList `json:"favorite,omitempty"`
}

// Tab struct
type Tab struct {
	Fav     bool `json:"favorite"`
	Topic   bool `json:"topic"`
	Article bool `json:"article"`
	Clips   bool `json:"clips"`
	Albums  bool `json:"albums"`
	Specil  bool `json:"specil"`
	Cinema  bool `json:"cinema"`
	Audios  bool `json:"audios"`
	Menu    bool `json:"menu"`
	PGCMenu bool `json:"pgc_menu"`
	Ticket  bool `json:"ticket"`
	Product bool `json:"product"`
}

// FavList struct
type FavList struct {
	Count int        `json:"count"`
	Items []*FavItem `json:"items"`
}

// FavideoList struct
type FavideoList struct {
	Count int            `json:"count"`
	Items []*FavideoItem `json:"items"`
}

// TopicList struct
type TopicList struct {
	Count int          `json:"count"`
	Items []*TopicItem `json:"items"`
}

// ArticleList struct
type ArticleList struct {
	Count int            `json:"count"`
	Items []*ArticleItem `json:"items"`
}

// ClipsList struct
type ClipsList struct {
	*bplus.PageInfo
	Items []*ClipsItem `json:"items"`
}

// AlbumsList struct
type AlbumsList struct {
	*bplus.PageInfo
	Items []*AlbumItem `json:"items"`
}

// SpList struct
type SpList struct {
	Count int       `json:"count"`
	Items []*SpItem `json:"items"`
}

// AudioList struct
type AudioList struct {
	Count int          `json:"count"`
	Items []*AudioItem `json:"items"`
}

// FromFav is
func (i *FavItem) FromFav(f *Folder) {
	i.MediaID = f.MediaID
	i.Fid = f.Fid
	i.Mid = f.Mid
	i.Name = f.Name
	if f.Cover != nil {
		i.Cover = f.Cover
	}
	i.CurCount = f.CurCount
	i.State = f.State
}

// FavItem struct
type FavItem struct {
	MediaID  int64   `json:"media_id"`
	Fid      int     `json:"fid"`
	Mid      int     `json:"mid"`
	Name     string  `json:"name"`
	CurCount int     `json:"cur_count"`
	State    int     `json:"state"`
	Cover    []Cover `json:"cover"`
}

// FromFavideo is
func (i *FavideoItem) FromFavideo(fv *Archive) {
	i.Aid = fv.Aid
	i.Title = fv.Title
	i.Pic = fv.Pic
	i.Name = fv.Author.Name
	i.PlayNum = int(fv.Stat.View)
	i.Danmaku = int(fv.Stat.Danmaku)
	i.Param = strconv.FormatInt(int64(fv.Aid), 10)
	i.Goto = model.GotoAv
	i.URI = model.FillURI(i.Goto, i.Param, model.AvHandler(fv.Arc))
	if fv.Rights.ArcPay == 1 && fv.Rights.ArcPayFreeWatch == 0 {
		i.VideoPay = 1
	}
	if fv.IsNormal() {
		i.Valid = 1
	}
}

// FavideoItem struct
type FavideoItem struct {
	Aid     int64  `json:"aid"`
	Title   string `json:"title"`
	Pic     string `json:"pic"`
	Name    string `json:"name"`
	PlayNum int    `json:"play_num"`
	Danmaku int    `json:"danmaku"`
	Goto    string `json:"goto"`
	Param   string `json:"param"`
	URI     string `json:"uri"`
	UGCPay  int32  `json:"ugc_pay"`
	Valid   int    `json:"valid"`
	//视频是否付费 1：付费 0：免费
	VideoPay int `json:"video_pay"`
}

// FromTopic is
func (i *TopicItem) FromTopic(tp *topic.List) {
	i.ID = tp.ID
	i.MID = tp.MID
	i.Name = tp.Name
	i.PCCover = tp.PCCover
	i.H5Cover = tp.H5Cover
	i.FavAt = tp.FavAt
	i.PCUrl = tp.PCUrl
	i.H5Url = tp.H5Url
	i.Desc = tp.Desc
	i.Param = strconv.FormatInt(int64(tp.ID), 10)
	i.Goto = model.GotoWeb
	i.URI = model.FillURI(i.Goto, i.Param, nil)
}

// TopicItem struct
type TopicItem struct {
	ID      int64  `json:"id"`
	MID     int64  `json:"mid"`
	Name    string `json:"name"`
	PCCover string `json:"pc_cover"`
	H5Cover string `json:"h5_cover"`
	FavAt   int64  `json:"fav_at"`
	PCUrl   string `json:"pc_url"`
	H5Url   string `json:"h5_url"`
	Desc    string `json:"desc"`
	Goto    string `json:"goto"`
	Param   string `json:"param"`
	URI     string `json:"uri"`
}

// FromArticle is
func (i *ArticleItem) FromArticle(ctx context.Context, af *article.Favorite) {
	i.ID = af.ID
	i.Title = af.Title
	i.BannerURL = af.BannerURL
	i.TemplateID = int(af.TemplateID)
	if af.Author != nil {
		i.Name = af.Author.Name
		i.UpMid = af.Author.Mid
	}
	i.ImageURLs = af.ImageURLs
	i.Summary = af.Summary
	i.FTime = af.FavoriteTime
	i.Param = strconv.FormatInt(af.ID, 10)
	i.Goto = model.GotoArticle
	articleInfo := artm.GetArticleInfo(ctx, int64(af.Type), af.ID, af.CoverAvid)
	i.Badge = articleInfo.Badge
	i.URI = articleInfo.Uri
}

// ArticleItem struct
type ArticleItem struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	TemplateID int      `json:"template_id"`
	BannerURL  string   `json:"banner_url"`
	Name       string   `json:"name"`
	ImageURLs  []string `json:"image_urls"`
	Summary    string   `json:"summary"`
	FTime      int64    `json:"favorite_time"`
	Goto       string   `json:"goto"`
	Param      string   `json:"param"`
	URI        string   `json:"uri"`
	UpMid      int64    `json:"up_mid"`
	Badge      string   `json:"badge"`
}

// FromClips is
func (i *ClipsItem) FromClips(c *bplus.ClipList) {
	i.ID = c.Content.Item.ID
	i.Name = c.Content.User.Name
	i.UID = c.Content.User.UID
	i.HeadURL = c.Content.User.HeadURL
	i.IsVIP = c.Content.User.IsVIP
	i.IsFollowed = c.Content.User.IsFollowed
	i.UploadTimeText = c.Content.Item.UploadTimeText
	i.Tags = c.Content.Item.Tags
	i.Cover = c.Content.Item.Cover
	i.VideoTime = c.Content.Item.VideoTime
	i.Desc = c.Content.Item.Desc
	i.DanakuNum = c.Content.Item.DanakuNum
	i.WatchedNum = c.Content.Item.WatchedNum
	i.Param = strconv.FormatInt(int64(c.Content.Item.ID), 10)
	i.Goto = model.GotoClip
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Status = c.Content.Item.ShowStatus
	i.Reply = c.Content.Item.Reply
	i.UploadTime = c.Content.Item.UploadTime
	i.Width = c.Content.Item.Width
	i.Height = c.Content.Item.Height
	i.FirstPic = c.Content.Item.FirstPic
	i.VideoPlayURL = c.Content.Item.VideoPlayURL
	i.BackupPlayURL = c.Content.Item.BackupPlayURL
	i.LikeNum = c.Content.Item.LikeNum
}

// ClipsItem struct
type ClipsItem struct {
	ID             int64    `json:"id,omitempty"`
	Name           string   `json:"name,omitempty"`
	UID            int64    `json:"uid,omitempty"`
	HeadURL        string   `json:"head_url,omitempty"`
	IsVIP          int      `json:"is_vip,omitempty"`
	UploadTimeText string   `json:"upload_time_text,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Cover          struct {
		Def string `json:"default,omitempty"`
	} `json:"cover,omitempty"`
	VideoTime     int      `json:"video_time,omitempty"`
	Desc          string   `json:"description,omitempty"`
	DanakuNum     int      `json:"damaku_num,omitempty"`
	WatchedNum    int      `json:"watched_num,omitempty"`
	Goto          string   `json:"goto,omitempty"`
	Param         string   `json:"param,omitempty"`
	URI           string   `json:"uri,omitempty"`
	Status        int      `json:"status,omitempty"`
	Reply         int      `json:"reply,omitempty"`
	FirstPic      string   `json:"first_pic,omitempty"`
	BackupPlayURL []string `json:"backup_playurl,omitempty"`
	IsFollowed    bool     `json:"is_followed,omitempty"`
	UploadTime    string   `json:"upload_time,omitempty"`
	Width         int      `json:"width,omitempty"`
	Height        int      `json:"height,omitempty"`
	VideoPlayURL  string   `json:"video_playurl,omitempty"`
	LikeNum       int      `json:"like_num,omitempty"`
}

// FromAlbum is
func (i *AlbumItem) FromAlbum(bp *bplus.AlbumList) {
	i.ID = bp.Content.ID
	i.Pic = bp.Content.Pic
	i.PicCount = bp.Content.PicCount
	i.ShowStatus = bp.Content.ShowStatus
	i.Param = strconv.FormatInt(int64(bp.Content.ID), 10)
	i.Goto = model.GotoAlbum
	i.URI = model.FillURI(i.Goto, i.Param, nil)
}

// AlbumItem struct
type AlbumItem struct {
	ID         int64             `json:"id"`
	Pic        []*bplus.Pictures `json:"pictures"`
	ShowStatus int               `json:"show_status"`
	PicCount   int               `json:"pictures_count"`
	Goto       string            `json:"goto"`
	Param      string            `json:"param"`
	URI        string            `json:"uri"`
}

// FromSp is
func (i *SpItem) FromSp(s *sp.Item) {
	i.SpID = s.SpID
	i.Title = s.Title
	i.Cover = s.Cover
	i.MCover = s.MCover
	i.SCover = s.SCover
	timeTmp, _ := time.Parse("2006-01-02 15:04", s.CTime)
	i.CTime = timeTmp.Unix()
	i.Param = strconv.FormatInt(int64(s.SpID), 10)
	i.Goto = model.GotoSp
	i.URI = model.FillURI(i.Goto, i.Param, nil)
}

// SpItem struct
type SpItem struct {
	SpID   int64  `json:"spid"`
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	MCover string `json:"m_cover"`
	SCover string `json:"s_cover"`
	CTime  int64  `json:"create_at"`
	Goto   string `json:"goto"`
	Param  string `json:"param"`
	URI    string `json:"uri"`
}

// FromAudio is
func (i *AudioItem) FromAudio(a *audio.FavAudio) {
	i.ID = a.ID
	i.Title = a.Title
	i.IsOpen = a.IsOpen
	i.Cover = a.ImgURL
	i.Count = a.RecordsNum
}

// AudioItem struct
type AudioItem struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	IsOpen int    `json:"is_open"`
	Count  int    `json:"count"`
}

// TabItem struct
type TabItem struct {
	Name string `json:"name"`
	Uri  string `json:"uri"`
	Tab  string `json:"tab"`
}

// SecondReply .
type SecondReply struct {
	Items []*TabItem `json:"items"`
}

// TabParam struct
type TabParam struct {
	MobiApp       string `form:"mobi_app"`
	Device        string `form:"device"`
	Build         int    `form:"build"`
	Platform      string `form:"platform"`
	Mid           int64  `form:"mid"`
	Business      string `form:"business"`
	AccessKey     string `form:"access_key"`
	ActionKey     string `form:"actionKey"`
	Filtered      string `form:"filtered"`
	TeenagersMode int    `form:"teenagers_mode"`
	LessonsMode   int    `form:"lessons_mode"`
}

// FormFav from fav
func (i *Folder2) FormFav(f *favmdl.Folder) {
	//nolint:gomnd
	i.MediaID = f.ID*100 + f.Mid%100
	i.ID = f.ID
	i.Mid = f.Mid
	i.Title = f.Name
	i.Cover = f.Cover
	i.Count = int(f.Count)
	if !f.IsPublic() {
		i.IsPublic = 1
	}
	i.CTime = f.CTime
	i.MTime = f.MTime
	i.IsDefault = f.IsDefault()
}

type Favorites struct {
	Page struct {
		Num   int `json:"num"`
		Size  int `json:"size"`
		Count int `json:"count"`
	} `json:"page"`
	List []*Favorite `json:"list"`
}

type Favorite struct {
	ID       int64      `json:"id"`
	Oid      int64      `json:"oid"`
	Mid      int64      `json:"mid"`
	Fid      int64      `json:"fid"`
	Type     int8       `json:"type"`
	State    int8       `json:"state"`
	CTime    xtime.Time `json:"ctime"`
	MTime    xtime.Time `json:"mtime"`
	Sequence uint64     `json:"sequence"`
}

type ChannelFav struct {
	HasMore      bool          `json:"has_more"`
	Offest       string        `json:"offset"`
	ViewMoreLink string        `json:"view_more_link"`
	List         []*SubChannel `json:"list"`
	//频道订阅总数
	Total int32 `json:"total"`
}

type SubChannel struct {
	Cid           int64  `json:"cid"`
	Cname         string `json:"cname"`
	FeaturedCnt   int32  `json:"featured_cnt"`
	Icon          string `json:"icon"`
	SubscribedCnt int32  `json:"subscribed_cnt"`
	Url           string `json:"url"`
}
