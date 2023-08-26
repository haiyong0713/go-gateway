package model

// PubResp Response for publish interface
type PubResp struct {
	CurrVer  int64   `json:"curr_ver"`
	DiffProd []int64 `json:"diff_prod"`
	DiffTest []int64 `json:"diff_test"`
}

// Ver reprensents the already generated versions
type Ver struct {
	FromVer int64
	ID      int64
}

// PushMsg .
type PushMsg struct {
	Key       string
	Time      int64
	Operation int
	Qps       int
}

// TableName gives the table name of the model
func (*Ver) TableName() string {
	return "resource_file"
}
