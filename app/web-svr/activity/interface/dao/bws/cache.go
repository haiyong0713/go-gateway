package bws

import (
	"context"
	"fmt"

	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

func midKey(bid, mid int64) string {
	return fmt.Sprintf("u_m_%d_%d", bid, mid)
}

func keyKey(bid int64, key string) string {
	return fmt.Sprintf("u_k_%d_%s", bid, key)
}

func usersVipKey(bid int64, vipKey string) string {
	return fmt.Sprintf("b_v_%d_%s", bid, vipKey)
}

func usersVipMidDateKey(bid int64, mid int64, date string) string {
	return fmt.Sprintf("b_m_%d_%d_%s", bid, mid, date)

}

func bwsPointsKey(id int64) string {
	return fmt.Sprintf("bws_pt_%d", id)
}
func pointsKey(id int64) string {
	return fmt.Sprintf("b_p_b_%d", id)
}

func rechargeLevelKey(id int64) string {
	return fmt.Sprintf("b_re_l_%d", id)
}

func rechargeAwardKey(id int64) string {
	return fmt.Sprintf("b_aw_l_%d", id)
}

func achievesKey(id int64) string {
	return fmt.Sprintf("b_a_%d", id)
}

func bwsSignKey(id int64) string {
	return fmt.Sprintf("b_p_si_%d", id)
}

func pointSignKey(pid int64) string {
	return fmt.Sprintf("b_set_si_%d", pid)
}

func fieldsListKey(bid int64) string {
	return fmt.Sprintf("b_fies_%d", bid)
}

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -sync=true -struct_name=Dao
	UsersMid(c context.Context, bid int64, key int64) (*bwsmdl.Users, error)
	// bts: -sync=true -struct_name=Dao
	UsersKey(c context.Context, bid int64, ukey string) (*bwsmdl.Users, error)
	// bts: -sync=true -struct_name=Dao
	Points(c context.Context, bid int64) ([]int64, error)
	// bts: -sync=true -struct_name=Dao
	BwsPoints(c context.Context, ids []int64) (map[int64]*bwsmdl.Point, error)
	// bts: -sync=true -struct_name=Dao
	BwsSign(c context.Context, ids []int64) (map[int64]*bwsmdl.PointSign, error)
	// bts: -sync=true -struct_name=Dao
	Signs(c context.Context, pid int64) ([]int64, error)
	// bts: -sync=true -struct_name=Dao
	Achievements(c context.Context, bid int64) (*bwsmdl.Achievements, error)
	// bts: -sync=true -struct_name=Dao
	UserAchieves(c context.Context, bid int64, key string) ([]*bwsmdl.UserAchieve, error)
	// bts: -sync=true -struct_name=Dao
	UserPoints(c context.Context, bid int64, key string) ([]*bwsmdl.UserPoint, error)
	// bts: -sync=true -struct_name=Dao
	UserLockPoints(c context.Context, bid int64, lockType int64, ukey string) ([]*bwsmdl.UserPoint, error)
	// bts: -sync=true -struct_name=Dao
	UserLockPointsDay(c context.Context, bid int64, lockType int64, ukey string, day string) ([]*bwsmdl.UserPoint, error)
	// bts: -sync=true -struct_name=Dao
	AchieveCounts(c context.Context, bid int64, day string) ([]*bwsmdl.CountAchieves, error)
	// bts: -sync=true -struct_name=Dao
	RechargeLevels(c context.Context, ids []int64) (map[int64]*bwsmdl.PointsLevel, error)
	// bts: -sync=true -struct_name=Dao
	RechargeAwards(c context.Context, ids []int64) (map[int64]*bwsmdl.PointsAward, error)
	// bts: -sync=true -struct_name=Dao
	PointLevels(c context.Context, bid int64) ([]int64, error)
	// bts: -sync=true -struct_name=Dao
	PointsAward(c context.Context, plID int64) ([]int64, error)
	// bts: -sync=true -struct_name=Dao
	CompositeAchievesPoint(c context.Context, mids []int64) (map[int64]int64, error)
	// bts: -sync=true -struct_name=Dao
	ActFields(c context.Context, bid int64) (*bwsmdl.ActFields, error)
	// bts: -struct_name=Dao -nullcache=[]*bwsmdl.UserTask{{TaskID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].TaskID==-1
	UserTasks(c context.Context, userToken string, day int64) ([]*bwsmdl.UserTask, error)
	// bts: -struct_name=Dao -nullcache=[]*bwsmdl.UserAward{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1
	UserAward(c context.Context, userToken string) ([]*bwsmdl.UserAward, error)
	// bts: -struct_name=Dao -nullcache=-1 -check_null_code=$==-1
	UserUnFinishVoteID(c context.Context, userToken string, pid int64) (int64, error)
	// bts: -struct_name=Dao -nullcache=-1 -check_null_code=$==-1
	UserLotteryTimes(c context.Context, userToken string) (int64, error)
	// bts: -sync=true -struct_name=Dao
	UserDetail(c context.Context, bid int64, mid int64, date string) (*bwsmdl.UserDetail, error)
	// bts: -sync=true -struct_name=Dao
	UserDetails(c context.Context, mids []int64, bid int64, date string) (map[int64]*bwsmdl.UserDetail, error)
	// bts: -sync=true -struct_name=Dao
	UsersVipKey(c context.Context, bid int64, ukey string) (*bwsmdl.VipUsersToken, error)
	// bts: -sync=true -struct_name=Dao
	UsersVipMidDate(c context.Context, bid int64, mid int64, date string) (*bwsmdl.VipUsersToken, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	//mc: -key=midKey -struct_name=Dao
	CacheUsersMid(c context.Context, bid int64, mid int64) (*bwsmdl.Users, error)
	//mc: -key=midKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheUsersMid(c context.Context, bid int64, value *bwsmdl.Users, mid int64) error
	//mc: -key=midKey -struct_name=Dao
	DelCacheUsersMid(c context.Context, bid int64, mid int64) error
	//mc: -key=keyKey -struct_name=Dao
	CacheUsersKey(c context.Context, bid int64, userKey string) (*bwsmdl.Users, error)
	//mc: -key=keyKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheUsersKey(c context.Context, bid int64, value *bwsmdl.Users, userKey string) error
	//mc: -key=keyKey -struct_name=Dao
	DelCacheUsersKey(c context.Context, bid int64, userKey string) error
	//mc: -key=pointsKey -struct_name=Dao
	CachePoints(c context.Context, key int64) ([]int64, error)
	//mc: -key=pointsKey -expire=d.mcExpire -encode=json -struct_name=Dao
	AddCachePoints(c context.Context, key int64, value []int64) error
	//mc: -key=pointsKey -struct_name=Dao
	DelCachePoints(c context.Context, key int64) error
	//mc: -key=bwsSignKey -struct_name=Dao
	CacheBwsSign(c context.Context, ids []int64) (map[int64]*bwsmdl.PointSign, error)
	//mc: -key=bwsSignKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheBwsSign(c context.Context, val map[int64]*bwsmdl.PointSign) error
	//mc: -key=bwsSignKey -struct_name=Dao
	DelCacheBwsSign(c context.Context, ids []int64) error
	//mc: -key=pointSignKey -struct_name=Dao
	CacheSigns(c context.Context, pid int64) ([]int64, error)
	//mc: -key=pointSignKey -expire=d.mcExpire -encode=json -struct_name=Dao
	AddCacheSigns(c context.Context, pid int64, val []int64) error
	//mc: -key=pointSignKey -struct_name=Dao
	DelCacheSigns(c context.Context, pid int64) error
	//mc: -key=bwsPointsKey -struct_name=Dao
	CacheBwsPoints(c context.Context, ids []int64) (map[int64]*bwsmdl.Point, error)
	//mc: -key=bwsPointsKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheBwsPoints(c context.Context, val map[int64]*bwsmdl.Point) error
	//mc: -key=bwsPointsKey -struct_name=Dao
	DelCacheBwsPoints(c context.Context, ids []int64) error
	//mc: -key=achievesKey -struct_name=Dao
	CacheAchievements(c context.Context, key int64) (*bwsmdl.Achievements, error)
	//mc: -key=achievesKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheAchievements(c context.Context, key int64, value *bwsmdl.Achievements) error
	//mc: -key=achievesKey -struct_name=Dao
	DelCacheAchievements(c context.Context, key int64) error
	//mc: -key=rechargeLevelKey -struct_name=Dao
	CacheRechargeLevels(c context.Context, ids []int64) (map[int64]*bwsmdl.PointsLevel, error)
	//mc: -key=rechargeLevelKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheRechargeLevels(c context.Context, val map[int64]*bwsmdl.PointsLevel) error
	//mc: -key=rechargeLevelKey -struct_name=Dao
	DelCacheRechargeLevels(c context.Context, ids []int64) error
	//mc: -key=rechargeAwardKey -struct_name=Dao
	CacheRechargeAwards(c context.Context, ids []int64) (map[int64]*bwsmdl.PointsAward, error)
	//mc: -key=rechargeAwardKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheRechargeAwards(c context.Context, val map[int64]*bwsmdl.PointsAward) error
	//mc: -key=rechargeAwardKey -struct_name=Dao
	DelCacheRechargeAwards(c context.Context, ids []int64) error
	//mc: -key=fieldsListKey -struct_name=Dao
	CacheActFields(c context.Context, bid int64) (*bwsmdl.ActFields, error)
	//mc: -key=fieldsListKey -expire=d.mcItemExpire -encode=pb -struct_name=Dao
	AddCacheActFields(c context.Context, bid int64, val *bwsmdl.ActFields) error
	//mc: -key=usersVipKey -struct_name=Dao
	CacheUsersVipKey(c context.Context, bid int64, userKey string) (*bwsmdl.VipUsersToken, error)
	//mc: -key=usersVipKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheUsersVipKey(c context.Context, bid int64, value *bwsmdl.VipUsersToken, userKey string) error
	//mc: -key=usersVipKey -struct_name=Dao
	DelCacheUsersVipKey(c context.Context, bid int64, userKey string) error
	//mc: -key=usersVipMidDateKey -struct_name=Dao
	CacheUsersVipMidDate(c context.Context, bid int64, mid int64, date string) (*bwsmdl.VipUsersToken, error)
	//mc: -key=usersVipMidDateKey -expire=d.mcExpire -encode=pb -struct_name=Dao
	AddCacheUsersVipMidDate(c context.Context, bid int64, value *bwsmdl.VipUsersToken, mid int64, date string) error
	//mc: -key=usersVipMidDateKey -struct_name=Dao
	DelCacheUsersVipMidDate(c context.Context, bid int64, mid int64, date string) error
}
