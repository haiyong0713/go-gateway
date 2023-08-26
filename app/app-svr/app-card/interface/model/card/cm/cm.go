package cm

import (
	"encoding/json"
	"strconv"

	"go-common/library/log"
	xtime "go-common/library/time"

	"github.com/pkg/errors"
)

// Ad is
type Ad struct {
	RequestID string                       `json:"request_id,omitempty"`
	AdsInfo   map[int64]map[int32]*AdsInfo `json:"ads_info,omitempty"`
	ClientIP  string                       `json:"-"`
}

// AdsInfo is
type AdsInfo struct {
	Index     int32   `json:"index,omitempty"`
	IsAd      bool    `json:"is_ad,omitempty"`
	CmMark    int64   `json:"cm_mark,omitempty"`
	AdInfo    *AdInfo `json:"ad_info,omitempty"`
	CardIndex int32   `json:"card_index,omitempty"`
}

// AdInfo is
type AdInfo struct {
	CreativeID      int64            `json:"creative_id,omitempty"`
	CreativeType    int32            `json:"creative_type,omitempty"`
	CardType        int32            `json:"card_type,omitempty"`
	CreativeContent *CreativeContent `json:"creative_content,omitempty"`
	AdCb            string           `json:"ad_cb,omitempty"`
	Resource        int64            `json:"resource,omitempty"`
	Source          int32            `json:"source,omitempty"`
	RequestID       string           `json:"request_id,omitempty"`
	IsAd            bool             `json:"is_ad,omitempty"`
	CmMark          int64            `json:"cm_mark,omitempty"`
	Index           int32            `json:"index,omitempty"`
	IsAdLoc         bool             `json:"is_ad_loc,omitempty"`
	CardIndex       int32            `json:"card_index,omitempty"`
	ClientIP        string           `json:"client_ip,omitempty"`
	Extra           json.RawMessage  `json:"extra,omitempty"`
	CreativeStyle   int32            `json:"creative_style,omitempty"`
	DiscardReason   string           `json:"discard_reason,omitempty"`
	RoomID          int64            `json:"room_id,omitempty"`
	LiveBookingID   int64            `json:"live_booking_id,omitempty"`
	TopViewID       int64            `json:"top_view_id,omitempty"`
	EpId            int64            `json:"epid,omitempty"`
}

// NewAd is
type NewAd struct {
	Resources map[int64][]*AdResource `json:"oversaturated_resources,omitempty"`
}

// AdResource is
type AdResource struct {
	RequestID  string       `json:"request_id,omitempty"`
	Resource   int64        `json:"resource_id,omitempty"`
	Source     int32        `json:"source_id,omitempty"`
	IsAdLoc    bool         `json:"is_ad_loc,omitempty"`
	ServerType int32        `json:"server_type,omitempty"`
	ClientIP   string       `json:"client_ip,omitempty"`
	CardIndex  int32        `json:"card_index,omitempty"`
	Index      int32        `json:"index,omitempty"`
	AdContents []*AdContent `json:"ad_contents,omitempty"`
}

// AdContent is
type AdContent struct {
	CreativeID           int64            `json:"creative_id,omitempty"`
	CreativeType         int32            `json:"creative_type,omitempty"`
	CardType             int32            `json:"card_type,omitempty"`
	CreativeContent      *CreativeContent `json:"creative_content,omitempty"`
	AdCb                 string           `json:"ad_cb,omitempty"`
	CreativeStyle        int32            `json:"creative_style,omitempty"`
	Extra                json.RawMessage  `json:"extra,omitempty"`
	CmMark               int64            `json:"cm_mark,omitempty"`
	IsAd                 bool             `json:"is_ad,omitempty"`
	DiscardReason        string           `json:"discard_reason,omitempty"`
	PromotionPurposeType int64            `json:"promotion_purpose_type,omitempty"`
	PromotionTargetID    string           `json:"promotion_target_id,omitempty"`
	LiveBookingID        int64            `json:"live_booking_id,omitempty"`
	StoryCartIcon        *StoryCartIcon   `json:"story_cart_icon,omitempty"`
}

