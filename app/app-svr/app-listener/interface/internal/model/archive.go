package model

import (
	"fmt"
	"strconv"

	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"
	arcSvc "go-gateway/app/app-svr/archive/service/api"
	purlSvcV2 "go-gateway/app/app-svr/playurl/service/api/v2"
)

const (
	// 未知稿件类型
	PlayItemUnknown int32 = 0
	// UGC稿件类型
	PlayItemUGC int32 = 1
	// OGV稿件类型
	PlayItemOGV int32 = 2
	//  音频类型
	PlayItemAudio int32 = 3
)

var playType2EntityType = map[int32]string{
	PlayItemUGC:   "av",
	PlayItemOGV:   "ep",
	PlayItemAudio: "au",
}

const (
	ArchiveStateDeletedByUp = -100
)

type (
	// 稿件详情
	ArchiveDetail struct {
		AC    *ArcControl // 播控
		Arc   *arcSvc.Arc
		Pages []*arcSvc.Page
		// up主信息
		UpInfo *v1.Author
		// 稿件数据
		Stat   *v1.BKStat
		Bvid   string
		Season *ArcUGCSeason
	}

	// UGC合集信息
	ArcUGCSeason struct {
		FavFolder
	}

	// 稿件播控信息
	ArcControl struct {
		CopyRightBan bool // 版权禁播
	}

	// 简单稿件信息
	ArchiveInfo struct {
		Arc *arcSvc.Arc
	}
)

const (
	PlayableDeletedByUp  = -2 // 被up主删除
	PlayableInvalid      = -1
	PlayableYES          = 0
	PlayableNO           = 1
	PlayableCopyrightBan = 2
)

func (ad ArchiveDetail) IsSteinsGate() bool {
	return ad.Arc.AttrVal(arcSvc.AttrBitSteinsGate) == arcSvc.AttrYes
}

func (ad ArchiveDetail) ToV1PlayItem(typ int32) *v1.PlayItem {
	return &v1.PlayItem{
		Oid: ad.Arc.Aid, ItemType: typ,
		Et: &v1.EventTracking{
			EntityType: playType2EntityType[typ],
			EntityId:   strconv.FormatInt(ad.Arc.Aid, 10),
		},
	}
}

func (ad ArchiveDetail) ToV1DetailItem(typ int32) *v1.DetailItem {
	var item *v1.DetailItem
	switch {
	// State >=0 用户才可见
	case ad.Arc.State < 0:
		item = &v1.DetailItem{
			Item: ad.ToV1PlayItem(typ),
			Arc: &v1.BKArchive{
				Oid:   ad.Arc.Aid,
				Title: ad.Arc.Title,
			},
			Owner:    ad.UpInfo,
			Playable: PlayableInvalid,
			Message:  conf.C.Res.Text.MsgArchiveInvalid,
		}
		if ad.Arc.State == ArchiveStateDeletedByUp {
			item.Playable = PlayableDeletedByUp
		}
	case ad.IsSteinsGate():
		item = &v1.DetailItem{
			Item:     ad.ToV1PlayItem(typ),
			Arc:      ad.toV1BKArchive(),
			Owner:    ad.UpInfo,
			Stat:     ad.Stat,
			Playable: PlayableNO,
			Message:  conf.C.Res.Text.MsgUnsupportedSteinsGate,
		}
	default:
		item = &v1.DetailItem{
			Item:          ad.ToV1PlayItem(typ),
			Arc:           ad.toV1BKArchive(),
			Owner:         ad.UpInfo,
			Stat:          ad.Stat,
			UgcSeasonInfo: ad.toV1UGCSeasonInfo(),
		}
		if ad.AC.CopyRightBan {
			item.Playable = PlayableCopyrightBan
			item.Message = conf.C.Res.Text.MsgCopyrightBanPlay
			// 版权稿件也不填充分p
		} else {
			for _, p := range ad.Pages {
				item.Parts = append(item.Parts, &v1.BKArcPart{
					Oid:      ad.Arc.Aid,
					SubId:    p.Cid,
					Title:    p.Part,
					Page:     p.Page,
					Duration: p.Duration,
				})
			}
		}
	}
	return item
}

