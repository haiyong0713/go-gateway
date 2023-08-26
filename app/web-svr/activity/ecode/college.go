package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	// ActivityInviterGetErr 获取邀请人失败
	ActivityInviterGetErr = xecode.New(75760)
	// ActivityGetBindCollegeErr 获得绑定学校失败
	ActivityGetBindCollegeErr = xecode.New(75761)
	// ActivityBindCollegeErr 绑定学校失败
	ActivityBindCollegeErr = xecode.New(75762)
	// ActivityGetAllCollegeErr 获取全量学校失败
	ActivityGetAllCollegeErr = xecode.New(75763)
	// ActivityGetProvinceRankErr 获取省排行榜数据失败
	ActivityGetProvinceRankErr = xecode.New(75764)
	// ActivityGetNationwideRankErr 获取全国排行榜数据失败
	ActivityGetNationwideRankErr = xecode.New(75765)
	// ActivityGetProvinceErr 获取省信息失败
	ActivityGetProvinceErr = xecode.New(75766)
	// ActivityCollegeMidInfoErr 获取用户信息失败
	ActivityCollegeMidInfoErr = xecode.New(75767)
	// ActivityCollegeGetErr 获取学院信息失败
	ActivityCollegeGetErr = xecode.New(75768)
	// ActivityCollegeGetTagErr 获取标签信息失败
	ActivityCollegeGetTagErr = xecode.New(75769)
	// ActivityCollegeMidNoBindCollegeErr 用户未绑定学校
	ActivityCollegeMidNoBindCollegeErr = xecode.New(75770)
	// ActivityCollegeMidFolloweErr 关注失败
	ActivityCollegeMidFolloweErr = xecode.New(75771)
	// ActivityCollegeMidNotBindErr 用户未绑定学校
	ActivityCollegeMidNotBindErr = xecode.New(75772)
	// ActivityCollegeInviterCollegeErr 获取邀请人学校失败
	ActivityCollegeInviterCollegeErr = xecode.New(75773)
)
