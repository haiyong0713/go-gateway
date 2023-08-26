package dynamic

// ParamActList .
type ParamActList struct {
	PageID    int64  `form:"page_id"`
	Ukey      string `form:"ukey"`
	Ps        int    `form:"ps" validate:"min=1"`
	Sid       int64  `form:"sid" validate:"min=1"`
	SortType  int    `form:"sort_type" validate:"min=0"`
	Offset    int64  `form:"offset" default:"0" validate:"min=0"`
	Attribute int64  `form:"attribute" default:"0" validate:"min=0"`
	Zone      int64  `form:"zone"  default:"0" validate:"min=0"`
	Goto      string `form:"goto" default:"dynamic"` //默认dynamic || resource
}

// ParamNewActList .
type ParamNewActList struct {
	Ps       int   `form:"ps" default:"10" validate:"min=1"`
	Sid      int64 `form:"sid" validate:"min=1"`
	SortType int   `form:"sort_type" validate:"min=0"`
	Offset   int64 `form:"offset" default:"0" validate:"min=0"`
	Zone     int64 `form:"zone"  default:"0" validate:"min=0"`
}

// ParamVideoAct .
type ParamVideoAct struct {
	Type int      `form:"type" validate:"min=1"`
	IDs  []string `form:"ids,split" validate:"min=1,max=50"`
}
