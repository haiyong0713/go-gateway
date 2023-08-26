package model

import (
	"errors"
	"fmt"
	"strconv"

	xtime "go-common/library/time"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"

	favMdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favSvc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	ugcSeasonSvc "git.bilibili.co/bapis/bapis-go/ugc-season/service"
)

const (
	// 未知收藏类型
	FavTypeUnknown = 0
	// 主站视频收藏夹
	FavTypeVideo = 2
	// 我的收藏与订阅
	FavTypeMediaList = 11
	// UGC合集
	FavTypeUgcSeason = 21
	// OGV，对应oid为epid
	FavTypeOgv = 24
	// 音频，对应oid为song id
	FavTypeAudio = 12
)

// 对于PlayItem，需要对其item type进行映射
var Play2Fav = map[int32]int32{
	PlayItemUnknown: FavTypeUnknown,
	PlayItemUGC:     FavTypeVideo,
	PlayItemOGV:     FavTypeOgv,
	PlayItemAudio:   FavTypeAudio,
}

var Fav2Play = map[int32]int32{
	FavTypeUnknown: PlayItemUnknown,
	FavTypeVideo:   PlayItemUGC,
	FavTypeAudio:   PlayItemAudio,
	FavTypeOgv:     PlayItemOGV,
}

var (
	ErrAnchorNotFound = errors.New("anchor item not found")
)

//nolint:gomnd
func CalculateIdx(idx int32, length int32, size int32) (head, tail int32) {
	if size >= length {
		return 0, length
	}
	// 前1/4 后3/4
	partBefore := size / 4
	partAfter := size / 4 * 3
	head = idx - partBefore
	if head < 0 {
		tail = idx + partAfter - head
		head = 0
	} else {
		tail = idx + partAfter
		if tail > length {
			tail = length
			head = tail - size
		}
	}
	return
}

type FavItemMeta struct {
	Otype int32
	Oid   int64
}

type FavFolderMeta struct {
	Typ int32
	Mid int64
	Fid int64
}

type FavItemAddAndDelMeta struct {
	Tp    int32
	Mid   int64
	Fid   int64
	Oid   int64
	Otype int32
}

// mlid = fid * 100 + mid % 100
func (fm FavFolderMeta) EncodeFolderID() int64 {
	switch fm.Typ {
	case FavTypeUgcSeason:
		return fm.Fid
	default:
		return fm.Fid*100 + fm.Mid%100
	}
}

func (fm FavFolderMeta) Hash() string {
	switch fm.Typ {
	case FavTypeUgcSeason:
		return fmt.Sprintf("%d-%d", fm.Typ, fm.Fid)
	default:
		return fmt.Sprintf("%d-%d", fm.Typ, fm.Fid*100+fm.Mid%100)
	}
}

func HashV1FavFolder(f *v1.FavFolder) string {
	return fmt.Sprintf("%d-%d", f.FolderType, f.Fid)
}

type FavFolder struct {
	CurrentMid int64
	*favMdl.Folder
	UgcSeason *ugcSeasonSvc.Season
	OwnerInfo MemberInfo
}

const (
	FFAttrVisibilityPublic      = 0x0
	FFAttrVisibilityPrivate     = 0x1
	FFAttrFolderTypeDefault     = 0x0
	FFAttrFolderTypeUserCreated = 0x2
)

func (f FavFolder) GetCover() string {
	if f.UgcSeason != nil {
		return f.UgcSeason.GetCover()
	}
	return f.Folder.Cover
}

func (f FavFolder) GetOwnerMid() int64 {
	if f.UgcSeason != nil {
		return f.UgcSeason.Mid
	}
	return f.Folder.Mid
}

func (f FavFolder) GetRecentRes() []*favMdl.Resource {
	if f.Folder != nil {
		return f.Folder.RecentRes
	}
	return nil
}

func (f FavFolder) IsPrivate() bool {
	if f.Folder == nil {
		return false
	}
	return !f.IsPublic()
}

// 是否为自己的 创建的
func (f FavFolder) IsOwned() bool {
	return f.CurrentMid == f.OwnerInfo.GetMid()
}

// state = 0 是正常
// state = 1 删除
func (f FavFolder) IsNormal() bool {
	if f.Folder == nil {
		return true
	}
	return f.State == 0
}

func (f FavFolder) Hash() string {
	if f.UgcSeason != nil {
		return fmt.Sprintf("%d-%d", FavTypeUgcSeason, f.UgcSeason.ID)
	}
	return fmt.Sprintf("%d-%d", f.Type, f.Mlid)
}

