package business

// Version struct.
type Version struct {
	Config int64 `json:"config"`
	FF     int64 `json:"ff"`
}

type ActiveLaser2Result struct {
	TaskID int64 `json:"task_id"`
}

type PcdnFile struct {
	Key      string `json:" key"`
	Url      string `json:" url"`
	MD5      string `json:" md5"`
	Size     int64  `json:" size"`
	Business string `json:" business"`
}
