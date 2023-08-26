package model

import (
	"fmt"
	"strconv"
	"strings"

	api "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"

	"go-common/library/log"
	xtime "go-common/library/time"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"
	arcMidV1 "go-gateway/app/app-svr/archive/middleware/v1"
)

const (
	// 收藏的歌单
	MenuFavored = 1
	// 创建的歌单
	MenuCreated = 2
	// 收藏的合辑
	CollectionFavored = 3
)

var MenuType2Name = map[int32]string{
	MenuFavored:       "收藏的歌单",
	MenuCreated:       "创建的歌单",
	CollectionFavored: "收藏的合辑",
}

const (
	SongAttrYes = 1
	// 过审
	SongStatusOK = 2

	// 128k
	SongQualityLow = 0
	// 192k
	SongQualityStd = 1
	// 320k
	SongQualityHigh = 2
)

type MusicMenu struct {
	MenuId       int64  `json:"menuId"`
	MenuType     int32  `json:"type"`
	Title        string `json:"title"`
	Desc         string `json:"intro"`
	CollectionId int64  `json:"collectionId"`
	ReplyNum     int64  `json:"commentNum"`
	PlayNum      int64  `json:"playNum"`
	ShareNum     int64  `json:"snum"`
	Total        int64  `json:"songNum"`
	Cover        string `json:"coverUrl"`
	FavoredNum   int64  `json:"collectNum"`
	IsFavored    int32  `json:"collected"`
	Ctime        int64  `json:"ctime"` // in ms
	Mid          int64  `json:"uid"`
	Username     string `json:"uname"`
	Avatar       string `json:"face"`
	IsDefault    int32  `json:"isDefault"`
	IsOff        int32  `json:"isOff"`
}

type MenuSongItem struct {
	Menu *api.Menu
	Song *api.Song
}

// ToMenuItem miss type
func (ms MenuSongItem) ToMenuItem(typ int32) MenuItem {
	data := ms.Menu
	menuItem := MenuItem{
		MenuId:   data.MenuId,
		MenuType: typ,
		Mid:      data.Mid,
		Title:    data.Title,
		Cover:    data.CoverUrl,
		Total:    data.Counts,
		// TODO IsOff, for what?
		IsOff: 0,
	}
	if !data.Private {
		menuItem.IsPublic = 1
	}
	if data.Default {
		menuItem.IsDefault = 1
	}
	return menuItem
}

// ToSongItem miss Author MaxQuality
func (ms MenuSongItem) ToSongItem() SongItem {
	s := ms.Song
	songItem := SongItem{
		SongId: s.SongId,
		Author: &v1.Author{
			Mid:      s.Mid,
			Relation: &v1.FollowRelation{},
		},
		Avid:     s.Aid,
		Cid:      s.Cid,
		Title:    s.Title,
		Desc:     s.Intro,
		Cover:    s.CoverUrl,
		Duration: s.Duration,
		Ctime:    xtime.Time(s.Ctime / 1000),
		Status:   int32(s.Status),
		Stat: &v1.BKStat{
			Like:      int32(s.LikeNum),
			Coin:      int32(s.CoinNum),
			Favourite: int32(s.CollectNum),
			Reply:     int32(s.RemarkNum),
			Share:     int32(s.ShareNum),
			View:      int32(s.PlayNum),
		},
		IsLocked: 0,
		IsOff:    int32(s.IsOff),
	}
	if s.IsCacheable {
		songItem.IsDownloadable = 1
	}
	return songItem
}

func (mm MusicMenu) ToMenuItem() MenuItem {
	return MenuItem{
		MenuId:    mm.MenuId,
		MenuType:  mm.MenuType,
		Title:     mm.Title,
		Desc:      mm.Desc,
		Total:     mm.Total,
		Cover:     mm.Cover,
		Ctime:     xtime.Time(mm.Ctime / 1000),
		Mid:       mm.Mid,
		Username:  mm.Username,
		Avatar:    mm.Avatar,
		PlayNum:   mm.PlayNum,
		ReplyNum:  mm.ReplyNum,
		IsDefault: mm.IsDefault,
		IsPublic:  1,
		IsOff:     mm.IsOff,
	}
}

