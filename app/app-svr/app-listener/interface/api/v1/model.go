package v1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"go-common/library/log"
)

const (
	EtOpRecommend = "reco"
	EtOpFavorite  = "fav"
	// ugc合集
	EtOpUGCSeason = "ugcss"
	EtOpHistory   = "hist"
	EtOpFinding   = "find"
	// 手动带入（目前是播放页三点）
	EtOpManual = "manual"
	// 老音频歌单
	EtOpAudioMenu = "menu"
	// 单音频卡（视频关联卡/动态音频卡/主站收藏内点击音频）
	EtOpAudioSingle = "ausingle"
	// 用户空间页音频投稿列表
	EtOpSpaceAudio = "spaceau"
	// 播单
	EtOpMediaList = "medialist"
	// BGM 音乐发现 版权pv续播场景
	EtOpMusicPv = "musicpv"
	// 未知
	EtOpUnknown = "unknown"
)

type EventTrackingOpt func(et *EventTracking)

var (
	OpHistory = func(et *EventTracking) {
		et.Operator = EtOpHistory
	}
	OpManual = func(et *EventTracking) {
		et.Operator = EtOpManual
	}
	OpRecommend = func(et *EventTracking) {
		et.Operator = EtOpRecommend
	}
	OpFavorite = func(et *EventTracking) {
		et.Operator = EtOpFavorite
	}
	OpFinding = func(et *EventTracking) {
		et.Operator = EtOpFinding
	}
	OpUGCSeason = func(et *EventTracking) {
		et.Operator = EtOpUGCSeason
	}
	OpAudioMenu = func(et *EventTracking) {
		et.Operator = EtOpAudioMenu
	}
	OpAudioSingle = func(et *EventTracking) {
		et.Operator = EtOpAudioSingle
	}
	OpSpaceAudio = func(et *EventTracking) {
		et.Operator = EtOpSpaceAudio
	}
	OpMediaList = func(et *EventTracking) {
		et.Operator = EtOpMediaList
	}
	OpMusicPv = func(et *EventTracking) {
		et.Operator = EtOpMusicPv
	}
	OpUnknown = func(et *EventTracking) {
		et.Operator = EtOpUnknown
	}
)

func OpByPlaylistSource(src PlaylistSource, subType ...string) EventTrackingOpt {
	switch src {
	case PlaylistSource_USER_FAVOURITE:
		if len(subType) > 0 {
			switch subType[0] {
			case "21":
				return OpUGCSeason
			default:
				return OpFavorite
			}
		}
		return OpFavorite
	case PlaylistSource_PICK_CARD:
		return OpFinding
	case PlaylistSource_AUDIO_COLLECTION:
		return OpAudioMenu
	case PlaylistSource_AUDIO_CARD:
		return OpAudioSingle
	case PlaylistSource_MEM_SPACE:
		return OpSpaceAudio
	case PlaylistSource_MEDIA_LIST:
		return OpMediaList
	case PlaylistSource_UP_ARCHIVE:
		return OpMusicPv
	default:
		return OpUnknown
	}
}

func (et *EventTracking) ComposeJson() {
	if et == nil {
		return
	}
	et.TrackJson = ""
	data, err := json.Marshal(et)
	if err == nil {
		et.TrackJson = string(data)
	} else {
		log.Error("error setting track json for et (%+v): %v", et, err)
	}
}

func (pi *PlayItem) SetEventTracking(opts ...EventTrackingOpt) *PlayItem {
	if pi == nil {
		return pi
	}
	if pi.Et == nil {
		pi.Et = &EventTracking{}
	}
	for _, f := range opts {
		f(pi.Et)
	}
	pi.Et.ComposeJson()
	return pi
}

// 根据服务端下发的et字段修正item type
func (pi *PlayItem) FixItemTypeByEt() {
	if pi == nil {
		return
	}
	if pi.Et != nil {
		switch pi.Et.EntityType {
		case "av":
			pi.ItemType = 1
		case "au":
			pi.ItemType = 3
		case "ep":
			pi.ItemType = 2
		}
	}
}

func (pi *PlayItem) Hash() string {
	if pi == nil {
		return ""
	}
	return fmt.Sprintf("%d-%d", pi.ItemType, pi.Oid)
}

func (pi *PlayItem) Equal(p *PlayItem) bool {
	if pi == nil || p == nil {
		return false
	}
	if pi == p {
		return true
	}
	if pi.ItemType == p.ItemType && pi.Oid == p.Oid {
		return true
	}
	return false
}

func (fi *FavItem) SetEventTracking(opts ...EventTrackingOpt) {
	if fi == nil {
		return
	}
	if fi.Et == nil {
		fi.Et = &EventTracking{}
	}
	for _, f := range opts {
		f(fi.Et)
	}
	fi.Et.ComposeJson()
}

func (di *DetailItem) IsPlayable() bool {
	return di.Playable == 0
}

type ApplyOrderOpt struct {
	Anchor *PlayItem
}

