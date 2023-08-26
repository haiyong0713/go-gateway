package service

import (
	"context"

	"go-common/component/metadata/device"
)

// 是否支持推荐播单翻页模式
func (s *Service) isRcmdOffsetCapable(ctx context.Context) bool {
	dev, ok := device.FromContext(ctx)
	if !ok {
		return false
	}
	const (
		_androidLimit = 6880000
		_iosLimit     = 68800000
	)
	return (dev.Plat() == device.PlatAndroid && dev.Build >= _androidLimit) || (dev.Plat() == device.PlatIPhone && dev.Build >= _iosLimit)
}