type MusicCollection struct {
	CollectionId int64  `json:"collection_id"`
	MenuId       int64  `json:"menu_id"`
	Title        string `json:"title"`
	Desc         string `json:"desc"`
	Total        int64  `json:"records_num"`
	Cover        string `json:"img_url"`
	Ctime        int64  `json:"ctime"` // in ms
	Mid          int64  `json:"mid"`
	Username     string `json:"uname"`
	Avatar       string `json:"avatar"`
	IsDefault    int32  `json:"is_default"`
	IsPublic     int32  `json:"is_open"`
}

func (mc MusicCollection) ToMenuItem() MenuItem {
	return MenuItem{
		MenuId:    mc.MenuId,
		MenuType:  MenuCreated,
		Title:     mc.Title,
		Desc:      mc.Desc,
		Total:     mc.Total,
		Cover:     mc.Cover,
		Ctime:     xtime.Time(mc.Ctime / 1000),
		Mid:       mc.Mid,
		Username:  mc.Username,
		Avatar:    mc.Avatar,
		IsDefault: mc.IsDefault,
		IsPublic:  mc.IsPublic,
		IsOff:     0,
	}
}

type MenuItem struct {
	MenuId    int64
	MenuType  int32
	Title     string
	Desc      string
	Total     int64
	Cover     string
	Ctime     xtime.Time
	Mid       int64
	Username  string
	Avatar    string
	PlayNum   int64
	ReplyNum  int64
	IsDefault int32
	IsPublic  int32
	IsOff     int32
}

func (mi MenuItem) CalculateAttr() int64 {
	var attr int64
	if mi.IsDefault == 1 {
		attr |= menuAttrIsDefault
	}
	if mi.IsPublic == 1 {
		attr |= menuAttrIsPublic
	}
	return attr
}

func (mi MenuItem) CalculateState(mid int64) int32 {
	if mi.IsOff == 1 && mid != mi.Mid {
		return -1
	}
	if mid != mi.Mid && mi.IsPublic == 0 {
		return -1
	}
	return 0
}

const (
	menuAttrIsDefault = 0x1
	menuAttrIsPublic  = 0x2
)

func (mi MenuItem) ToV1MusicMenu(mid int64) *v1.MusicMenu {
	return &v1.MusicMenu{
		Id:       mi.MenuId,
		MenuType: mi.MenuType,
		Title:    mi.Title,
		Desc:     mi.Desc,
		Cover:    mi.Cover,
		Owner: &v1.MusicMenuAuthor{
			Mid:    mi.Mid,
			Name:   mi.Username,
			Avatar: mi.Avatar,
		},
		State: mi.CalculateState(mid),
		Attr:  mi.CalculateAttr(),
		Stat: &v1.MusicMenuStat{
			Play:  mi.PlayNum,
			Reply: mi.ReplyNum,
		},
		Total: mi.Total,
		Ctime: mi.Ctime,
		Uri:   fmt.Sprintf("bilibili://podcast/legacy?id=%d&extra_id=%d&source=2", mi.MenuId, mi.MenuType),
	}
}

type MenuDetail struct {
	Menu  MenuItem
	Songs []SongItem
}

func (md MenuDetail) ToV1PlayItems() []*v1.PlayItem {
	ret := make([]*v1.PlayItem, 0, len(md.Songs))
	for _, s := range md.Songs {
		ret = append(ret, &v1.PlayItem{
			ItemType: PlayItemAudio,
			Oid:      s.SongId,
		})
	}
	return ret
}

type SongItem struct {
	SongId         int64
	Avid           int64
	Cid            int64
	Title          string
	Desc           string
	Cover          string
	Duration       int64 // seconds
	Ctime          xtime.Time
	Status         int32 // 歌曲状态,-1 未获取转码URL -2 正常获取 -3 获取失败 -4 转码出错 默认0 审核中1 审核通过2 审核不通过3
	Author         *v1.Author
	MaxQuality     int32 // Qn中最高的清晰度
	Qn             []SongQn
	Stat           *v1.BKStat
	IsDownloadable int32
	IsLocked       int32
	IsOff          int32
}

