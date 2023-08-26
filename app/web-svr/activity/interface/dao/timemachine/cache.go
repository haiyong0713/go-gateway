package timemachine

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao -nullcache=&timemachine.UserYearReport2020{Mid:0} -check_null_code=$==nil||$.Mid==0 -sync=true
	UserYearReport2020(c context.Context, mid int64) (*timemachine.UserYearReport2020, error)
	// bts: -struct_name=Dao -nullcache=&timemachine.UserInfo{Mid:0} -check_null_code=$==nil||$.Mid==0 -sync=true
	UserInfoByMid(c context.Context, mid int64) (*timemachine.UserInfo, error)
}
