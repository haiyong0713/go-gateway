package show

import xtime "go-common/library/time"

// HotWordDatabus .
type GoodHisDatabus struct {
	Old *GoodHisRes `json:"old"`
	New *GoodHisRes `json:"new"`
}

type GoodHisRes struct {
	Aid         int64      `json:"aid"`
	Achievement string     `json:"achievement"`
	Deleted     int        `json:"deleted"`
	AddDate     xtime.Time `json:"addDate"`
}

func (v *GoodHisRes) IsDeleted() bool {
	return v.Deleted == 1
}