func (si SongItem) IsNormal() bool {
	if si.Status != SongStatusOK || si.IsOff == SongAttrYes || si.IsLocked == SongAttrYes ||
		si.SongId == 0 {
		return false
	}
	return true
}

func (si SongItem) ToV1FormatDescriptions() []*v1.FormatDescription {
	ret := make([]*v1.FormatDescription, 0, len(si.Qn))
	for _, q := range si.Qn {
		fd := &v1.FormatDescription{
			Format:      "mp4",
			Description: q.Bps,
			DisplayDesc: q.Desc,
		}
		switch {
		case q.Typ <= SongQualityLow:
			fd.Quality = _480p
		case q.Typ <= SongQualityStd:
			fd.Quality = _720p
		case q.Typ <= SongQualityHigh:
			fd.Quality = _1080p
		default:
			fd.Quality = _1080pHighBitRate
		}
		ret = append(ret, fd)
	}
	return ret
}

func (si SongItem) ToV1PlayItem() *v1.PlayItem {
	return &v1.PlayItem{
		ItemType: PlayItemAudio,
		Oid:      si.SongId,
		Et: &v1.EventTracking{
			EntityType: playType2EntityType[PlayItemAudio],
			EntityId:   strconv.FormatInt(si.SongId, 10),
		},
	}
}

func (si SongItem) ToV1BKArchive() *v1.BKArchive {
	return &v1.BKArchive{
		Oid:          si.SongId,
		Title:        si.Title,
		Cover:        si.Cover,
		Desc:         si.Desc,
		Duration:     si.Duration,
		Rid:          0, // TODO: music rid
		Rname:        "音乐",
		Publish:      si.Ctime,
		DisplayedOid: "AU" + strconv.Itoa(int(si.SongId)),
		Copyright:    1,
		Rights: &v1.BKArcRights{
			NoReprint: 1,
		},
	}
}

func (si SongItem) ToV1DetailItem() *v1.DetailItem {
	ret := &v1.DetailItem{
		Item: si.ToV1PlayItem(),
		Arc:  si.ToV1BKArchive(),
		Parts: []*v1.BKArcPart{
			{
				Oid:      si.SongId,
				SubId:    si.SongId,
				Title:    "p1",
				Duration: si.Duration,
				Page:     1,
			},
		},
		Owner: si.Author,
		Stat:  si.Stat,
	}
	// 关联视频
	if si.Avid > 0 {
		ret.AssociatedItem = &v1.PlayItem{
			ItemType: PlayItemUGC,
			Oid:      si.Avid,
		}
		if si.Cid > 0 {
			ret.AssociatedItem.SubId = []int64{si.Cid}
		}
	}
	if !si.IsNormal() {
		ret.Arc = &v1.BKArchive{
			Oid: si.SongId, Title: si.Title,
		}
		ret.Parts = nil
		ret.Stat = nil
		ret.Playable = PlayableInvalid
		ret.Message = conf.C.Res.Text.MsgArchiveInvalid
		ret.AssociatedItem = nil
	}
	return ret
}

func (si SongItem) ToV1FavItemAuthor() *v1.FavItemAuthor {
	if si.Author == nil {
		return nil
	}
	return &v1.FavItemAuthor{
		Mid:  si.Author.Mid,
		Name: si.Author.Name,
	}
}

func (si SongItem) ToV1FavItemStat() *v1.FavItemStat {
	if si.Stat == nil {
		return nil
	}
	return &v1.FavItemStat{
		View:  si.Stat.View,
		Reply: si.Stat.Reply,
	}
}

//nolint:deadcode,varcheck
const (
	_360p             = 16
	_480p             = 32
	_720p             = 64
	_1080p            = 80
	_1080pHighBitRate = 112
	_4k               = 120
)

func (si SongItem) ChooseQuality(args *arcMidV1.PlayerArgs) (ret int32) {
	if len(si.Qn) == 1 {
		return si.Qn[0].Typ
	}

	switch {
	case args.Qn <= _480p:
		ret = SongQualityLow
	case args.Qn <= _720p:
		ret = SongQualityStd
	case args.Qn <= _1080p:
		ret = SongQualityHigh
	default:
		ret = SongQualityStd
	}
	if ret <= si.MaxQuality {
		return
	}
	return si.MaxQuality
}