func (ad ArchiveDetail) toV1BKArchive() *v1.BKArchive {
	return &v1.BKArchive{
		Oid:          ad.Arc.Aid,
		DisplayedOid: ad.Bvid,
		Title:        ad.Arc.Title,
		Cover:        ad.Arc.Pic,
		Desc:         ad.Arc.Desc,
		Duration:     ad.Arc.Duration,
		Rid:          ad.Arc.TypeID,
		Rname:        ad.Arc.TypeName,
		Publish:      ad.Arc.PubDate,
		Copyright:    ad.Arc.Copyright,
		Rights: &v1.BKArcRights{
			NoReprint: ad.Arc.Rights.NoReprint,
		},
	}
}

func (ad ArchiveDetail) toV1UGCSeasonInfo() *v1.FavFolder {
	if ad.Season == nil || ad.Season.UgcSeason == nil {
		return nil
	}
	return ad.Season.ToV1FavFolder()
}

func (ad ArchiveDetail) UGCSeasonMetaHash() string {
	if ad.Arc.GetSeasonID() == 0 {
		return ""
	}
	return fmt.Sprintf("%d-%d", FavTypeUgcSeason, ad.Arc.GetSeasonID())
}

// 是否是互动视频
func (ai ArchiveInfo) IsSteinsGate() bool {
	return ai.Arc.AttrVal(arcSvc.AttrBitSteinsGate) == arcSvc.AttrYes
}

func (ai ArchiveInfo) ToV1FavItemAuthor() *v1.FavItemAuthor {
	return &v1.FavItemAuthor{
		Mid:  ai.Arc.GetAuthor().Mid,
		Name: ai.Arc.GetAuthor().Name,
	}
}

func (ai ArchiveInfo) ToV1FavItemStat() *v1.FavItemStat {
	return &v1.FavItemStat{
		View:  ai.Arc.GetStat().View,
		Reply: ai.Arc.GetStat().Reply,
	}
}

func (ai ArchiveInfo) ToV1PickArchive(pickID, cardID int64) *v1.PickArchive {
	ret := &v1.PickArchive{
		Item: &v1.PlayItem{
			ItemType: PlayItemUGC,
			Oid:      ai.Arc.Aid,
			Et: &v1.EventTracking{
				Batch:      strconv.FormatInt(pickID, 10),
				TrackId:    strconv.FormatInt(cardID, 10),
				EntityType: playType2EntityType[PlayItemUGC],
				EntityId:   strconv.FormatInt(ai.Arc.Aid, 10),
			}},
		Owner: &v1.PickArchiveAuthor{Mid: ai.Arc.Author.Mid, Name: ai.Arc.Author.Name},
	}
	ret.Item.SetEventTracking(v1.OpFinding)
	if ai.Arc.State < 0 {
		ret.Title = conf.C.Res.Text.MsgArchiveInvalid
		ret.State = -1
		ret.Message = conf.C.Res.Text.MsgArchiveInvalid
		return ret
	}
	ret.Parts = int32(ai.Arc.Videos)
	ret.Cover = ai.Arc.Pic
	ret.Duration = ai.Arc.Duration
	ret.Title = ai.Arc.Title
	ret.StatView = ai.Arc.Stat.View
	ret.StatReply = ai.Arc.Stat.Reply
	return ret
}

type (
	PlayUrlInfo struct {
		Arc          *arcSvc.SimpleArc
		CopyrightBan bool // 版权禁播
		Cid          int64
		Qn           uint32
		Format       string
		QnType       int32
		PlayInfo     interface{}
		PlayUrl      *v1.PlayInfo_PlayUrl
		PlayDash     *v1.PlayInfo_PlayDash
		Volume       *purlSvcV2.VolumeInfo
		FnVer, FnVal uint32
		Formats      []*v1.FormatDescription
		VideoCodecID uint32
		Length       uint64
		Code         uint32
		Message      string
	}
)

