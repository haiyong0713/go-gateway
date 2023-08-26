package college

const (
	// TabNormal 普通列表
	TabNormal = 2
	// TabWhite 白名单列表
	TabWhite = 1
)

// DB ...
type DB struct {
	ID          int64  `json:"id"`
	TagID       int64  `json:"tag_id"`
	Name        string `json:"college_name"`
	ProvinceID  int64  `json:"province_id"`
	Province    string `json:"province"`
	White       string `json:"white"`
	MID         int64  `json:"mid"`
	RelationMid string `json:"relation_mid"`
	Score       int64  `json:"score"`
}

// Version ...
type Version struct {
	Version int   `json:"version"`
	Time    int64 `json:"time"`
}

// College ...
type College struct {
	ID          int64   `json:"id"`
	Name        string  `json:"college_name"`
	ProvinceID  int64   `json:"province_id"`
	Province    string  `json:"province"`
	White       []int64 `json:"white"`
	MID         int64   `json:"mid"`
	RelationMid []int64 `json:"relation_mid"`
	Score       int64   `json:"score"`
	Aids        []int64 `json:"aids"`
	TabList     []int64 `json:"tab_list"`
	TagID       int64   `json:"tag_id"`
}

// Detail ...
type Detail struct {
	ID             int64   `json:"id"`
	Score          int64   `json:"score"`
	NationwideRank int     `json:"nationwide_rank"`
	ProvinceRank   int     `json:"province_rank"`
	Province       string  `json:"province"`
	TabList        []int64 `json:"tab_list"`
	Name           string  `json:"name"`
	RelationMid    []int64 `json:"relation_mid"`
}