type SongQn struct {
	Typ  int32 // 0-流畅128 1-标准192 2-高品质320
	Desc string
	Bps  string // 320kbit/s
	Size int64  // in bytes
}

// "/audio/music-service-c/menus/%d"
type SongInMenu struct {
	SongId       int64  `json:"song_id"`
	Avid         string `json:"avid"`             // av123456 关联的视频
	Cid          int64  `json:"cid"`              // avid有的时候也不一定有
	SourceType   int32  `json:"cr_type"`          // 1-音乐 2-有声节目 3-视频转音频
	CreationType int32  `json:"creation_type_id"` // 1-原创 2-翻唱
	Title        string `json:"title"`
	Desc         string `json:"intro"`
	Cover        string `json:"cover_url"`
	Duration     int64  `json:"duration"` // seconds
	Ctime        int64  `json:"ctime"`    // in ms
	Ptime        int64  `json:"ptime"`    // 过审时间 in ms
	Status       int32  `json:"status"`   // 歌曲状态,-1 未获取转码URL -2 正常获取 -3 获取失败 -4 转码出错 默认0 审核中1 审核通过2 审核不通过3
	//LyricUrl     string `json:"lyric_url"` // 这个接口不会出的
	Mid          int64  `json:"mid"`
	Username     string `json:"author"`
	UploaderName string `json:"uploader_name"` // 上传时的用户名
	Qn           []struct {
		Typ  int32  `json:"type"`
		Desc string `json:"desc"`
		Tag  string `json:"tag"`  // HQ/SQ
		Bps  string `json:"bps"`  // 320kbit/s
		Size int64  `json:"size"` // in bytes
	} `json:"qualities"`
	Category []struct {
		Id   int32  `json:"cateId"`
		Name string `json:"cateInfo"`
	} `json:"songCate"`
	CommentNum     int64  `json:"comment_num"`
	PlayNum        int64  `json:"play_num"`
	IsDownloadable int32  `json:"is_cacheable"`
	IsCooperate    int32  `json:"is_cooper"` // 合作稿件
	IsRankable     int32  `json:"is_indexable"`
	IsLocked       int32  `json:"is_lock"`
	IsOff          int32  `json:"is_off"`
	IsPGC          int32  `json:"is_pgc"`
	Limit          int32  `json:"limit"`      // 受限 ??
	LitmitDesc     string `json:"litmitdesc"` // 受限说明
}

func (sim SongInMenu) ToSongItem() SongItem {
	ret := SongItem{
		SongId:   sim.SongId,
		Title:    sim.Title,
		Desc:     sim.Desc,
		Cover:    sim.Cover,
		Duration: sim.Duration,
		Ctime:    xtime.Time(sim.Ctime / 1000),
		Status:   sim.Status,
		Author: &v1.Author{
			Mid:      sim.Mid,
			Name:     sim.Username,
			Relation: &v1.FollowRelation{},
		},
		Stat: &v1.BKStat{
			Reply: int32(sim.CommentNum),
			View:  int32(sim.PlayNum),
		},
		IsDownloadable: sim.IsDownloadable,
		IsLocked:       sim.IsLocked,
		IsOff:          sim.IsOff,
	}
	ret.Qn = make([]SongQn, 0, len(sim.Qn))
	for _, qn := range sim.Qn {
		ret.Qn = append(ret.Qn, SongQn{
			Typ:  qn.Typ,
			Desc: qn.Desc,
			Bps:  qn.Bps,
			Size: qn.Size,
		})
		if qn.Typ > ret.MaxQuality {
			ret.MaxQuality = qn.Typ
		}
	}
	return ret
}

