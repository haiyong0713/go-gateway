package generator

type CodeInfo struct {
	TmplFilePath         string
	GoFilePath           string
	ProjectName          string
	ServiceName          string
	ServicePath          string
	ServiceProtoFile     string
	ServiceInterfaceInfo *ServiceInterfaceInfo
	TaskFlowInfo         *TaskFlowInfo
	RpcClientMap         map[string]*RpcClientInfo
	HttpRequestMap       map[string]*HttpRequestInfo
}

type TaskFlowInfo struct {
	TaskProtoFile     string
	TaskUdfFile       string
	TaskList          []*TaskInfo
	ServiceInputType  string
	ServiceOutputType string
}

type TaskDSL struct {
	TaskList []*TaskInfo `json:"taskList"`
}

type TaskInfo struct {
	Name         string               `json:"name"`
	Type         string               `json:"type"`
	InputList    []*TaskInputInfo     `json:"input"`
	OutputInfo   *TaskOutputInfo      `json:"output"`
	DiscoveryId  string               `json:"discoveryId,omitempty"`
	RpcInterface string               `json:"rpcInterface,omitempty"`
	MappingRule  []*MappingRuleDetail `json:"mappingRule,omitempty"`
	Url          string               `json:"url,omitempty"`
	Method       string               `json:"method,omitempty"`
	Query        []*MappingRuleDetail `json:"query,omitempty"`
	UrlBody      []*MappingRuleDetail `json:"urlBody,omitempty"`
	JsonBody     []*MappingRuleDetail `json:"jsonBody,omitempty"`
	Header       []*MappingRuleDetail `json:"header,omitempty"`
	ImportPb     map[string]string
}

type MappingRuleDetail struct {
	Src       string `json:"src"`
	Dest      string `json:"dest"`
	MapFunc   string `json:"mapFunc,omitempty"`
	From      string
	To        string
	SrcObject string
}

type TaskInputInfo struct {
	Name      string `json:"name,omitempty"`
	RowValue  string `json:"value"`
	Ignore    bool   `json:"ignore"`
	PreTask   string
	ValueFrom string
	Value     string
	ParamType string
}

type TaskOutputInfo struct {
	Name       string `json:"name,omitempty"`
	RowType    string `json:"type"`
	IsResponse bool   `json:"isResponse"`
	ReturnType string
}
