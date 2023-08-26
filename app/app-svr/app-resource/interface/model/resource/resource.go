package static

// Resource 资源
type Resource struct {
	Type       string          `json:"type"`        //资源业务类型
	List       []*ResourceItem `json:"list"`        //资源列表
	ExtraValue string          `json:"extra_value"` //资源业务额外信息
}

type ResourceItem struct {
	TaskId         string `json:"task_id"`     //业务任务ID
	FileName       string `json:"file_name"`   //业务资源文件名
	Type           string `json:"type"`        //业务资源细分类型
	URL            string `json:"url"`         //资源下载链接
	Hash           string `json:"hash"`        //文件hash值
	Size           int    `json:"size"`        //文件大小
	ExpectDwType   int    `json:"expect_dw"`   //资源下载方式,1为cdn,2为pcdn
	FileEffectTime int64  `json:"effect_time"` //资源生效时间
	FileExpireTime int64  `json:"expire_time"` //资源过期时间
	FileUploadTime int64  `json:"-"`           //资源上传时间
	Priority       int    `json:"priority"`    //下载优先级
}

type DwTime struct {
	Type int8           `json:"type"` //域名类型，1:cdn, 2:pcdn
	Peak []*DwTimePiece `json:"peak"` //下载高峰时间段
	Low  []*DwTimePiece `json:"low"`  //下载低峰时间段
}

type DwTimePiece struct {
	Start int64 `json:"start"` //起始时间
	End   int64 `json:"end"`   //结束时间
}

type ResourceDownloadRequest struct {
	MobiApp  string `json:"mobi_app"`
	Device   string `json:"device"`
	Build    int    `json:"build"`
	Mid      int64  `json:"mid"`
	Buvid    string `json:"buvid"`
	Platform int8   `json:"platform"`
	Type     string `json:"type"`
	Ver      string `json:"ver"`
}

type ResourceDownloadResponse struct {
	Ver      string             `json:"ver"`      //资源业务版本号(根据item文件hash聚合)
	Resource []*Resource        `json:"resource"` //资源业务类型
	Dwtime   map[string]*DwTime `json:"dwtime"`   //下载高峰时间段
}

type CdnDwTime struct {
	Peak []*DwTimePiece `json:"peak"` //下载高峰时间段
	Low  []*DwTimePiece `json:"low"`  //下载低峰时间段
}

type ResourceTaskId struct {
	FileName   string `json:"file_name"`   //文件名
	UploadTime int64  `json:"upload_time"` //上传时间
}
