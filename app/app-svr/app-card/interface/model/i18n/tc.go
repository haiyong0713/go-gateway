package i18n

import (
	"context"
	"fmt"

	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
	"go-common/library/text/translate/chinese"
	chineseV2 "go-common/library/text/translate/chinese.v2"

	feature "go-gateway/app/app-svr/feature/service/sdk"

	"golang.org/x/text/language"
)

func concatLocalID(in locale.LocaleIDs) string {
	if in.Script == "" {
		return fmt.Sprintf("%s_%s", in.Language, in.Region)
	}
	return fmt.Sprintf("%s-%s_%s", in.Language, in.Script, in.Region)
}

func ConstructLocaleID(in locale.Locale) (string, string) {
	slocale := concatLocalID(in.SLocale)
	clocale := concatLocalID(in.CLocale)
	return slocale, clocale
}

func tcTranslateRequired(dev *device.Device) bool {
	return (dev.RawMobiApp == "android_i" && dev.Build >= 3000000) || (dev.IsAndroid() && (dev.Build >= 6160000)) || (dev.IsIOS() && (dev.Build >= 61600000))
}

func PreferTraditionalChinese(ctx context.Context, slocale, clocale string) bool {
	if dev, ok := device.FromContext(ctx); ok {
		if !tcTranslateRequired(&dev) {
			return false
		}
	}
	tag, err := language.Parse(slocale)
	if err != nil {
		return false
	}
	script, _ := tag.Script()
	return script.String() == "Hant"
}

func PreferTraditionalChineseV2(ctx context.Context, keyName, slocale, clocale string) bool {
	if dev, ok := device.FromContext(ctx); ok {
		if !feature.GetBuildLimit(ctx, keyName, &feature.OriginResutl{
			BuildLimit: tcTranslateRequired(&dev),
		}) {
			return false
		}
	}
	tag, err := language.Parse(slocale)
	if err != nil {
		return false
	}
	script, _ := tag.Script()
	return script.String() == "Hant"
}

func TranslateAsTC(in ...*string) {
	for _, v := range in {
		*v = chinese.Convert(context.Background(), *v)
	}
}
func TranslateAsTCV2(in ...*string) {
	for _, v := range in {
		*v = chineseV2.Convert(context.Background(), *v)
	}
}
