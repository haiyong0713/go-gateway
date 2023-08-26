package dynamicV2

import "encoding/json"

type RcmdInfo struct {
	Code        int                     `json:"-"`
	Listm       map[int64]*RcmdInfoItem `json:"-"`
	ZoneID      int64                   `json:"-"`
	SchoolID    int64                   `json:"-"`
	UserFeature json.RawMessage         `json:"user_feature"`
}

type RcmdInfoItem struct {
	DynamicID  int64           `json:"dynamic_id"`
	UpID       int64           `json:"upid"`
	TrackID    string          `json:"trackid"`
	Goto       string          `json:"goto"`
	Rid        int64           `json:"r_id"`
	FromType   string          `json:"from_type"`
	Source     string          `json:"source"`
	AvFeature  json.RawMessage `json:"av_feature"`
	RcmdReason string          `json:"rcmd_reason"`
	Pos        int             `json:"pos"`
}

func (r *RcmdInfo) FromRcmdInfoDynID(i *RcmdReply) {
	r.Code = i.Code
	r.UserFeature = i.UserFeature
	rcmdInfocm := map[int64]*RcmdInfoItem{}
	for _, v := range i.Items {
		tmp := &RcmdInfoItem{
			DynamicID:  v.DynamicID,
			UpID:       v.UpID,
			TrackID:    v.TrackID,
			Goto:       v.Goto,
			Rid:        v.ID,
			FromType:   v.FromType,
			Source:     v.Source,
			AvFeature:  v.AvFeature,
			RcmdReason: v.RcmdReason.Content,
		}
		rcmdInfocm[v.DynamicID] = tmp
	}
	r.Listm = rcmdInfocm
}

func (r *RcmdInfo) FromRcmdInfoAvID(i *RcmdReply) {
	r.Code = i.Code
	r.UserFeature = i.UserFeature
	rcmdInfocm := map[int64]*RcmdInfoItem{}
	for _, v := range i.Items {
		tmp := &RcmdInfoItem{
			DynamicID:  v.DynamicID,
			TrackID:    v.TrackID,
			UpID:       v.UpID,
			Goto:       v.Goto,
			Rid:        v.ID,
			FromType:   v.FromType,
			Source:     v.Source,
			AvFeature:  v.AvFeature,
			RcmdReason: v.RcmdReason.Content,
		}
		rcmdInfocm[v.ID] = tmp
	}
	r.Listm = rcmdInfocm
}
