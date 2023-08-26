package channel_v2

import (
	"testing"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
	"go-gateway/app/app-svr/app-channel/interface/model"

	"github.com/stretchr/testify/assert"
)

func mineCond(dev pd.Device) *pd.PDContext {
	p := pd.WithDevice(dev).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidG().And().Build(">", 6140000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">", 61400000)
	}).Or().IsPlatIPhoneI().OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidI().And().Build(">", int64(2042030))
	})
	return p
}

func oldCond(plat int8, build int) bool {
	return (model.IsAndroid(plat) && build > 6140000) || (model.IsIPhone(plat) && build > 61400000) || plat == model.PlatIPhoneI || (plat == model.PlatAndroidI && build > 2042030)
}

func TestMineCond(t *testing.T) {
	dev := pd.NewCommonDevice("iphone", "phone", "", 61400001)
	p := mineCond(dev)
	assert.True(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(1, 61400001))

	dev = pd.NewCommonDevice("iphone", "phone", "", 61400000)
	p = mineCond(dev)
	assert.False(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(1, 61400000))

	dev = pd.NewCommonDevice("iphone_i", "phone", "", 61300000)
	p = mineCond(dev)
	assert.True(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(5, 61300000))

	dev = pd.NewCommonDevice("android", "phone", "", 0)
	p = mineCond(dev)
	assert.False(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(0, 6140000))

	dev = pd.NewCommonDevice("android_g", "phone", "", 6130000)
	p = mineCond(dev)
	assert.False(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(4, 6130000))

	dev = pd.NewCommonDevice("android_i", "phone", "", 2042030)
	p = mineCond(dev)
	assert.False(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(8, 2042030))

	dev = pd.NewCommonDevice("android_i", "phone", "", 2042031)
	p = mineCond(dev)
	assert.True(t, p.MustFinish())
	assert.Equal(t, p.MustFinish(), oldCond(8, 2042031))

}