func (f FavFolder) ToV1FavFolder() *v1.FavFolder {
	// 删除｜ 别人私有
	if !f.IsNormal() || (!f.IsOwned() && f.IsPrivate()) {
		return &v1.FavFolder{
			Fid:        f.Mlid,
			FolderType: f.Type,
			Owner:      f.OwnerInfo.ToV1FavFolderAuthor(),
			Name:       "收藏夹已失效",
			Attr:       f.Attr,
			Favored:    f.Favored,
			Ctime:      f.CTime,
			Mtime:      f.MTime,
			State:      -1,
		}
	}

	if f.UgcSeason != nil {
		return &v1.FavFolder{
			Fid:          f.UgcSeason.ID,
			FolderType:   FavTypeUgcSeason,
			Owner:        f.OwnerInfo.ToV1FavFolderAuthor(),
			Name:         f.UgcSeason.Title,
			Cover:        f.UgcSeason.Cover,
			Desc:         f.UgcSeason.Intro,
			Count:        int32(f.UgcSeason.EpCount),
			Attr:         FFAttrVisibilityPublic | FFAttrFolderTypeUserCreated,
			Favored:      1,
			Mtime:        f.UgcSeason.Ptime,
			StatFavCnt:   f.UgcSeason.Stat.Fav,
			StatShareCnt: f.UgcSeason.Stat.Share,
			StatLikeCnt:  f.UgcSeason.Stat.Like,
			StatPlayCnt:  f.UgcSeason.Stat.View,
			StatReplyCnt: f.UgcSeason.Stat.Reply,
		}
	}

	return &v1.FavFolder{
		Fid:          f.Mlid,
		FolderType:   f.Type,
		Owner:        f.OwnerInfo.ToV1FavFolderAuthor(),
		Name:         f.Name,
		Cover:        f.Cover,
		Desc:         f.Description,
		Count:        f.Count,
		Attr:         f.Attr,
		Favored:      f.Favored,
		Ctime:        f.CTime,
		Mtime:        f.MTime,
		StatFavCnt:   f.FavedCount,
		StatShareCnt: f.ShareCount,
		StatLikeCnt:  f.LikeCount,
		StatPlayCnt:  f.PlayCount,
		StatReplyCnt: f.ReplyCount,
	}
}

type FavItemDetail struct {
	Item      *favSvc.ModelFavorite // fav svc model item
	FavFolder *FavFolder            // 收藏与订阅
	// 以下是给正常稿件使用的字段
	AuthorMid         int64
	AuthorName        string
	ViewCnt, ReplyCnt int32
	Title, Cover      string
	Duration          int64
	State             int32
	Message           string
}

func (fd FavItemDetail) ToV1PlayItem() *v1.PlayItem {
	switch fd.Item.Type {
	case FavTypeVideo:
		return &v1.PlayItem{
			ItemType: PlayItemUGC,
			Oid:      fd.Item.Oid,
			Et: &v1.EventTracking{
				EntityType: playType2EntityType[PlayItemUGC],
				EntityId:   strconv.FormatInt(fd.Item.Oid, 10),
			},
		}
	case FavTypeOgv:
		return nil
	case FavTypeAudio:
		return &v1.PlayItem{
			ItemType: PlayItemAudio,
			Oid:      fd.Item.Oid,
			Et: &v1.EventTracking{
				EntityType: playType2EntityType[PlayItemAudio],
				EntityId:   strconv.FormatInt(fd.Item.Oid, 10),
			},
		}
	default:
		return nil
	}
}

func (fd FavItemDetail) EqualsPlayItem(p *v1.PlayItem) bool {
	if p == nil {
		return false
	}
	oid := p.Oid
	// 特殊处理: ogv 类型对比epid
	if p.ItemType == PlayItemOGV {
		if len(p.SubId) <= 0 {
			return false
		}
		oid = p.SubId[0]
	}
	return Play2Fav[p.ItemType] == fd.Item.Type && oid == fd.Item.Oid
}

func (fd FavItemDetail) ToV1FavItem(fmeta FavFolderMeta) *v1.FavItem {
	ret := &v1.FavItem{
		ItemType: Fav2Play[fd.Item.Type],
		Oid:      fd.Item.Oid,
		Fid:      fmeta.EncodeFolderID(),
		Mid:      fd.Item.Mid,
		Mtime:    xtime.Time(fd.Item.Mtime),
		Ctime:    xtime.Time(fd.Item.Ctime),
		Et: &v1.EventTracking{
			Batch:      strconv.FormatInt(int64(fmeta.Typ), 10),
			TrackId:    strconv.FormatInt(fmeta.EncodeFolderID(), 10),
			EntityType: playType2EntityType[Fav2Play[fd.Item.Type]],
			EntityId:   strconv.FormatInt(fd.Item.Oid, 10),
		},
	}
	if fmeta.Typ == FavTypeUgcSeason {
		ret.Et.Batch = ""
		ret.SetEventTracking(v1.OpUGCSeason)
	} else {
		ret.SetEventTracking(v1.OpFavorite)
	}
	return ret
}

func (fd FavItemDetail) ToV1FavFolder() (ret *v1.FavFolder) {
	if fd.FavFolder != nil {
		ret = fd.FavFolder.ToV1FavFolder()
		ret.Favored = 1
	} else {
		ret = &v1.FavFolder{
			Fid:        fd.Item.Oid,
			FolderType: fd.Item.Type,
			Name:       "UP主删除",
			Attr:       FFAttrVisibilityPublic | FFAttrFolderTypeUserCreated,
			State:      -1,
			Favored:    1,
			Ctime:      xtime.Time(fd.Item.Ctime),
			Mtime:      xtime.Time(fd.Item.Mtime),
		}
	}
	return
}