// "/audio/music-service-c/songs/playing"
type SongInPlaying struct {
	SongId   int64  `json:"id"`
	Avid     string `json:"avid"` // av123456 关联的视频
	Title    string `json:"title"`
	Desc     string `json:"intro"`
	Cover    string `json:"cover_url"`
	Duration int64  `json:"duration"` // seconds
	Ctime    int64  `json:"ctime"`    // in ms
	//Status       int32  `json:"status"`   // 歌曲状态,-1 未获取转码URL -2 正常获取 -3 获取失败 -4 转码出错 默认0 审核中1 审核通过2 审核不通过3
	LyricUrl    string `json:"lyric_url"` // 这个接口才有
	Mid         int64  `json:"mid"`
	Username    string `json:"author"`
	Avatar      string `json:"up_img"`
	IsFollowing int32  `json:"up_is_follow"` // 是否关注up
	Fans        int64  `json:"fans"`         // up粉丝数量
	//UploaderName string `json:"uploader_name"` // 上传时的用户名
	Qn []struct {
		Typ  int32  `json:"type"`
		Desc string `json:"desc"`
		Tag  string `json:"tag"`  // HQ/SQ
		Bps  string `json:"bps"`  // 320kbit/s
		Size int64  `json:"size"` // in bytes
	} `json:"qualities"`
	CommentNum     int64  `json:"reply_count"`
	PlayNum        int64  `json:"play_count"`
	CoinNum        int64  `json:"coin_num"`
	CoinCeiling    int64  `json:"coinceiling"`
	ShareNum       int64  `json:"snum"`
	IsDownloadable bool   `json:"is_cacheable"`
	IsOff          int32  `json:"is_off"`
	IsFavored      int32  `json:"is_collect"`
	Limit          int32  `json:"limit"`      // 受限 ??
	LitmitDesc     string `json:"litmitdesc"` // 受限说明
}

func (sip SongInPlaying) ToSongItem() SongItem {
	ret := SongItem{
		SongId:   sip.SongId,
		Title:    sip.Title,
		Desc:     sip.Desc,
		Cover:    sip.Cover,
		Duration: sip.Duration,
		Ctime:    xtime.Time(sip.Ctime / 1000),
		Status:   SongStatusOK,
		Author: &v1.Author{
			Mid:      sip.Mid,
			Name:     sip.Username,
			Avatar:   sip.Avatar,
			Relation: &v1.FollowRelation{},
		},
		Stat: &v1.BKStat{
			Reply: int32(sip.CommentNum),
			Coin:  int32(sip.CoinNum),
			View:  int32(sip.PlayNum),
		},
		IsLocked: 0,
		IsOff:    sip.IsOff,
	}
	if sip.IsFollowing == 1 {
		ret.Author.Relation.Status = v1.FollowRelation_FOLLOWING
	}
	ret.Qn = make([]SongQn, 0, len(sip.Qn))
	for _, qn := range sip.Qn {
		ret.Qn = append(ret.Qn, SongQn{
			Typ:  qn.Typ,
			Desc: qn.Desc,
			Bps:  qn.Bps,
			Size: qn.Size,
		})
		if qn.Typ > ret.MaxQuality {
			ret.MaxQuality = qn.Typ
		}
	}
	if sip.IsDownloadable {
		ret.IsDownloadable = 1
	}
	if len(sip.Avid) > 2 && strings.HasPrefix(sip.Avid, "av") {
		if avid, err := strconv.ParseInt(sip.Avid[2:], 10, 64); err == nil {
			ret.Avid = avid
		}
	}
	return ret
}

type SongPlayingDetail struct {
	Song SongItem
	URL  SongUrl
}

func (spd SongPlayingDetail) CanPlay() (int32, string) {
	return PlayableYES, ""
}

func (spd SongPlayingDetail) ToV1PlayInfo(arg *arcMidV1.PlayerArgs) *v1.PlayInfo {
	// seconds -> milliseconds
	songLen := uint64(spd.Song.Duration * 1000)
	ret := &v1.PlayInfo{
		Qn:     uint32(arg.Qn),
		Format: "mp4",
		QnType: QnTypeMP4,
		Info: &v1.PlayInfo_PlayUrl{
			PlayUrl: &v1.PlayURL{
				Durl: spd.URL.ToV1ResponseURL(songLen),
			},
		},
		Fnver:        uint32(arg.Fnver),
		Fnval:        uint32(arg.Fnval),
		Formats:      spd.Song.ToV1FormatDescriptions(),
		VideoCodecid: 7,
		Length:       songLen,
	}
	return ret
}

// "/audio/music-service-c/url"
type SongUrl struct {
	CDNS    []string
	SongId  int64 `json:"sid"`
	Size    int64 `json:"size"`
	Timeout int64 `json:"timeout"`
}