func (pu PlayUrlInfo) CanPlay() (int32, string) {
	switch {
	case pu.Arc.Attribute>>arcSvc.AttrBitSteinsGate&int32(1) == arcSvc.AttrYes:
		return PlayableNO, conf.C.Res.Text.MsgUnsupportedSteinsGate
	case pu.CopyrightBan:
		return PlayableCopyrightBan, conf.C.Res.Text.MsgCopyrightBanPlay
	}
	return PlayableYES, ""
}

const (
	QnTypeFLV  = 1
	QnTypeDASH = 2
	QnTypeMP4  = 3
)

func (pu PlayUrlInfo) ToV1PlayInfo() *v1.PlayInfo {
	tmp := &v1.PlayInfo{
		Qn:           pu.Qn,
		Format:       pu.Format,
		QnType:       pu.QnType,
		Fnver:        pu.FnVer,
		Fnval:        pu.FnVal,
		Formats:      pu.Formats,
		VideoCodecid: pu.VideoCodecID,
		Length:       pu.Length,
		Code:         pu.Code,
		Message:      pu.Message,
		Volume:       pu.Volume,
	}
	const (
		_measured_i_limit = -9
	)
	// 针对measured_i小于-9的稿件不下发音量均衡信息
	if tmp.Volume != nil && tmp.Volume.MeasuredI <= _measured_i_limit {
		tmp.Volume = nil
	}
	if pu.QnType == QnTypeDASH {
		tmp.Info = pu.PlayDash
	} else {
		tmp.Info = pu.PlayUrl
	}
	return tmp
}

func TransformFormats(in []*purlSvcV2.FormatDescription) (ret []*v1.FormatDescription) {
	for _, f := range in {
		ret = append(ret, &v1.FormatDescription{
			Quality:     f.Quality,
			Format:      f.Format,
			Description: f.Description,
			DisplayDesc: f.DisplayDesc,
			Superscript: f.Superscript,
		})
	}
	return
}

func TransformPlayUrl(durl []*purlSvcV2.ResponseUrl) (ret *v1.PlayURL) {
	if durl == nil {
		return
	}
	ret = new(v1.PlayURL)
	for _, url := range durl {
		ret.Durl = append(ret.Durl, &v1.ResponseUrl{
			Order:     url.Order,
			Length:    url.Length,
			Size_:     url.Size_,
			Ahead:     url.Ahead,
			Vhead:     url.Vhead,
			Url:       url.Url,
			BackupUrl: url.BackupUrl,
			Md5:       url.Md5,
		})
	}
	return
}

func TransformPlayDash(dash *purlSvcV2.ResponseDash) (ret *v1.PlayDASH) {
	if dash == nil {
		return
	}
	ret = &v1.PlayDASH{
		Duration:      dash.Duration,
		MinBufferTime: dash.MinBufferTime,
	}
	for _, d := range dash.Audio {
		ret.Audio = append(ret.Audio, &v1.DashItem{
			Id:          d.Id,
			BaseUrl:     d.BaseUrl,
			BackupUrl:   d.BackupUrl,
			Bandwidth:   d.Bandwidth,
			MimeType:    d.MimeType,
			Codecs:      d.Codecs,
			SegmentBase: TransformDashSegmentBase(d.SegmentBase),
			Codecid:     d.Codecid,
			Md5:         d.Md5,
			Size_:       d.Size_,
		})
	}
	return
}

func TransformDashSegmentBase(in *purlSvcV2.DashSegmentBase) (ret *v1.DashSegmentBase) {
	if in == nil {
		return
	}
	ret = &v1.DashSegmentBase{
		Initialization: in.Initialization,
		IndexRange:     in.IndexRange,
	}
	return
}
