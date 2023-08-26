package vogue

const (
	TimeFormat = "2006-01-02 15:04:05"
)

// Page
type Page struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}

// ExportCsvParam
type ExportCsvParam struct {
	FileNameFormat string
	FileNameParams []interface{}
	Header         []string
	Result         [][]string
}
