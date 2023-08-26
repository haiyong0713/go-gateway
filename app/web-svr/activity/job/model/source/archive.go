package source

import (
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
)

const (
	// ArchiveStateNormal 状态正常
	ArchiveStateNormal = 1
	// ArchiveStateNotNormal 状态不正常
	ArchiveStateNotNormal = 0
)

// Archive 稿件
type Archive struct {
	Aid           int64  `json:"aid"`
	Title         string `json:"title"`
	Mid           int64  `json:"mid"`
	View          int64  `json:"view"`
	Danmaku       int64  `json:"danmaku"`
	Reply         int64  `json:"reply"`
	Fav           int64  `json:"fav"`
	Coin          int64  `json:"coin"`
	Share         int64  `json:"share"`
	Like          int64  `json:"like"`
	Videos        int64  `json:"videos"`
	Score         int64  `json:"score"`
	TypeID        int64  `json:"tid"`
	NoRank        bool   `json:"no_rank"`
	NoDynamic     bool   `json:"no_dynamic"`
	NoRecommend   bool   `json:"no_recommend"`
	NoHot         bool   `json:"no_hot"`
	NoFansDynamic bool   `json:"no_fans_dynamic"`
	NoSearch      bool   `json:"no_search"`
	NoOversea     bool   `json:"no_oversea"`
	State         int    `json:"state"`
	Ctime         int64  `json:"ctime"`
	PubTime       int64  `json:"pub_time"`
}

// IsNormal 是否正常
func (a *Archive) IsNormal() bool {
	if a.State == ArchiveStateNormal {
		return true
	}
	return false
}

// ArchiveGroup ...
type ArchiveGroup struct {
	Archive      ArchiveBatch
	MID          int64
	Score        int64
	NewScore     int64
	HistoryRank  int64
	HistoryScore int64
	Diff         int64
	OID          int64
}

// OidArchiveGroup ...
type OidArchiveGroup struct {
	Data      []*ArchiveGroup
	TopLength int
}

// ArchiveBatch 根据播放排序
type ArchiveBatch []*Archive

func (s ArchiveBatch) Len() int           { return len(s) }
func (s ArchiveBatch) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ArchiveBatch) Less(i, j int) bool { return s[i].View > s[j].View }

// ScoreFunc ...
func (a *Archive) ScoreFunc(s func(*Archive) int64) {
	a.Score = s(a)
}

// Len 返回长度
func (r *OidArchiveGroup) Len() int {
	return len(r.Data)
}

// TopLen 返回top的长度
func (r *OidArchiveGroup) TopLen() int {
	if r.Len() < r.TopLength {
		return r.Len()
	}
	return r.TopLength
}

// Less 比较
func (r *OidArchiveGroup) Less(i, j int) bool {
	if (*r).Data[i].Score == (*r).Data[j].Score {
		if (*r).Data[i].HistoryRank != 0 && (*r).Data[j].HistoryRank != 0 {
			return (*r).Data[i].HistoryRank < (*r).Data[j].HistoryRank
		}
		if (*r).Data[i].HistoryRank == 0 && (*r).Data[j].HistoryRank == 0 {
			return (*r).Data[i].OID < (*r).Data[j].OID
		}
		if (*r).Data[i].HistoryRank == 0 {
			return false
		}
		if (*r).Data[j].HistoryRank == 0 {
			return true
		}
	}
	return (*r).Data[i].Score > (*r).Data[j].Score
}

// Swap 交换
func (r *OidArchiveGroup) Swap(i, j int) {
	(*r).Data[i], (*r).Data[j] = (*r).Data[j], (*r).Data[i]
}

// Cut 切除
func (r *OidArchiveGroup) Cut(len int) {
	(*r).Data = (*r).Data[:len]
}

// Append 添加
func (r *OidArchiveGroup) Append(remainData rankmdl.Interface) {
	remain := remainData.(*OidArchiveGroup)
	pRemain := *remain
	for i := range pRemain.Data {
		(*r).Data = append((*r).Data, pRemain.Data[i])
	}
}
