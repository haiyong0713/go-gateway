package service

import (
	"context"
	"fmt"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/exp/ab"
	"go-common/library/log"
)

var abtestFlag = map[string]*ab.StringFlag{
	"background":      ab.String("player_background_exp", "后台播放abtest实验", _missValue),
	"ipad_background": ab.String("hd_switch_type", "ipad后台播放实验", _missValue),
}

type Experiment interface {
	Exp(ctx context.Context)
	GetResultAfterExp() interface{}
}

type BoolValue struct {
	Value bool
}

func (b BoolValue) GetResultAfterExp() interface{} {
	return b.Value
}

type EnumValue struct {
	Value int64
}

func (e EnumValue) GetResultAfterExp() interface{} {
	return e.Value
}

type BackgroundWithExp struct {
	BoolValue
	LastModified int64
}

func (b *BackgroundWithExp) Exp(ctx context.Context) {
	dev, _ := device.FromContext(ctx)
	if dev.RawMobiApp == "iphone" && dev.Build < 66300000 { //ios 662不做实验
		return
	}
	if b.LastModified != 0 {
		return
	}
	func() {
		//pad新版本做pad新实验，不是pad则继续执行之前粉版的老实验
		if padCanOutputPlayConf(ctx) {
			if GetExpConfigFromContext(ctx).padIsNewDevice && abTestRun(ctx, abtestFlag["ipad_background"]) {
				b.Value = true
			}
			return
		}
		if GetExpConfigFromContext(ctx).isNewDevice && abTestRun(ctx, abtestFlag["background"]) {
			b.Value = true
		}
	}()
}

type ColorFilterWithExp struct {
	EnumValue
	LastModified int64
}

func (b *ColorFilterWithExp) Exp(ctx context.Context) {
}

type SubtitleWithExp struct {
	BoolValue
	LastModified int64
}

func (b *SubtitleWithExp) Exp(ctx context.Context) {

}

type DolbyWithExp struct {
	BoolValue
	LastModified int64
}

func (b *DolbyWithExp) Exp(ctx context.Context) {

}

type LossLessWithExp struct {
	BoolValue
	LastModified int64
}

func (b *LossLessWithExp) Exp(ctx context.Context) {

}

type PanoramaWithExp struct {
	BoolValue
	LastModified int64
}

func (b *PanoramaWithExp) Exp(ctx context.Context) {

}

type ShakeWithExp struct {
	BoolValue
	LastModified int64
}

func (b *ShakeWithExp) Exp(ctx context.Context) {

}

func abTestRun(ctx context.Context, flag *ab.StringFlag) bool {
	t, ok := ab.FromContext(ctx)
	if !ok {
		return false
	}
	expRes := flag.Value(t)
	return expRes == "1" || expRes == "11"
}

func batchGetDurationBetweenExpAndNow(expTimes ...string) map[string]time.Duration {
	var (
		res = make(map[string]time.Duration, len(expTimes))
	)
	for _, v := range expTimes {
		expStartTime, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			log.Error("time.ParseInLocation error(%+v), raw time(%s)", err, v)
			continue
		}
		res[v] = time.Since(expStartTime)
	}
	return res
}

func buildPeriods(in map[string]time.Duration) (string, map[string]string) {
	//来自账号新老用户接口：
	//每个区间定义为 start-end
	//实例： 0-24,36-60 则表示需要判断 用户（buvid）是否为0-24小时注册或者36-60小时注册
	var (
		periods            string
		periodsWithExpTime = make(map[string]string, len(in))
	)
	for expTime, duration := range in {
		periodsWithExpTime[expTime] = fmt.Sprintf("0-%d", int64(duration.Hours()))
		if periods == "" {
			periods = fmt.Sprintf("0-%d", int64(duration.Hours()))
			continue
		}
		periods = fmt.Sprintf("%s,0-%d", periods, int64(duration.Hours()))
	}
	return periods, periodsWithExpTime
}