func (su SongUrl) ToV1ResponseURL(length uint64) []*v1.ResponseUrl {
	ret := []*v1.ResponseUrl{
		{
			Order:  1,
			Length: length,
			Size_:  uint64(su.Size),
			Url:    su.CDNS[0],
		},
	}
	if len(su.CDNS) > 1 {
		for i := 1; i < len(su.CDNS); i++ {
			ret[0].BackupUrl = append(ret[0].BackupUrl, su.CDNS[i])
		}
	}
	return ret
}

// "/x/internal/v1/audio/songs/search/baseQuerySongInfo"
type SongInDetail struct {
	SongId       int64  `json:"songId"`
	Avid         string `json:"avid"` // 关联的视频avid  av123456
	Cid          int64  `json:"cid"`  // 关联的视频cid
	Title        string `json:"title"`
	Cover        string `json:"coverUrl"`
	SourceType   int32  `json:"crType"`
	CreationType int32  `json:"creationTypeId"`
	Status       int32  `json:"musicStatus"`
	Duration     int64  `json:"duration"`
	LyricUrl     string `json:"lyricUrl"`
	Ctime        int64  `json:"cts"`
	Mid          int64  `json:"mid"`
	Username     string `json:"midName"`
	Numbers      struct {
		Heat       int64 `json:"heat"`
		PlayNum    int64 `json:"playNum"`
		CommentNum int64 `json:"replyNum"`
		ShareNum   int64 `json:"shareNum"`
		FavoredNum int64 `json:"collectNum"`
		FansNum    int64 `json:"fansCount"`
	} `json:"numberAttribute"`
	IsDel          int32 `json:"isDel"`
	IsOff          int32 `json:"isOff"`
	IsPGC          int32 `json:"isPgc"`
	IsLocked       int32 `json:"isLocked"`
	IsDownloadable int32 `json:"isCacheble"`
	IsRankable     int32 `json:"isIndexable"`
}

//nolint:gomnd
func (sid SongInDetail) ToSongItem() SongItem {
	ret := SongItem{
		SongId:   sid.SongId,
		Cid:      sid.Cid,
		Title:    sid.Title,
		Cover:    sid.Cover,
		Duration: sid.Duration,
		Ctime:    xtime.Time(sid.Ctime / 1000),
		Status:   sid.Status,
		Author: &v1.Author{
			Mid:      sid.Mid,
			Name:     sid.Username,
			Relation: &v1.FollowRelation{},
		},
		Stat: &v1.BKStat{
			Reply:     int32(sid.Numbers.CommentNum),
			View:      int32(sid.Numbers.PlayNum),
			Favourite: int32(sid.Numbers.FavoredNum),
			Share:     int32(sid.Numbers.ShareNum),
		},
		IsDownloadable: sid.IsDownloadable,
		IsLocked:       sid.IsLocked,
		IsOff:          sid.IsOff,
	}
	if len(sid.Avid) > 2 {
		var err error
		ret.Avid, err = strconv.ParseInt(sid.Avid[2:], 10, 64)
		if err != nil {
			ret.Avid = 0
			log.Warn("invalid avid string for SongInDetail(%+v). Discarded", sid)
		}
	}
	return ret
}

//nolint:gomnd
func (sid SongInDetail) WriteSongItem(s *SongItem) {
	s.Duration = sid.Duration
	s.Status = sid.Status
	s.IsDownloadable = sid.IsDownloadable
	s.IsLocked = sid.IsLocked
	s.Stat.Favourite = int32(sid.Numbers.FavoredNum)
	s.Stat.Share = int32(sid.Numbers.ShareNum)
	if len(sid.Avid) > 2 {
		var err error
		s.Avid, err = strconv.ParseInt(sid.Avid[2:], 10, 64)
		if err != nil {
			s.Avid = 0
			log.Warn("invalid avid string for SongInDetail(%+v). Discarded", sid)
		}
	}
	s.Cid = sid.Cid
	s.IsOff = sid.IsOff
}

