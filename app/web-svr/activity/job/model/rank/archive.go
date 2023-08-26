package rank

// ArchiveScore archive score
type ArchiveScore struct {
	Aid     int64
	Mid     int64
	Score   int64
	History int
}

// ArchiveScoreMap  批量分数map结果
type ArchiveScoreMap map[int64]*ArchiveScore

// ArchiveBatch  批量稿件map信息
type ArchiveBatch map[int64]*ArchiveStat

// Score return the score of archiveBatch
func (a *ArchiveBatch) Score(s func(*ArchiveStat) int64) *ArchiveScoreMap {
	archiveBatch := *a
	res := ArchiveScoreMap{}
	for k, v := range archiveBatch {
		archiveScore := ArchiveScore{
			Aid:   k,
			Mid:   v.Mid,
			Score: s(v),
		}
		(*a)[k].Score = s(v)
		res[k] = &archiveScore
	}
	return &res
}

// ArchiveScoreBatch  批量稿件分数结果
type ArchiveScoreBatch struct {
	TopLength int
	Data      []*ArchiveScore
}

// Len 返回长度
func (r *ArchiveScoreBatch) Len() int {
	return len(r.Data)
}

// TopLen 返回top的长度
func (r *ArchiveScoreBatch) TopLen() int {
	if r.Len() < r.TopLength {
		return r.Len()
	}
	return r.TopLength
}

// Less 比较
func (r *ArchiveScoreBatch) Less(i, j int) bool {
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
func (r *ArchiveScoreBatch) Swap(i, j int) {
	(*r).Data[i], (*r).Data[j] = (*r).Data[j], (*r).Data[i]
}

// Cut 切除
func (r *ArchiveScoreBatch) Cut(len int) {
	(*r).Data = (*r).Data[:len]
}

// Append 添加
func (r *ArchiveScoreBatch) Append(remainData Interface) {
	remain := remainData.(*ArchiveScoreBatch)
	pRemain := *remain
	for i := range pRemain.Data {
		(*r).Data = append((*r).Data, pRemain.Data[i])
	}
}
