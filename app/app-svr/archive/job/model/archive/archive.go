package archive

import "go-common/library/time"

const (
	StateOpen            = 0
	StateForbidRecicle   = -2
	StateForbidPolice    = -3
	StateForbidLock      = -4
	StateForbidFixed     = -6
	StateForbidUserDelay = -40
	StateForbidUpDelete  = -100
	StateForbidSteins    = -20
	CopyrightOriginal    = 1
	// attribute yes and no
	AttrYes = int32(1)
	AttrNo  = int32(0)
	// attribute bit
	AttrBitNoRank        = uint(0)
	AttrBitNoSearch      = uint(4)
	AttrBitNoRecommend   = uint(6)
	AttrBitIsBangumi     = uint(11)
	AttrBitBadgepay      = uint(18)
	AttrBitIsCooperation = uint(24)

	// archive up_from
	UpFromAnnualReport = int64(31)
)

type UpInfo struct {
	Nw  *Archive
	Old *Archive
}

// archive
type Archive struct {
	ID        int64     `json:"id"`
	Mid       int64     `json:"mid"`
	TypeID    int16     `json:"typeid"`
	HumanRank int       `json:"humanrank"`
	Duration  int       `json:"duration"`
	Title     string    `json:"title"`
	Cover     string    `json:"cover"`
	Content   string    `json:"content"`
	Tag       string    `json:"tag"`
	Attribute int32     `json:"attribute"`
	Copyright int8      `json:"copyright"`
	AreaLimit int8      `json:"arealimit"`
	State     int       `json:"state"`
	Author    string    `json:"author"`
	Access    int       `json:"access"`
	Forward   int       `json:"forward"`
	PubTime   string    `json:"pubtime"`
	Reason    string    `json:"reject_reason"`
	Round     int8      `json:"round"`
	CTime     string    `json:"ctime"`
	MTime     time.Time `json:"mtime"`
}

func (a *Archive) IsSyncState() bool {
	if a.State >= 0 || a.State == StateForbidUserDelay || a.State == StateForbidUpDelete || a.State == StateForbidRecicle || a.State == StateForbidPolice ||
		a.State == StateForbidLock || a.State == StateForbidSteins {
		return true
	}
	return false
}

type ArgStat struct {
	Aid    int64
	Field  int
	Value  int
	RealIP string
}

// AttrVal get attribute value.
func (a *Archive) AttrVal(bit uint) int32 {
	return (a.Attribute >> bit) & int32(1)
}

// Staff is
type Staff struct {
	Aid        int64  `json:"aid"`
	Mid        int64  `json:"mid"`
	Title      string `json:"title"`
	Ctime      string `json:"ctime"`
	IndexOrder int64  `json:"index_order"`
	Attribute  int64  `json:"attribute"`
}

type ArcType struct {
	ID   int64
	PID  int64
	Name string
}

type ArcVideo struct {
	Aid        int64  `json:"aid"`
	Cid        int64  `json:"cid"`
	FirstFrame string `json:"first_frame"`
}
