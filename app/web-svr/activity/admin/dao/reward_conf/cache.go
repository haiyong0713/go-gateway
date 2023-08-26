package reward_conf

import "context"

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao
	AwardConfList(ctx context.Context) ([]int64, error)
}
