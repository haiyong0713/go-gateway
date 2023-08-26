package like

import (
	xtime "go-common/library/time"
)

// PubSub .
type PubSub struct {
	ID         int64      `json:"id"`
	Oid        int64      `json:"oid"`
	Type       int64      `json:"type"`
	State      int64      `json:"state"`
	Stime      xtime.Time `json:"stime"`
	Etime      xtime.Time `json:"etime"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
	Name       string     `json:"name"`
	ActURL     string     `json:"act_url"`
	Lstime     xtime.Time `json:"lstime"`
	Letime     xtime.Time `json:"letime"`
	Cover      string     `json:"cover"`
	Dic        string     `json:"dic"`
	H5Cover    string     `json:"h5_cover"`
	AndroidURL string     `json:"android_url"`
	IosURL     string     `json:"ios_url"`
	ChildSids  string     `json:"child_sids"`
	Calendar   string     `json:"calendar"`
}

// DisplayLike .
func (sub *SubjectItem) DisplayLike() bool {
	var dis = false
	switch sub.Type {
	case PICTURELIKE, DRAWYOOLIKE, TEXTLIKE, VIDEOLIKE, VIDEO2, VIDEO, SMALLVIDEO, MUSIC, PHONEVIDEO, STORYKING:
		if sub.AttrFlag(FLAGRANKCLOSE) {
			dis = true
		}
	}
	return dis
}

func (sub *SubjectItem) IsVideoCollection() bool {
	return sub.Type == VIDEO2 || sub.Type == PHONEVIDEO
}

func (sub *SubjectItem) IsVideoLike() bool {
	return sub.Type == VIDEOLIKE
}

func (sub *SubjectItem) IsARTICLE() bool {
	return sub.Type == ARTICLE
}

// AttrFlag .
func (sub *SubjectItem) AttrFlag(bit int64) bool {
	if sub.Flag == 0 {
		return true
	}
	if (sub.Flag & bit) == bit {
		return false
	}
	return true
}

// IsDailyLike .
func (sub *SubjectItem) IsDailyLike() bool {
	return !sub.AttrFlag(FLAGDAILYLIKETYPE)
}

func (sub *SubjectItem) IsAttrNew() bool {
	return ((sub.Flag >> FLAGATTRISNEW) & int64(1)) == 1
}

func (sub *SubjectItem) IsForbidHotList() bool {
	return ((sub.Flag >> FLATFORBIDHOTLIST) & int64(1)) == 1
}

func (sub *SubjectItem) IsForbidCancel() bool {
	return ((sub.Flag >> FLATFORBIDCANCEL) & int64(1)) == 1
}

func (sub *SubjectItem) IsQuestionnaire() bool {
	return ((sub.Flag >> FLAGQUESTIONNAIRE) & int64(1)) == 1
}

func (sub *SubjectItem) IsDuplicateSubmit() bool {
	return ((sub.Flag >> FLAGDUPLICATESUBMIT) & int64(1)) == 1
}

// IsShieldRank 是否禁止排行
func (sub *SubjectItem) IsShieldRank() bool {
	return ((sub.ShieldFlag >> ShieldNoRank) & int64(1)) == 1
}

// IsShieldDynamic 是否禁止动态
func (sub *SubjectItem) IsShieldDynamic() bool {
	return ((sub.ShieldFlag >> ShieldNoDynamic) & int64(1)) == 1
}

// IsShieldRecommend 是否禁止推荐
func (sub *SubjectItem) IsShieldRecommend() bool {
	return ((sub.ShieldFlag >> ShieldNoRecommend) & int64(1)) == 1
}

// IsShieldHot 是否禁止热门
func (sub *SubjectItem) IsShieldHot() bool {
	return ((sub.ShieldFlag >> ShieldNoHot) & int64(1)) == 1
}

// IsShieldFansDynamic 是否禁止粉丝动态
func (sub *SubjectItem) IsShieldFansDynamic() bool {
	return ((sub.ShieldFlag >> ShieldNoFansDynamic) & int64(1)) == 1
}

// IsShieldSearch 是否禁止搜索
func (sub *SubjectItem) IsShieldSearch() bool {
	return ((sub.ShieldFlag >> ShieldNoSearch) & int64(1)) == 1
}

// IsShieldOversea 是否禁止海外
func (sub *SubjectItem) IsShieldOversea() bool {
	return ((sub.ShieldFlag >> ShieldNoOversea) & int64(1)) == 1
}

// TextType .
func (sub *SubjectItem) TextType() bool {
	res := false
	switch sub.Type {
	case TEXT, TEXTLIKE, QUESTION, RESERVATION, CLOCKIN, USERACTIONSTAT:
		res = true
	}
	return res
}

func (sub *SubjectItem) CacheReserve() bool {
	if (sub.Type == RESERVATION && sub.IsAttrNew()) || sub.Type == CLOCKIN || sub.Type == USERACTIONSTAT || sub.Type == UPRESERVATIONARC || sub.Type == UPRESERVATIONLIVE {
		return true
	}
	return false
}

func (sub *SubjectItem) CacheOnly() bool {
	if sub.Type == QUESTION || sub.Type == RESERVATION || sub.Type == CLOCKIN || sub.Type == USERACTIONSTAT || sub.Type == UPRESERVATIONARC || sub.Type == UPRESERVATIONLIVE {
		return true
	}
	return false
}

// PublicData .
func (sub *SubjectItem) PublicData() *PubSub {
	return &PubSub{
		ID:         sub.ID,
		Oid:        sub.Oid,
		Type:       sub.Type,
		State:      sub.State,
		Stime:      sub.Stime,
		Etime:      sub.Etime,
		Ctime:      sub.Ctime,
		Mtime:      sub.Mtime,
		Name:       sub.Name,
		ActURL:     sub.ActURL,
		Lstime:     sub.Lstime,
		Letime:     sub.Letime,
		Cover:      sub.Cover,
		Dic:        sub.Dic,
		H5Cover:    sub.H5Cover,
		AndroidURL: sub.AndroidURL,
		IosURL:     sub.IosURL,
		ChildSids:  sub.ChildSids,
		Calendar:   sub.Calendar,
	}
}
