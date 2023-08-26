package midInt64

import (
	"context"
	"math"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

func CheckHasInt64InMids(input ...int64) bool {
	for _, v := range input {
		if v > math.MaxInt32 {
			return true
		}
	}
	return false
}

func IsDisableInt64MidVersion(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidI().Or().IsPlatAndroidB().And().Build("<", int64(6500000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().Or().IsMobiAppIPhoneI().Or().IsPlatIPhoneB().And().Build("<", int64(65000000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33000000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build("<", int64(1070000))
	}).MustFinish()
}
