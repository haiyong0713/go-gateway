package rank

import (
	go_common_library_time "go-common/library/time"
)

// DB rank db struct
type DB struct {
	ID           int64                       `json:"id"`
	SID          int64                       `json:"sid"`
	Mid          int64                       `json:"mid"`
	NickName     string                      `json:"nickname"`
	Rank         int                         `json:"rank"`
	Score        int64                       `json:"score"`
	State        int64                       `json:"state"`
	Batch        int64                       `json:"batch"`
	Remark       string                      `json:"remark"`
	RemarkOrigin interface{}                 `json:"_"`
	Ctime        go_common_library_time.Time `json:"ctime"`
	Mtime        go_common_library_time.Time `json:"mtime"`
}

// MemberRankTimes 用户上榜次数
type MemberRankTimes struct {
	Mid   int64
	Times int
}

// MemberRankHighest 用户最好成绩
type MemberRankHighest struct {
	Mid  int64
	Rank int
}

// Redis rank redis struct
type Redis struct {
	Mid   int64   `json:"mid"`
	Rank  int     `json:"rank"`
	Score int64   `json:"score"`
	Aids  []int64 `json:"aids"`
	Diff  int64   `json:"diff"`
}

// HandWriteRemark 手书活动remark结构
type HandWriteRemark struct {
	Fans int64 `json:"fans"`
}

// Interface rank interface
type Interface interface {
	// pLen is the length of rank
	Len() int
	// TopLen is the length of top rank
	TopLen() int
	// Less reports whether the element with
	// index i should sort before the element with index j.
	Less(i, j int) bool
	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
	// Cut cut the length of data
	Cut(len int)
	// Append remainData to origin data
	Append(remainData Interface)
}

func heapify(data Interface, i, n int) {
	c1 := 2*i + 1
	c2 := 2*i + 2
	swap := i
	if c1 < n && data.Less(swap, c1) {
		swap = c1
	}
	if c2 < n && data.Less(swap, c2) {
		swap = c2
	}
	if swap != i {
		data.Swap(i, swap)
		heapify(data, swap, n)
	}
}

func buildHeap(data Interface) {
	lastNode := data.TopLen() - 1
	parent := (lastNode - 1) / 2
	for i := parent; i >= 0; i-- {
		heapify(data, i, data.TopLen())
	}
}

func remain(data Interface) {
	if data.Len() > data.TopLen() {
		for i := data.TopLen(); i < data.Len(); i++ {
			if data.Less(i, 0) {
				data.Swap(i, 0)
				heapify(data, 0, data.TopLen())
			}
		}
		data.Cut(data.TopLen())
	}
}

// Add add remain
func Add(data Interface, remainData Interface) {
	data.Append(remainData)
	remain(data)
	sort(data)
}
func sort(data Interface) {
	for i := data.TopLen() - 1; i >= 0; i-- {
		data.Swap(i, 0)
		heapify(data, 0, i)
	}
}

// Sort 排序
func Sort(data Interface) {
	buildHeap(data)
	remain(data)
	sort(data)
}
