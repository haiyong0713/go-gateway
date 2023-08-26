package model

const (
	DefaultVodAddCategory = "blizzcon_2020"
)

type VodAddStatus int

const (
	VodAddStatusWithdraw VodAddStatus = 0 // 下架
	VodAddStatusPushUp   VodAddStatus = 1 // 上架
)

// VodAddReq /vod/add request
type VodAddReq struct {
	BVID        string       `json:"bvId" form:"bvId" validate:"required"`           // 稿件BVID
	Page        int          `json:"page" form:"page" validate:"min=1"`              // 分P号，只能为1
	Category    string       `json:"category" form:"category" validate:"required"`   // 类别，目前只允许为blizzcon_2020
	Title       string       `json:"title" form:"title" validate:"required"`         // 稿件标题
	Description string       `json:"description" form:"description"`                 // 描述
	Stage       string       `json:"stage" form:"stage" validate:"required"`         // 游戏分类。传tag
	Duration    int64        `json:"duration" form:"duration" validate:"min=1"`      // 时长。单位为秒
	Thumbnail   string       `json:"thumbnail" form:"thumbnail" validate:"required"` // 封面图
	Status      VodAddStatus `json:"status" form:"status"`                           // 上下架状态
	Timestamp   int64        `json:"ts" form:"ts" validate:"min=1"`                  // 接口请求时间。毫秒
	Sign        string       `json:"sign" form:"sign" validate:"required"`           // 签名字符串
}

type VodAddReply struct {
	Status int    `json:"errcode"` // 0为成功
	MSG    string `json:"errmsg"`
}
