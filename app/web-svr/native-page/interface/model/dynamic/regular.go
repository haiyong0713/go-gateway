package dynamic

import (
	"encoding/json"
	"path"
	"regexp"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/api"
)

const (
	TsCompareNoChange = 0 //无改动
	TsCompareAuto     = 1 //自动送审
	TsCompareManual   = 2 //人工送审
)

func AdminCheck(ver string) bool {
	return strings.Contains(ver, "admin")
}

func MetaCheck(meta string, width, length int32) error {
	switch strings.TrimPrefix(path.Ext(meta), ".") {
	case "jpg", "jpeg", "png":
	default:
		return ecode.Error(ecode.RequestErr, "背景图片不合法")
	}
	// 宽高比大于等于22:75（即宽高比1125*330的头图大小），小于等于1：1
	per := float64(width) / float64(length)
	leftPre := float64(75) / float64(22)
	if per > leftPre || per < 1.0 {
		return ecode.Error(ecode.RequestErr, "图片宽高比例必须在75：22与1：1之间")
	}
	return nil
}

// ModuleComparison .
// bool : 是否需要送审
// bool 是否无需送审编辑组件信息
func ModuleComparison(oldM, newM []*NativeTsModuleExt) int {
	// 没有组件
	if len(oldM) == 0 && len(newM) == 0 {
		return TsCompareNoChange
	}
	rst := TsCompareNoChange
	onlyMap := make(map[string]*api.NativeTsModule)
	for _, v := range oldM {
		categoryTmp := &api.NativeModule{Category: int64(v.Category)}
		switch {
		case categoryTmp.IsStatement(): //文本组件
			onlyMap[v.Ukey] = &api.NativeTsModule{Remark: v.Remark}
		case categoryTmp.IsClick(): //自定义点击组件
			// 只会有一个自定义点击组件
			onlyMap[v.Ukey] = &api.NativeTsModule{Meta: v.Meta, Width: v.Width, Length: v.Length}
		case categoryTmp.IsResourceID(), categoryTmp.IsNewVideoID(), categoryTmp.IsActCapsule(), categoryTmp.IsRecommend():
			rst = TsCompareAuto
		case categoryTmp.IsCarouselImg():
			onlyMap[v.Ukey] = func() *api.NativeTsModule {
				if len(v.Resources) == 0 || v.Resources[0].Ext == "" {
					return &api.NativeTsModule{}
				}
				rawExt := &ResourceExt{}
				if err := json.Unmarshal([]byte(v.Resources[0].Ext), rawExt); err != nil {
					log.Error("Fail to unmarshal ResourceExt, ext=%s error=%+v", v.Resources[0].Ext, err)
					return &api.NativeTsModule{}
				}
				return &api.NativeTsModule{Meta: rawExt.ImgUrl, Width: rawExt.Width, Length: rawExt.Length}
			}()
		default: //其余组件暂不支持
			continue
		}
	}
	for _, v := range newM {
		categoryTmp := &api.NativeModule{Category: int64(v.Category)}
		if categoryTmp.IsResourceID() || categoryTmp.IsNewVideoID() || categoryTmp.IsActCapsule() || categoryTmp.IsRecommend() {
			rst = TsCompareAuto
			continue
		}
		oldV, ok := onlyMap[v.Ukey]
		if !ok {
			return TsCompareManual
		}
		switch {
		case categoryTmp.IsStatement(): //文本组件
			if v.Remark != oldV.Remark {
				return TsCompareManual
			}
		case categoryTmp.IsClick(), categoryTmp.IsCarouselImg(): //自定义点击组件
			// 只会有一个自定义点击组件
			if v.Meta != oldV.Meta || v.Width != oldV.Width || v.Length != oldV.Length {
				return TsCompareManual
			}
		default: //其余组件暂不支持
			continue
		}
	}
	return rst
}

func ParseUserAgent2MobiApp(ua string) string {
	if ua == "" {
		return ""
	}
	if matched, err := regexp.Match(`iPhone`, []byte(ua)); err == nil && matched {
		return "iphone"
	}

	if matched, err := regexp.Match(`android`, []byte(ua)); err == nil && matched {
		return "android"
	}
	return ""
}