type FillV1FavItemOpt struct {
	//SeasonDetails map[int32]SeasonDetail
	//EpDetails     map[int32]EpisodeDetail
	ArchiveInfos map[int64]ArchiveInfo
	AudioInfos   map[int64]SongItem
	FilterAvs    map[int64]string
	EpCards      map[int32]EpCard
}

func (ffio FillV1FavItemOpt) filterArchive(fd FavItemDetail) (valid bool, filtered bool, msg string) {
	var aid int64
	// 先默认标记有效
	valid = true
	switch fd.Item.Type {
	case FavTypeVideo:
		aid = fd.Item.Oid
	case FavTypeOgv:
		if ffio.EpCards != nil {
			if ec, ok := ffio.EpCards[int32(fd.Item.Oid)]; ok {
				aid = ec.Ec.GetAid()
			} else {
				valid = false
			}
		} else {
			valid = false
		}
	}
	if aid > 0 {
		if ffio.ArchiveInfos == nil {
			return false, true, conf.C.Res.Text.MsgArchiveInvalid
		}
		if arcInfo := ffio.ArchiveInfos[aid]; arcInfo.Arc == nil {
			return false, true, conf.C.Res.Text.MsgArchiveInvalid
		} else if arcInfo.Arc.GetState() < 0 {
			return false, true, conf.C.Res.Text.MsgArchiveInvalid
		}
		if ffio.FilterAvs != nil {
			if msg := ffio.FilterAvs[aid]; len(msg) > 0 {
				return true, true, msg
			}
		}
	}
	return
}

func (ffio FillV1FavItemOpt) getV1FavItemElems(fd FavItemDetail) (arcInfo ArchiveInfo, epDetail EpCard, au *v1.FavItemAuthor, stat *v1.FavItemStat) {
	switch fd.Item.Type {
	case FavTypeOgv:
		if ffio.EpCards != nil && ffio.ArchiveInfos != nil {
			epDetail = ffio.EpCards[int32(fd.Item.Oid)]
			arcInfo = ffio.ArchiveInfos[epDetail.Ec.GetAid()]
		}
	case FavTypeVideo:
		if ffio.ArchiveInfos != nil {
			arcInfo = ffio.ArchiveInfos[fd.Item.Oid]
		}
	}
	if arcInfo.Arc == nil && epDetail.Ec == nil {
		return ArchiveInfo{}, EpCard{}, nil, nil
	}
	return arcInfo, epDetail, arcInfo.ToV1FavItemAuthor(), arcInfo.ToV1FavItemStat()
}

func (fd FavItemDetail) FillV1FavItemDetail(fmeta FavFolderMeta, opt FillV1FavItemOpt) *v1.FavItemDetail {
	ret := new(v1.FavItemDetail)
	ret.Item = fd.ToV1FavItem(fmeta)
	switch fd.Item.Type {
	case FavTypeOgv, FavTypeVideo:
		if (opt.EpCards == nil && fd.Item.Type == FavTypeOgv) || opt.ArchiveInfos == nil || opt.FilterAvs == nil {
			return nil
		}
		// 过滤视频稿件
		valid, filter, msg := opt.filterArchive(fd)
		var arc ArchiveInfo
		var ec EpCard
		arc, ec, ret.Owner, ret.Stat = opt.getV1FavItemElems(fd)
		if valid {
			if filter {
				ret.State, ret.Message = PlayableNO, msg
			}
			if fd.Item.Type == FavTypeOgv {
				ret.Duration, ret.Cover, ret.Parts = ec.Ec.GetDuration(), ec.Ec.GetCover(), 1 // OGV EP稿件默认按单p处理
				ret.Name = ec.Ec.GetSeason().GetTitle() + " " + ec.Ec.GetMeta().GetShortLongTitle()
			} else {
				ret.Name, ret.Cover = arc.Arc.Title, arc.Arc.Pic
				ret.Duration = arc.Arc.Duration
				ret.Parts = int32(arc.Arc.Videos)
			}
		} else {
			ret.Stat = nil // 保证抹掉stat信息
			ret.State, ret.Message = PlayableInvalid, msg
		}
	case FavTypeAudio:
		if opt.AudioInfos == nil {
			return nil
		}
		m := opt.AudioInfos[fd.Item.Oid]
		ret.Owner = m.ToV1FavItemAuthor()
		if !m.IsNormal() {
			ret.State = PlayableInvalid
			ret.Message = conf.C.Res.Text.MsgArchiveInvalid
		} else {
			ret.Name, ret.Cover = m.Title, m.Cover
			ret.Stat = m.ToV1FavItemStat()
			ret.Duration, ret.Parts = m.Duration, 1
		}
	default:
		return nil
	}

	return ret
}
