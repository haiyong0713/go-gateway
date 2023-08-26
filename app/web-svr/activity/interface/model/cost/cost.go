package cost

import go_common_library_time "go-common/library/time"

// UserCostInfoDB user_cost_info表结构
type UserCostInfoDB struct {
	ID         int64                       `json:"id"`
	Mid        int64                       `json:"mid"`
	OrderId    string                      `json:"order_id"`
	AwardId    string                      `json:"award_id"`
	ActivityId string                      `json:"activity_id"`
	CostType   int                         `json:"cost_type"`
	CostValue  int                         `json:"cost_value"`
	Status     int                         `json:"status"`
	Ctime      go_common_library_time.Time `json:"ctime"`
	Mtime      go_common_library_time.Time `json:"ctime"`
}

// AwardConfigDataDB 奖品配置表
type AwardConfigDataDB struct {
	ID         int64                       `json:"id"`
	ActivityId string                      `json:"activity_id"`
	AwardId    string                      `json:"award_id"`
	StockId    int64                       `json:"stock_id"`
	CostType   int                         `json:"cost_type"`
	CostValue  int                         `json:"cost_value"`
	ShowTime   go_common_library_time.Time `json:"show_time"`
	EndTime    go_common_library_time.Time `json:"end_time"`
	Order      int                         `json:"order"`
	Creator    string                      `json:"creator"`
	Status     int                         `json:"status"`
	Ctime      go_common_library_time.Time `json:"ctime"`
	Mtime      go_common_library_time.Time `json:"mtime"`
}
