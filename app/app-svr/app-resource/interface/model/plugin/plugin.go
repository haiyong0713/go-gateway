package plugin

type Plugin struct {
	Name     string `json:"name"`
	Package  string `json:"package"`
	Policy   int8   `json:"policy"`
	VerCode  int64  `json:"ver_code"`
	VerName  string `json:"ver_name"`
	Size     int64  `json:"size"`
	MD5      string `json:"md5"`
	URL      string `json:"url"`
	Enable   bool   `json:"enable"`
	Force    bool   `json:"force"`
	Clear    bool   `json:"clear"`
	MinBuild int    `json:"min_build"`
	MaxBuild int    `json:"max_build"`
	BaseCode int    `json:"base_code"`
	BaseName string `json:"base_name"`
	Desc     string `json:"desc"`
	Coverage int    `json:"-"`
}

type Plugins []*Plugin

func (ps Plugins) Len() int           { return len(ps) }
func (ps Plugins) Less(i, j int) bool { return ps[i].VerCode > ps[j].VerCode }
func (ps Plugins) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

type TraceParam struct {
	Service   string `form:"service"`
	Operation string `form:"operation"`
	Cookie    string
}

type DependenceReply struct {
	OperationName []string `json:"operation_name"`
}

type TraceEdge struct {
	SrcService     string            `json:"src_service"`
	DstService     string            `json:"dst_service"`
	OperationTable []*OperationTable `json:"operation_table"`
}

type OperationTable struct {
	SrcOperation string `json:"src_operation"`
	DstOperation string `json:"dst_operation"`
}