type StoryCartIcon struct {
	IconURL   string `json:"icon_url,omitempty"`
	IconText  string `json:"icon_text,omitempty"`
	IconTitle string `json:"icon_title,omitempty"`
}

// CreativeContent is
type CreativeContent struct {
	Title    string `json:"title,omitempty"`
	Desc     string `json:"description,omitempty"`
	VideoID  int64  `json:"video_id,omitempty"`
	UserName string `json:"username,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	ImageMD5 string `json:"image_md5,omitempty"`
	LogURL   string `json:"log_url,omitempty"`
	LogMD5   string `json:"log_md5,omitempty"`
	URL      string `json:"url,omitempty"`
	ClickURL string `json:"click_url,omitempty"`
	ShowURL  string `json:"show_url,omitempty"`
}

const (
	_cardAdInlineLive = 44
	_cardAdLive       = 63
	_cardAdInlineAv   = 74
	_adLiveType       = 8
	_cardAdPgc        = 97
	_cardAdInlinePgc  = 98
	_adPgcType        = 20
)

// AdChange is
func (advert *Ad) AdChange(resourceID int64, cardType int32) (adm map[int32][]*AdInfo, adAidm, adRoomidm map[int64]struct{}) {
	if advert == nil || len(advert.AdsInfo) == 0 {
		return
	}
	adsInfo, ok := advert.AdsInfo[resourceID]
	if !ok {
		return
	}
	adm = map[int32][]*AdInfo{}
	adAidm = make(map[int64]struct{}, len(adsInfo))
	adRoomidm = make(map[int64]struct{}, len(adsInfo))
	for source, info := range adsInfo {
		if info == nil {
			continue
		}
		var adInfo *AdInfo
		if info.AdInfo != nil {
			adInfo = info.AdInfo
			adInfo.RequestID = advert.RequestID
			adInfo.Resource = resourceID
			adInfo.Source = source
			adInfo.IsAd = info.IsAd
			adInfo.IsAdLoc = true
			adInfo.CmMark = info.CmMark
			adInfo.Index = info.Index
			adInfo.CardIndex = info.CardIndex
			adInfo.ClientIP = advert.ClientIP
			if adInfo.CreativeID != 0 && adInfo.CardType == cardType {
				adAidm[adInfo.CreativeContent.VideoID] = struct{}{}
			}
		} else {
			adInfo = &AdInfo{
				RequestID: advert.RequestID,
				Resource:  resourceID,
				Source:    source,
				IsAdLoc:   true,
				IsAd:      info.IsAd,
				CmMark:    info.CmMark,
				Index:     info.Index,
				CardIndex: info.CardIndex,
				ClientIP:  advert.ClientIP,
			}
		}
		//兼容老广告逻辑
		// adm[adInfo.CardIndex-1] = adInfo
		adm[adInfo.CardIndex-1] = []*AdInfo{adInfo}
	}

	return
}

// NewAdChange is
func (advert *NewAd) NewAdChange(resourceID int64, cardType int32) (adm map[int32][]*AdInfo, adAidm map[int64]struct{}, adRoomidm map[int64]struct{}) {
	if advert == nil || len(advert.Resources) == 0 {
		return
	}
	adsInfo, ok := advert.Resources[resourceID]
	if !ok {
		return
	}
	adm = map[int32][]*AdInfo{}
	adAidm = map[int64]struct{}{}
	adRoomidm = map[int64]struct{}{}
	for _, infos := range adsInfo {
		adInfo := &AdInfo{
			Resource:  infos.Resource,
			Source:    infos.Source,
			RequestID: infos.RequestID,
			Index:     infos.Index,
			IsAdLoc:   true,
			CardIndex: infos.CardIndex,
			ClientIP:  infos.ClientIP,
		}
		if len(infos.AdContents) == 0 {
			adm[adInfo.CardIndex-1] = []*AdInfo{adInfo}
			continue
		}
		for _, info := range infos.AdContents {
			tmp := &AdInfo{}
			*tmp = *adInfo
			tmp.CreativeID = info.CreativeID
			tmp.CreativeType = info.CreativeType
			tmp.CardType = info.CardType
			tmp.CreativeContent = info.CreativeContent
			tmp.AdCb = info.AdCb
			tmp.IsAd = info.IsAd
			tmp.CmMark = info.CmMark
			tmp.Extra = info.Extra
			tmp.CreativeStyle = info.CreativeStyle
			tmp.LiveBookingID = info.LiveBookingID
			if tmp.CreativeID != 0 && tmp.CardType == cardType {
				adAidm[tmp.CreativeContent.VideoID] = struct{}{}
			}
			if (tmp.CardType == _cardAdLive || tmp.CardType == _cardAdInlineLive) && info.PromotionPurposeType == _adLiveType {
				rid, err := strconv.ParseInt(info.PromotionTargetID, 10, 64)
				if err == nil {
					adRoomidm[rid] = struct{}{}
					tmp.RoomID = rid
				} else {
					log.Error("Failed to parse ad live id: %+v", errors.WithStack(err))
				}
			}
			adm[tmp.CardIndex-1] = append(adm[tmp.CardIndex-1], tmp)
		}
	}
	return
}

func AdChangeV2(infos *AdResource, cardType int32) (adm []*AdInfo, adAid, adRoomid, adEpid int64) {
	if infos == nil {
		return
	}
	adInfo := &AdInfo{
		Resource:  infos.Resource,
		Source:    infos.Source,
		RequestID: infos.RequestID,
		Index:     infos.Index,
		IsAdLoc:   true,
		CardIndex: infos.CardIndex,
		ClientIP:  infos.ClientIP,
	}
	if len(infos.AdContents) == 0 {
		adm = []*AdInfo{adInfo}
	} else {
		for _, info := range infos.AdContents {
			if info == nil {
				continue
			}
			tmp := &AdInfo{}
			*tmp = *adInfo
			tmp.CreativeID = info.CreativeID
			tmp.CreativeType = info.CreativeType
			tmp.CardType = info.CardType
			tmp.CreativeContent = info.CreativeContent
			tmp.AdCb = info.AdCb
			tmp.IsAd = info.IsAd
			tmp.CmMark = info.CmMark
			tmp.Extra = info.Extra
			tmp.CreativeStyle = info.CreativeStyle
			tmp.LiveBookingID = info.LiveBookingID
			if tmp.CreativeID != 0 && tmp.CardType == cardType {
				adAid = tmp.CreativeContent.VideoID
			}
			if tmp.CreativeID != 0 && tmp.CardType == _cardAdInlineAv {
				adAid = tmp.CreativeContent.VideoID
			}
			if (tmp.CardType == _cardAdLive || tmp.CardType == _cardAdInlineLive) && info.PromotionPurposeType == _adLiveType {
				rid, err := strconv.ParseInt(info.PromotionTargetID, 10, 64)
				if err == nil {
					adRoomid = rid
					tmp.RoomID = rid
				} else {
					log.Error("Failed to parse ad live id: %+v", errors.WithStack(err))
				}
			}
			if (tmp.CardType == _cardAdPgc || tmp.CardType == _cardAdInlinePgc) && info.PromotionPurposeType == _adPgcType {
				epid, err := strconv.ParseInt(info.PromotionTargetID, 10, 64)
				if err == nil {
					adEpid = epid
					tmp.EpId = epid
				} else {
					log.Error("Failed to parse ad pgc id: %+v", errors.WithStack(err))
				}
			}
			tmp.DiscardReason = info.DiscardReason
			adm = append(adm, tmp)
		}
	}
	return
}

func ConstructAdInfos(infos *AdResource) []*AdInfo {
	if infos == nil {
		return nil
	}
	adInfo := &AdInfo{
		Resource:  infos.Resource,
		Source:    infos.Source,
		RequestID: infos.RequestID,
		Index:     infos.Index,
		IsAdLoc:   true,
		CardIndex: infos.CardIndex,
		ClientIP:  infos.ClientIP,
	}
	if len(infos.AdContents) == 0 {
		return []*AdInfo{adInfo}
	}
	adInfos := make([]*AdInfo, 0, len(infos.AdContents))
	for _, info := range infos.AdContents {
		if info == nil {
			continue
		}
		sub := &AdInfo{}
		*sub = *adInfo
		sub.CreativeID = info.CreativeID
		sub.CreativeType = info.CreativeType
		sub.CardType = info.CardType
		sub.CreativeContent = info.CreativeContent
		sub.AdCb = info.AdCb
		sub.IsAd = info.IsAd
		sub.CmMark = info.CmMark
		sub.Extra = info.Extra
		sub.CreativeStyle = info.CreativeStyle
		sub.DiscardReason = info.DiscardReason
		adInfos = append(adInfos, sub)
	}
	return adInfos
}

func ConstructAdIndexMapFrom(resourceID int64, advert *NewAd) map[int32][]*AdInfo {
	if advert == nil || len(advert.Resources) == 0 {
		return nil
	}
	adsInfo, ok := advert.Resources[resourceID]
	if !ok {
		return nil
	}
	adIndexMap := make(map[int32][]*AdInfo, len(adsInfo))
	for _, infos := range adsInfo {
		adInfo := &AdInfo{
			Resource:  infos.Resource,
			Source:    infos.Source,
			RequestID: infos.RequestID,
			Index:     infos.Index,
			IsAdLoc:   true,
			CardIndex: infos.CardIndex,
			ClientIP:  infos.ClientIP,
		}
		if len(infos.AdContents) == 0 {
			adIndexMap[adInfo.CardIndex-1] = []*AdInfo{adInfo}
			continue
		}
		for _, info := range infos.AdContents {
			sub := &AdInfo{}
			*sub = *adInfo
			sub.CreativeID = info.CreativeID
			sub.CreativeType = info.CreativeType
			sub.CardType = info.CardType
			sub.CreativeContent = info.CreativeContent
			sub.AdCb = info.AdCb
			sub.IsAd = info.IsAd
			sub.CmMark = info.CmMark
			sub.Extra = info.Extra
			sub.CreativeStyle = info.CreativeStyle
			adIndexMap[sub.CardIndex-1] = append(adIndexMap[sub.CardIndex-1], sub)
		}
	}
	return adIndexMap
}

// StoryAdResource is
type StoryAdResource struct {
	RequestID       string     `json:"request_id,omitempty"`
	ResourceID      int64      `json:"resource_id,omitempty"`
	SourceID        int32      `json:"source_id,omitempty"`
	IsAdLoc         bool       `json:"is_ad_loc,omitempty"`
	ServerType      int32      `json:"server_type,omitempty"`
	ClientIP        string     `json:"client_ip,omitempty"`
	CardIndex       int32      `json:"card_index,omitempty"`
	StoryVideoID    int64      `json:"story_video_id,omitempty"`
	AdvertiseType   int32      `json:"advertise_type,omitempty"`
	StoryLiveRoomId int64      `json:"story_live_room_id,omitempty"`
	StoryUpMid      int64      `json:"story_up_mid,omitempty"`
	AdContent       *AdContent `json:"ad_content,omitempty"`
}

func AsStoryAdInfo(arg *StoryAdResource) (*AdInfo, bool) {
	if arg == nil {
		return nil, false
	}
	adInfo := &AdInfo{
		Resource:  arg.ResourceID,
		Source:    arg.SourceID,
		RequestID: arg.RequestID,
		IsAdLoc:   true,
		CardIndex: arg.CardIndex,
		ClientIP:  arg.ClientIP,
	}
	if arg.AdContent == nil {
		return adInfo, false
	}
	adInfo.CreativeID = arg.AdContent.CreativeID
	adInfo.AdCb = arg.AdContent.AdCb
	adInfo.IsAd = arg.AdContent.IsAd
	adInfo.Extra = arg.AdContent.Extra
	return adInfo, true
}

func AsStoryCartIcon(arg *StoryAdResource) (*StoryCartIcon, bool) {
	if arg == nil || arg.AdContent == nil || arg.AdContent.StoryCartIcon == nil {
		return nil, false
	}
	return arg.AdContent.StoryCartIcon, true
}

type CmInfo struct {
	HidePlayButton    bool       `json:"hide_play_button,omitempty"`
	ReservationTime   xtime.Time `json:"reservation_time,omitempty"`
	ReservationNum    int64      `json:"reservation_num,omitempty"`
	ReservationStatus int64      `json:"reservation_status,omitempty"`
}
