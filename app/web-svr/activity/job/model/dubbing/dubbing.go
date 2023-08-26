package dubbing

// RedisDubbing rank redis struct
type RedisDubbing struct {
	Mid   int64 `json:"mid"`
	Score int64 `json:"score"`
	Diff  int64 `json:"diff"`
	Rank  int   `json:"rank"`
}

// MapMidDubbingScore ...
type MapMidDubbingScore struct {
	Mid   int64                   `json:"mid"`
	Score map[int64]*RedisDubbing `json:"score"`
}
