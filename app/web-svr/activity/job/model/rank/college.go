package rank

// College ...
type College struct {
	ID      int64   `json:"id"`
	Score   int64   `json:"score"`
	History int     `json:"history"`
	Aids    []int64 `json:"aids"`
}

// CollegeScore ...
type CollegeScore struct {
	TopLength int
	Data      []*College
}

// Len 返回长度
func (r *CollegeScore) Len() int {
	return len(r.Data)
}

// TopLen 返回top的长度
func (r *CollegeScore) TopLen() int {
	if r.Len() < r.TopLength {
		return r.Len()
	}
	return r.TopLength
}

// Less 比较
func (r *CollegeScore) Less(i, j int) bool {
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
func (r *CollegeScore) Swap(i, j int) {
	(*r).Data[i], (*r).Data[j] = (*r).Data[j], (*r).Data[i]
}

// Cut 切除
func (r *CollegeScore) Cut(len int) {
	(*r).Data = (*r).Data[:len]
}

// Append 添加
func (r *CollegeScore) Append(remainData Interface) {
	remain := remainData.(*CollegeScore)
	pRemain := *remain
	for i := range pRemain.Data {
		(*r).Data = append((*r).Data, pRemain.Data[i])
	}
}
