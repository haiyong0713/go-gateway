package anticrawler

import (
	"os"
	"strconv"
	"strings"

	"go-common/library/log"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	_featureKey = "common.mogul"
	_sampleKey  = "common.mogul.sample"
)

var _featureSvc *feature.Feature

func init() {
	_featureSvc = feature.New(nil)
}

func wList(mid int64, buvid string) bool {
	if mid == 0 && buvid == "" {
		return false
	}
	list := strings.Split(_featureSvc.BusinessConfig(_featureKey), ",")
	if len(list) == 1 && list[0] == "" {
		return false
	}
	midStr := strconv.FormatInt(mid, 10)
	for _, val := range list {
		if val == "" {
			continue
		}
		switch val {
		case midStr, buvid:
			return true
		default:
		}
	}
	return false
}

func isSample(path string, sample int64) bool {
	if sample < 0 { // sample:-1||[0-99]
		return true
	}
	list := strings.Split(_featureSvc.BusinessConfig(_sampleKey), ",")
	if len(list) == 0 {
		return false
	}
	listm := map[string]int64{}
	for _, s := range list {
		ss := strings.Split(s, ":")
		if len(ss) < 2 {
			continue
		}
		rate, err := strconv.ParseInt(ss[1], 10, 64)
		if err != nil {
			log.Error("%+v", err)
			continue
		}
		listm[ss[0]] = rate
	}
	key := os.Getenv("APP_ID") + path
	samplingRate, ok := listm[key]
	if !ok {
		samplingRate = listm["all"]
	}
	if sample < samplingRate { // sample:[0-99],sampling_rate:1-100
		return true
	}
	return false
}