type SongInDynamic struct {
	SongId       int64  `json:"id"`
	Cover        string `json:"cover"`
	Title        string `json:"title"`
	Desc         string `json:"intro"`
	Mid          int64  `json:"upId"`
	Avatar       string `json:"upperAvatar"`
	Username     string `json:"upper"`
	UploaderName string `json:"author"`
	Ctime        int64  `json:"ctime"`
	CommentNum   int64  `json:"replyCnt"`
	PlayNum      int64  `json:"playCnt"`
}

func (sd SongInDynamic) ToSongItem() SongItem {
	return SongItem{
		SongId: sd.SongId,
		Title:  sd.Title,
		Desc:   sd.Desc,
		Cover:  sd.Cover,
		Ctime:  xtime.Time(sd.Ctime / 1000),
		Author: &v1.Author{
			Mid:      sd.Mid,
			Name:     sd.Username,
			Avatar:   sd.Avatar,
			Relation: &v1.FollowRelation{},
		},
		Stat: &v1.BKStat{
			Reply: int32(sd.CommentNum),
			View:  int32(sd.PlayNum),
		},
	}
}

type SongInSpace struct {
	SongId int64  `json:"auid"`
	Title  string `json:"title"`
	Cover  string `json:"coverUrl"`
	Mid    int64  `json:"uid"`
	Ctime  int64  `json:"ctime"`
}

func (sis SongInSpace) ToSongItem() SongItem {
	return SongItem{
		SongId: sis.SongId,
		Title:  sis.Title,
		Cover:  sis.Cover,
		Author: &v1.Author{
			Mid:      sis.Mid,
			Relation: &v1.FollowRelation{},
		},
		Stat:  &v1.BKStat{},
		Ctime: xtime.Time(sis.Ctime / 1000),
	}
}

type SongInSpaceV3 struct {
	SongId   int64  `json:"id"`
	Avid     int64  `json:"aid"`
	Title    string `json:"title"`
	Mid      int64  `json:"uid"`
	Author   string `json:"author"` // name
	Cover    string `json:"cover"`
	Ctime    int64  `json:"ctime"` // in seconds
	Duration int64  `json:"duration"`
	IsCoop   int32  `json:"is_cooper"`
	IsOff    int32  `json:"isOff"`
	PlayCnt  int32  `json:"play"`
	ReplyCnt int32  `json:"reply"`
}

func (sisv3 SongInSpaceV3) ToSongItem() SongItem {
	return SongItem{
		SongId: sisv3.SongId,
		Title:  sisv3.Title,
		Cover:  sisv3.Cover,
		Author: &v1.Author{
			Mid:      sisv3.Mid,
			Name:     sisv3.Author,
			Relation: &v1.FollowRelation{},
		},
		Stat: &v1.BKStat{
			View:  sisv3.PlayCnt,
			Reply: sisv3.ReplyCnt,
		},
		Ctime: xtime.Time(sisv3.Ctime),
		IsOff: sisv3.IsOff,
	}
}

type PersonalMenuStatus struct {
	// 是否有单曲收藏
	HasSong bool `json:"song"`
	// 是否有收藏的歌单
	HasMenu bool `json:"menu"`
	// 是否有创建的歌单
	HasMenuCreated bool `json:"menu_created"`
	// 是否有收藏的合辑
	HasCollection bool `json:"collection"`
	// 是否有收藏的专辑（废弃）
	HasPGCMenu bool `json:"pgc_menu"`
}

type MusicMenuList struct {
	Typ            int32
	CurrentPageNum int64
	List           []MenuItem
	HasMore        bool
	Total          int64
}

func (mml MusicMenuList) ToV1MainFavMusicMenuListResp(mid int64) *v1.MainFavMusicMenuListResp {
	ret := &v1.MainFavMusicMenuListResp{
		TabType: mml.Typ,
		HasMore: mml.HasMore,
	}
	if mml.HasMore {
		ret.Offset = strconv.FormatInt(mml.CurrentPageNum+1, 10)
	}
	ret.MenuList = make([]*v1.MusicMenu, 0, len(mml.List))
	for _, m := range mml.List {
		mo := m.ToV1MusicMenu(mid)
		if mo != nil {
			ret.MenuList = append(ret.MenuList, mo)
		}
	}
	return ret
}
