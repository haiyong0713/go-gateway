package midInt64

import (
	"context"
	"testing"

	"go-common/component/metadata/device"

	"github.com/stretchr/testify/assert"
)

var (
	limitedCtx1 = device.NewContext(context.Background(), device.Device{RawMobiApp: "android", Build: 6490000})
	limitedCtx2 = device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone", Build: 64900000})
	limitedCtx3 = device.NewContext(context.Background(), device.Device{RawMobiApp: "ipad", Build: 32900000})
	limitedCtx4 = device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone_i", Build: 64900000})
	limitedCtx5 = device.NewContext(context.Background(), device.Device{RawMobiApp: "android_i", Build: 6490000})
	limitedCtx6 = device.NewContext(context.Background(), device.Device{RawMobiApp: "android_hd", Build: 1049999})

	ctx1 = device.NewContext(context.Background(), device.Device{RawMobiApp: "android", Build: 6500000})
	ctx2 = device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone", Build: 65000000})
	ctx3 = device.NewContext(context.Background(), device.Device{RawMobiApp: "ipad", Build: 33000000})
	ctx4 = device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone_i", Build: 65000000})
	ctx5 = device.NewContext(context.Background(), device.Device{RawMobiApp: "android_i", Build: 6500000})
	ctx6 = device.NewContext(context.Background(), device.Device{RawMobiApp: "android_hd", Build: 1070000})
)

func TestFilterMidMaxInt32(t *testing.T) {
	assert.Equal(t, true, IsDisableInt64MidVersion(limitedCtx1))
	assert.Equal(t, true, IsDisableInt64MidVersion(limitedCtx2))
	assert.Equal(t, true, IsDisableInt64MidVersion(limitedCtx3))
	assert.Equal(t, true, IsDisableInt64MidVersion(limitedCtx4))
	assert.Equal(t, true, IsDisableInt64MidVersion(limitedCtx5))
	assert.Equal(t, true, IsDisableInt64MidVersion(limitedCtx6))

	assert.Equal(t, false, IsDisableInt64MidVersion(ctx1))
	assert.Equal(t, false, IsDisableInt64MidVersion(ctx2))
	assert.Equal(t, false, IsDisableInt64MidVersion(ctx3))
	assert.Equal(t, false, IsDisableInt64MidVersion(ctx4))
	assert.Equal(t, false, IsDisableInt64MidVersion(ctx5))
	assert.Equal(t, false, IsDisableInt64MidVersion(ctx6))

	assert.Equal(t, true, CheckHasInt64InMids(0, 2147483648))
	assert.Equal(t, false, CheckHasInt64InMids(0, 2147483647))
}
