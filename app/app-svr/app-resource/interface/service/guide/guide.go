package guide

import (
	"encoding/json"
	"hash/crc32"
	"io/ioutil"
	"os"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/guide"
)

var (
	_emptyinterest = []*guide.Interest{}
)

// Service interest service.
type Service struct {
	c            *conf.Config
	cache        []*guide.Interest
	interestPath string
	// infoc
	logCh chan interface{}
}

// New new a interest service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		cache:        []*guide.Interest{},
		interestPath: c.InterestJSONFile,
		// infoc
		logCh: make(chan interface{}, 1024),
	}
	s.loadInterestJSON()
	return
}

// Interest buvid or time gray
func (s *Service) Interest(mobiApp, buvid string, now time.Time) (res []*guide.Interest) {
	res = s.cache
	// if buvid != "" && mobiApp == "android" {
	// 	if crc32.ChecksumIEEE([]byte(reverseString(buvid)))%5 < 2 {
	// 		log.Info("interest_buvid_miss")
	// 		res = _emptyinterest
	// 		return
	// 	}
	// }
	if len(res) == 0 {
		log.Info("interest_null")
		res = _emptyinterest
		return
	}
	log.Info("interest_hit")
	return
}

// Interest2 is
func (s *Service) Interest2(mobiApp, buvid string) (res *guide.InterestTM) {
	res = &guide.InterestTM{}
	// switch group := int(crc32.ChecksumIEEE([]byte(buvid)) % 20); group {
	// case 9, 18:
	// 	res = &guide.InterestTM{
	// 		Interests: s.cache,
	// 	}
	// }
	//nolint:gomnd
	id := crc32.ChecksumIEEE([]byte(reverseString(buvid))) % 100
	//nolint:gomnd
	if id < 5 {
		res.FeedType = 1
	}
	switch mobiApp {
	case "iphone_b", "android_b":
		res.FeedType = 0
	}
	return
}

func reverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// loadInterestJSON load interest json
func (s *Service) loadInterestJSON() {
	file, err := os.Open(s.interestPath)
	if err != nil {
		log.Error("os.Open(%s) error(%v)", s.interestPath, err)
		return
	}
	defer file.Close()
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error("ioutil.ReadAll err %v", err)
		return
	}
	res := []*guide.Interest{}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal() file(%s) error(%v)", s.interestPath, err)
	}
	s.cache = res
	log.Info("loadInterestJSON success")
}