func (so *SortOption) ApplyOrderToV1PlayItems(items []*PlayItem, opt *ApplyOrderOpt) (ret []*PlayItem) {
	if so == nil {
		return items
	}
	switch so.Order {
	case ListOrder_NO_ORDER, ListOrder_ORDER_NORMAL:
		ret = items
	case ListOrder_ORDER_REVERSE:
		ret = make([]*PlayItem, 0, len(items))
		for i := len(items) - 1; i >= 0; i-- {
			ret = append(ret, items[i])
		}
	case ListOrder_ORDER_RANDOM:
		if len(items) >= 1 {
			r := rand.New(rand.NewSource(time.Now().Unix()))
			ret = make([]*PlayItem, len(items))
			for i := len(items); i >= 1; i-- {
				tgt := r.Intn(i)
				items[tgt], items[i-1] = items[i-1], items[tgt]
				ret[i-1] = items[i-1]
			}
		} else {
			ret = items
		}
		// 如果有锚点，则把锚点放到第一位
		if opt != nil && opt.Anchor != nil {
			for i, item := range ret {
				if opt.Anchor.Equal(item) {
					ret[i], ret[0] = ret[0], ret[i]
					break
				}
			}
		}
	}

	return ret
}

const (
	AttrYes = 0x1

	MenuAttrDefaultBit = 0x0
)

func (mm *MusicMenu) IsDefaultMenu() bool {
	return mm.Attr>>MenuAttrDefaultBit&1 == AttrYes
}

func (phr *PlayHistoryResp) ApplyHistoryTag(local0h int64) {
	if phr == nil || len(phr.List) <= 0 {
		return
	}
	applyHistoryTag(phr.List, local0h)
}

func applyHistoryTag(list []*DetailItem, local0h int64) {
	today0h := local0h
	if today0h <= 0 {
		today0h = today0hUnix()
	}
	for _, m := range list {
		if m.LastPlayTime <= 0 {
			continue
		}
		diff := m.LastPlayTime - today0h
		switch {
		case diff >= 0: // 播放时间大于今天0点，算今天
			m.HistoryTag = "今天"
		case -diff < 3600*24: // 间隔小于24小时，算昨天
			m.HistoryTag = "昨天"
		default:
			m.HistoryTag = "更早"
		}
	}
}

// 计算当前时区今天0点的unix timestamp
func today0hUnix() int64 {
	_, offset := time.Now().Zone()
	now := time.Now()
	//nolint:gomnd
	timeElapsed := (now.Unix() + int64(offset)) % (3600 * 24)
	return now.Add(-(time.Second * time.Duration(timeElapsed))).Unix()
}

func (p *PlayURLResp) AddExpireTime() {
	if p == nil || p.PlayerInfo == nil {
		return
	}
	for _, info := range p.PlayerInfo {
		var query url.Values
		switch inf := info.Info.(type) {
		case *PlayInfo_PlayDash:
			if inf == nil || inf.PlayDash == nil {
				continue
			}
			if len(inf.PlayDash.Audio) > 0 {
				dashItem := inf.PlayDash.Audio[0]
				if uri, err := url.Parse(dashItem.BaseUrl); err != nil {
					log.Warn("error parse dash playurl. using default 2h expire time: %v", err)
					info.ExpireTime = uint64(time.Now().Add(2 * time.Hour).Unix())
					continue
				} else {
					query = uri.Query()
				}
			}
		case *PlayInfo_PlayUrl:
			if inf == nil || inf.PlayUrl == nil {
				continue
			}
			if len(inf.PlayUrl.Durl) > 0 {
				durlItem := inf.PlayUrl.Durl[0]
				if uri, err := url.Parse(durlItem.Url); err != nil {
					log.Warn("error parse durl playurl. using default 2h expire time: %v", err)
					info.ExpireTime = uint64(time.Now().Add(2 * time.Hour).Unix())
					continue
				} else {
					query = uri.Query()
				}
			}
		}
		if query != nil {
			val := query.Get("deadline")
			if ddl, err := strconv.ParseUint(val, 10, 64); err != nil {
				log.Warn("error parse playurl query deadline. using default 2h expire time: %v", err)
				info.ExpireTime = uint64(time.Now().Add(2 * time.Hour).Unix())
			} else {
				info.ExpireTime = ddl
			}
		} else {
			log.Warn("no deadline detected for playurl(%+v). using default 2h expire time", p.Item)
			info.ExpireTime = uint64(time.Now().Add(2 * time.Hour).Unix())
		}
	}
}

func (ro *RcmdOffset) UnmarshalFromBase64(data string) error {
	if ro == nil {
		return fmt.Errorf("unexpected nil rcmdOffset")
	}
	if len(data) <= 0 {
		ro.Page = 1
		return nil
	}
	pb, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return ro.Unmarshal(pb)
}

func (ro *RcmdOffset) MarshalToBase64(ctx context.Context) string {
	if ro == nil {
		return ""
	}
	res, err := ro.Marshal()
	if err != nil {
		log.Errorc(ctx, "error encoding RcmdOffset: %v", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(res)
}

var _serverRcmdMapping = map[RcmdPlaylistReq_RcmdFrom]int64{
	RcmdPlaylistReq_UP_ARCHIVE:   1,
	RcmdPlaylistReq_INDEX_ENTRY:  3,
	RcmdPlaylistReq_ARCHIVE_VIEW: 4,
}

func (r *RcmdPlaylistReq) GetServerRcmdFromType() int64 {
	/*
		enum RecommendFromType {
			Default     = 0;
			VideoPlay   = 1; // 视频播放页三点
			Mine        = 2; // 我的页入口
			HomeTopLeft = 3; // 首页左上角入口
			HalfScreenPlayer = 4; // 半屏播放器入口
		}
	*/
	return _serverRcmdMapping[r.GetFrom()]
}
