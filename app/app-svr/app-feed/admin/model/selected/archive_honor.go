package selected

const (
	ArchiveHonorActionUpdate = "update" // 插入、更新行为
	ArchiveHonorActionDelete = "delete" // 删除行为

	ArchiveHonorTypeWeekly = 2
	//ArchiveHonorDescMustsee = "入站必刷%d大视频" // 入站必刷荣誉稿件展示内容

)

type ArchiveHonor struct {
	Action string `json:"action"` // action分为update和delete
	Aid    int64  `json:"aid"`    // 稿件 id
	Type   int    `json:"type"`   // 稿件类型
	Url    string `json:"url"`    // url 地址
	NaUrl  string `json:"na_url"`
	Desc   string `json:"desc"` // 展示描述
}
