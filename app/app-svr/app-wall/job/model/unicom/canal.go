package unicom

import (
	"encoding/json"
)

type CanalMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

// Comment: 联通用户绑定
type UnicomUserBind struct {
	// Comment: 自增ID
	ID int64 `json:"id"`
	// Comment: 手机号码
	Phone string `json:"phone"`
	// Comment: 用户手机伪码，采用net取号加密方式加密
	Usermob string `json:"usermob"`
	// Comment: 用户id
	// Default: 0
	Mid int64 `json:"mid"`
	// Comment: 用户流量
	// Default: 0
	Flow int64 `json:"flow"`
	// Comment: 用户积分
	// Default: 0
	Integral int64 `json:"integral"`
	// Comment: 用户状态 0解除绑定、1绑定
	// Default: 0
	State int64 `json:"state"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime string `json:"ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime string `json:"mtime"`
	// Comment: 上个月更新时间
	// Default: 0000-00-00 00:00:00
	Monthlytime string `json:"monthlytime"`
}

// Comment: 联通礼包
type UnicomUserPacks struct {
	// Comment: 自增ID
	ID int64 `json:"id"`
	// Comment: 礼包类型
	// Default: 0
	Ptype int64 `json:"ptype"`
	// Comment: 礼包描述
	Pdesc string `json:"pdesc"`
	// Comment: 礼包总量
	// Default: 0
	Amount int64 `json:"amount"`
	// Comment: 礼包积分
	// Default: 0
	Integral int64 `json:"integral"`
	// Comment: 0 无上限、1 有上限
	// Default: 0
	Capped int64 `json:"capped"`
	// Comment: 参数
	Param string `json:"param"`
	// Comment: 礼包状态 0失效、1有效
	// Default: 0
	State int64 `json:"state"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime string `json:"ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime string `json:"mtime"`
	// Comment: 礼包原价（相对integral现价）
	Original int64 `json:"original"`
	// Comment: 礼包类型 0福利包,1流量包
	// Default: 0
	Kind int64 `json:"kind"`
	// Comment: 礼包封面
	Cover string `json:"cover"`
}

// Comment: 联通信息同步
type UnicomOrder struct {
	// Comment: 自增ID
	ID int64 `json:"id"`
	// Comment: 用户手机伪码，采用net取号加密方式加密
	Usermob string `json:"usermob"`
	// Comment: 内容提供商ID
	Cpid string `json:"cpid"`
	// Comment: Sp业务ID
	Spid int64 `json:"spid"`
	// Comment: 操作类型 0 订购，1 退订， 2 体验卡订购， 3 WO卡订购
	Type int64 `json:"type"`
	// Comment: 订购时间 时间格式
	// Default: 0000-00-00 00:00:00
	Ordertime string `json:"ordertime"`
	// Comment: 退订时间
	// Default: 0000-00-00 00:00:00
	Canceltime string `json:"canceltime"`
	// Comment: 失效时间
	// Default: 0000-00-00 00:00:00
	Endtime string `json:"endtime"`
	// Comment: 推广渠道编号
	Channelcode int64 `json:"channelcode"`
	// Comment: 用户所属省份，中文名称
	Province string `json:"province"`
	// Comment: 用户所属地市，中文名称
	Area string `json:"area"`
	// Comment: 订购类型 0 按内容订购， 1 包月订购
	Ordertype int64 `json:"ordertype"`
	// Comment: 视频编码
	Videoid string `json:"videoid"`
	// Comment: 当前用户使用流量统计的截止时间
	// Default: 0000-00-00 00:00:00
	Time string `json:"time"`
	// Comment: 用户使用流量(单位:KB,流量为time表示的月份流量)
	// Default: 0
	Flowbyte int64 `json:"flowbyte"`
	// Comment: 创建时间
	// Default: 0000-00-00 00:00:00
	Ctime string `json:"ctime"`
	// Comment: 修改时间
	// Default: 0000-00-00 00:00:00
	Mtime string `json:"mtime"`
}

// Comment: 移动信息同步
type MobileOrder struct {
	// Comment: 自增ID
	ID int64 `json:"id"`
	// Comment: 订单编号，互联网计费平台内部订购唯一标识
	Orderid string `json:"orderid"`
	// Comment: 用户伪码
	Userpseudocode string `json:"userpseudocode"`
	// Comment: 外部交易ID
	Channelseqid string `json:"channelseqid"`
	// Comment: 业务资费(单位：分)
	Price int64 `json:"price"`
	// Comment: 操作时间
	// Default: 0000-00-00 00:00:00
	Actiontime string `json:"actiontime"`
	// Comment: 订购状态
	Actionid int64 `json:"actionid"`
	// Comment: 生效时间
	// Default: 0000-00-00 00:00:00
	Effectivetime string `json:"effectivetime"`
	// Comment: 失效时间
	// Default: 0000-00-00 00:00:00
	Expiretime string `json:"expiretime"`
	// Comment: 渠道合作方编码
	Channelid string `json:"channelid"`
	// Comment: 产品编码
	Productid string `json:"productid"`
	// Comment: 订购类型 0 测试 1 正式
	Ordertype int64 `json:"ordertype"`
	// Comment: 剩余流量占比
	Threshold int64 `json:"threshold"`
	// Comment: 统计时间，省公司统计阀值的时间
	// Default: 0000-00-00 00:00:00
	Resulttime string `json:"resulttime"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime string `json:"ctime"`
	// Comment: 修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime string `json:"mtime"`
}

// Comment: 联通免流伪码信息表
type UnicomUsermobInfo struct {
	// Comment: 自增ID
	ID int64 `json:"id"`
	// Comment: 用户手机伪码，采用net取号加密方式加密
	Usermob string `json:"usermob"`
	// Comment: fake_id
	FakeID string `json:"fake_id"`
	// Comment: 当前 fake_id 所属周期
	Period int64 `json:"period"`
	// Comment: 生成 fake_id 月份
	Month string `json:"month"`
	// Comment: 创建时间
	Ctime string `json:"ctime"`
	// Comment: 修改时间
	Mtime string `json:"mtime"`
}
