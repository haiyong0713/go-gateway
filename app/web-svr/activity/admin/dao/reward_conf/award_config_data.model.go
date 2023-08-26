package reward_conf

import (
	"go-common/library/time"
)

// AwardConfigData represents a row from 'award_config_data'.
type AwardConfigData struct {
	ID         uint64    `json:"id"`          // 主键
	AwardID    string    `json:"award_id"`    // 奖品id/奖池id
	StockID    int32     `json:"stock_id"`    // 库存id
	CostType   int8      `json:"cost_type"`   // 是否有效 1=抽奖 2=积分兑换
	CostValue  int32     `json:"cost_value"`  // 消耗数量
	ShowTime   time.Time `json:"show_time"`   // 展示时间
	Order      int32     `json:"order"`       // 排序
	Creator    string    `json:"creator"`     // 创建人
	Status     int8      `json:"status"`      // 是否有效 1=有效 0=无效
	Ctime      time.Time `json:"ctime"`       // 创建时间
	Mtime      time.Time `json:"mtime"`       // 修改时间
	ActivityID string    `json:"activity_id"` // 活动唯一标识
	EndTime    time.Time `json:"end_time"`    // 展示时间
}
