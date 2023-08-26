package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
)

func TestPreferTraditionalChinese(t *testing.T) {
	asTC := []string{"zh-Hant", "zh-Hant_TW", "zh-Hant_HK", "zh-Hant_MO", "zh_TW", "zh_MO", "zh_HK"}
	for _, i := range asTC {
		assert.True(t, PreferTraditionalChinese(context.Background(), i, ""), i)
	}
	for _, i := range asTC {
		ctx := device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone", Build: 61500000})
		assert.False(t, PreferTraditionalChinese(ctx, i, ""), i)
	}
	for _, i := range asTC {
		ctx := device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone", Build: 61600000})
		assert.True(t, PreferTraditionalChinese(ctx, i, ""), i)
	}

	asSC := []string{"zh-Hans", "zh-Hans_CN", "zh_CN"}
	for _, i := range asSC {
		assert.False(t, PreferTraditionalChinese(context.Background(), i, ""), i)
	}
	for _, i := range asSC {
		ctx := device.NewContext(context.Background(), device.Device{RawMobiApp: "iphone", Build: 61600000})
		assert.False(t, PreferTraditionalChinese(ctx, i, ""), i)
	}
}

func TestConstructLocaleID(t *testing.T) {
	assert.Equal(t, "zh_US", concatLocalID(locale.LocaleIDs{
		Language: "zh",
		Region:   "US",
	}))
	assert.Equal(t, "zh-Hans_US", concatLocalID(locale.LocaleIDs{
		Language: "zh",
		Script:   "Hans",
		Region:   "US",
	}))
}
