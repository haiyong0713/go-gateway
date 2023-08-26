package model

import (
	"encoding/json"
	"go-common/library/log"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
)

const (
	Sentinel     = 93600
	ONLINE       = 1
	OFFLINE      = 0
	LastTime     = 2147454847
	TabTypeUpAct = 3 //UP主活动
)

type UserTab struct {
	ID        int64      `json:"id" form:"id"`
	TabType   int        `json:"tab_type" form:"tab_type"`
	Mid       int64      `json:"mid" form:"mid"`
	TabName   string     `json:"tab_name" form:"tab_name"`
	TabOrder  int64      `json:"tab_order" form:"tab_order" default:"0"`
	TabCont   int64      `json:"tab_cont" form:"tab_cont"`
	Stime     xtime.Time `json:"stime" form:"stime"`
	Etime     xtime.Time `json:"etime" form:"etime"`
	Online    int        `json:"online" form:"online" default:"-1"`
	Deleted   int        `json:"deleted" form:"deleted" default:"0"`
	IsDefault int        `json:"is_default" form:"is_default" default:"0"`
	Limits    string     `json:"limits" form:"limits"`
	H5Link    string     `json:"h5_link" form:"h5_link"`
}

type Limit struct {
	Conditions string `json:"conditions"`
	Plat       int32  `json:"plat"`
	Build      int32  `json:"build"`
}

func (tab *UserTab) IsLimitValidated(plat, build int32) bool {
	limits := make([]*Limit, 0)
	if err := json.Unmarshal([]byte(tab.Limits), &limits); err != nil {
		log.Error("UserTab Unmarshal mid(%d) err(%v)", tab.Mid, err)
		return false
	}
	if len(limits) == 0 {
		return true
	}
	for _, l := range limits {
		if plat != l.Plat {
			continue
		}
		switch l.Conditions {
		case "gt":
			{
				return build > l.Build || l.Build == 0
			}
		case "lt":
			{
				return build < l.Build
			}
		case "eq":
			{
				return build == l.Build
			}
		case "ne":
			{
				return build != l.Build
			}
		}
	}
	return false
}
func (tab *UserTab) ConvertToReply() *pb.UserTabReply {
	h5Link := tab.H5Link
	if tab.TabCont != 0 {
		h5Link = ""
	}
	return &pb.UserTabReply{
		TabType:   int32(tab.TabType),
		Mid:       tab.Mid,
		TabName:   tab.TabName,
		TabOrder:  int32(tab.TabOrder),
		TabCont:   tab.TabCont,
		IsDefault: int32(tab.IsDefault),
		H5Link:    h5Link,
	}
}
