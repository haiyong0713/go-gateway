package rank

// ArchiveInterface 稿件排名计算接口
type ArchiveInterface interface {
	Score() *MidScoreMap
}

const handWriteTopLen = 100

// ArchiveStat is the stat of archive 稿件计数信息
type ArchiveStat struct {
	Mid int64 `json:"mid"`
	Aid int64 `json:"aid"`
	// 播放数
	View int32 `json:"view"`
	// 弹幕数
	Danmaku int32 `json:"danmaku"`
	// 评论数
	Reply int32 `json:"reply"`
	// 收藏数
	Fav int32 `json:"favorite"`
	// 投币数
	Coin int32 `json:"coin"`
	// 分享数
	Share int32 `json:"share"`
	// 当前排名
	NowRank int32 `json:"now_rank"`
	// 点赞数
	Like int32 `json:"like"`
	// 稿件一共有多少分P
	Videos int64 `json:"videos"`
	// Adjust  人工调整值
	Adjust int64 `json:"adjust"`
	// Score
	Score int64 `json:"score"`
}

// MidScore mid score
type MidScore struct {
	Mid     int64
	Score   int64
	History int
	Diff    int64
}

// MidScoreMap  批量用户分数map结果
type MidScoreMap map[int64]*MidScore

// ArchiveStatMap  批量稿件map信息
type ArchiveStatMap map[int64][]*ArchiveStat

// Score return the score of archiveBatch
func (a *ArchiveStatMap) Score(s func(*ArchiveStat) int64) *MidScoreMap {
	archiveBatch := *a
	res := MidScoreMap{}
	for k, v := range archiveBatch {
		var score int64
		for index, i := range v {
			if i != nil {
				arc := i
				score += s(arc)
				((*a)[k])[index].Score = s(arc)
			}
		}
		midScore := MidScore{
			Mid:   k,
			Score: score,
		}
		res[k] = &midScore
	}
	return &res
}

// MidScoreBatch  批量用户分数结果
type MidScoreBatch struct {
	TopLength int
	Data      []*MidScore
}

// Len 返回长度
func (r *MidScoreBatch) Len() int {
	return len(r.Data)
}

// TopLen 返回top的长度
func (r *MidScoreBatch) TopLen() int {
	if r.Len() < r.TopLength {
		return r.Len()
	}
	return r.TopLength
}

// Less 比较
func (r *MidScoreBatch) Less(i, j int) bool {
	if (*r).Data[i].Score == (*r).Data[j].Score {
		if (*r).Data[i].History != 0 && (*r).Data[j].History != 0 {
			return (*r).Data[i].History < (*r).Data[j].History
		}
		if (*r).Data[i].History == 0 {
			return false
		}
		if (*r).Data[j].History == 0 {
			return true
		}
	}
	return (*r).Data[i].Score > (*r).Data[j].Score
}

// Swap 交换
func (r *MidScoreBatch) Swap(i, j int) {
	(*r).Data[i], (*r).Data[j] = (*r).Data[j], (*r).Data[i]
}

// Cut 切除
func (r *MidScoreBatch) Cut(len int) {
	(*r).Data = (*r).Data[:len]
}

// Append 添加
func (r *MidScoreBatch) Append(remainData Interface) {
	remain := remainData.(*MidScoreBatch)
	pRemain := *remain
	for i := range pRemain.Data {
		(*r).Data = append((*r).Data, pRemain.Data[i])
	}
}

// ViewArchive 根据播放排序
type ViewArchive []*ArchiveStat

func (s ViewArchive) Len() int           { return len(s) }
func (s ViewArchive) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ViewArchive) Less(i, j int) bool { return s[i].View > s[j].View }
