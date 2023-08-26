package timemachine

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"
)

//go:generate kratos tool redisgen

type _redis interface {
	// redis: -key=userYearReport2020Key -struct_name=Dao
	CacheUserYearReport2020(c context.Context, mid int64) (*timemachine.UserYearReport2020, error)
	// redis: -key=userYearReport2020Key -expire=d.UserYearReport2020Expire -encode=json -struct_name=Dao -check_null_code=$==nil||$.Mid==0
	AddCacheUserYearReport2020(c context.Context, mid int64, value *timemachine.UserYearReport2020) error
	// redis: -key=userYearReport2020UserInfoKey -struct_name=Dao
	CacheUserInfoByMid(c context.Context, mid int64) (*timemachine.UserInfo, error)
	// redis: -key=userYearReport2020UserInfoKey -expire=d.UserYearReport2020Expire -encode=json -struct_name=Dao -check_null_code=$==nil||$.Mid==0
	AddCacheUserInfoByMid(c context.Context, mid int64, value *timemachine.UserInfo) error
	// redis: -key=userYearReport2020UserInfoKey -struct_name=Dao
	DelCacheUserInfoByMid(c context.Context, mid int64) error
}

func userYearReport2020Key(mid int64) string {
	return fmt.Sprintf("a_u_y_r_2020_%d", mid)
}

func userYearReport2020UserInfoKey(mid int64) string {
	return fmt.Sprintf("a_u_y_r_2020_u_i_%d", mid)
}
