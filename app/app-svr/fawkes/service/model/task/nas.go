package task

type Statistics struct {
	BatchSum     int64
	RateStr      string
	Rate         float64
	FailList     []int64
	DeleteFailed int64
	UpdateFail   int64
}

// BuildKey buildId and appKey
type BuildKey struct {
	AppKey  string `form:"app_key" json:"app_key"`
	BuildId int64  `form:"build_id" json:"build_id"`
}

type DeleteKeys struct {
	DeleteKeys []*BuildKey `json:"delete_keys"`
}

type DeleteResult struct {
	NeedDelete   []int64 `json:"need_delete"`
	DeletedId    []int64 `json:"deleted_id"`
	FailedId     []int64 `json:"failed_id_list"`
	AffectedRows int64   `json:"affected_rows"`
}

type PackDeleteState int64

const (
	FileNotExist PackDeleteState = -2
	Deleted      PackDeleteState = -1
	Active       PackDeleteState = 0
)
