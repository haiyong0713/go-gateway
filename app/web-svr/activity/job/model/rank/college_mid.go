package rank

// CollegeMidInfo ...
type CollegeMidInfo struct {
	MID     int64 `json:"mid"`
	Score   int64 `json:"score"`
	History int   `json:"history"`
}

// CollegeMidScore  批量用户分数结果
type CollegeMidScore struct {
	TopLength int
	Data      []*CollegeMidInfo
}

// Len 返回长度
func (r *CollegeMidScore) Len() int {
	return len(r.Data)
}

// TopLen 返回top的长度
func (r *CollegeMidScore) TopLen() int {
	if r.Len() < r.TopLength {
		return r.Len()
	}
	return r.TopLength
}

// Less 比较
func (r *CollegeMidScore) Less(i, j int) bool {
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
func (r *CollegeMidScore) Swap(i, j int) {
	(*r).Data[i], (*r).Data[j] = (*r).Data[j], (*r).Data[i]
}

// Cut 切除
func (r *CollegeMidScore) Cut(len int) {
	(*r).Data = (*r).Data[:len]
}

// Append 添加
func (r *CollegeMidScore) Append(remainData Interface) {
	remain := remainData.(*CollegeMidScore)
	pRemain := *remain
	for i := range pRemain.Data {
		(*r).Data = append((*r).Data, pRemain.Data[i])
	}
}
