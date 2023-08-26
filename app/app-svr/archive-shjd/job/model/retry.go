package model

// is
const (
	TypeForUpdateVideo   = int(0)
	TypeForDelVideo      = int(1)
	TypeForUpdateArchive = int(2)
	TypeForVideoShot     = int(3)
	TypeForInternal      = int(4)
)

// RetryItem struct
type RetryItem struct {
	Tp     int      `json:"type"`
	AID    int64    `json:"aid"`
	CID    int64    `json:"cid"`
	Old    *Archive `json:"new_archive"`
	Nw     *Archive `json:"old_archive"`
	Action string   `json:"action"`
	Count  int64    `json:"cnt"`
	HdCnt  int64    `json:"hd_count"`
	HdImg  string   `json:"hd_image"`
	SdCnt  int64    `json:"sd_cnt"`
	SdImg  string   `json:"sd_image"`
}
