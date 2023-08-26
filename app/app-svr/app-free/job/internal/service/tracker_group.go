package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/app-free/job/internal/model"
)

var (
	_ipCountm = map[string]Count{}
	lock      sync.Mutex
)

type Count struct {
	Count    int
	LastTime time.Time
}

func (s *Service) initTrackerRailGun(cfg *railgun.KafkaConfig, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewKafkaInputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.trackerUnpack, s.trackerDo)
	g := railgun.NewRailGun("客户端播放日志", nil, inputer, processor)
	s.trackerRailGun = g
	g.Start()
}

// nolint:gomnd
func (s *Service) trackerUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	record := strings.Split(string(msg.Payload()), "|")
	if len(record) < 40 {
		return nil, nil
	}
	ctime := record[22]
	ct, _ := strconv.ParseInt(ctime, 10, 64)
	return &railgun.SingleUnpackMsg{
		Group: ct,
		Item:  record,
	}, nil
}

func (s *Service) trackerDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	// http://berserker.bilibili.co/#/lancer/online/job/157077472678801?breadcrumb=%E7%A7%BB%E5%8A%A8%E7%AB%AFtracker%E6%95%B0%E6%8D%AE%E9%87%87%E6%A0%B7
	// https://info.bilibili.co/pages/viewpage.action?pageId=59393010
	// https://info.bilibili.co/pages/viewpage.action?pageId=30138298
	record := item.([]string)
	appID, buvid, versionCode, eventID, ctime, extendedFields := record[3], record[5], record[18], record[20], record[22], record[28]
	_ = versionCode
	ct, _ := strconv.ParseInt(ctime, 10, 64)
	hCtime := time.Unix(0, ct*1e6)
	switch eventID {
	case "main.ijk.http_build.tracker":
	default:
		return railgun.MsgPolicyIgnore
	}
	if appID != "1" {
		return railgun.MsgPolicyIgnore
	}
	if buvid == "" {
		return railgun.MsgPolicyIgnore
	}
	if extendedFields == "" {
		return railgun.MsgPolicyIgnore
	}
	var ef *model.ExtendedFields
	if err := json.Unmarshal([]byte(extendedFields), &ef); err != nil {
		log.Error("%+v", err)
		return railgun.MsgPolicyIgnore
	}
	switch ef.Mode {
	case "101", "102", "103", "201", "202", "203", "401", "402", "403", "501", "502", "503":
	default:
		return railgun.MsgPolicyIgnore
	}
	log.Warn("consume message ctime:%s record:%+v", hCtime, record)
	// nolint:unparam
	checkFunc := func(bus string, ip, url, httpCode string) {
		if url == "" || url == "null" {
			return
		}
		if ip == "" || ip == "null" {
			return
		}
		if !model.IsPublicIP(ip) {
			return
		}
		ipInt := model.InetAtoN(ip)
		if s.matchIP(ipInt) {
			return
		}
		s.JsAndWsAlarm(url, ip, httpCode)
	}
	checkFunc("video", ef.VideoIP, ef.VideoURL, ef.VideoHTTPCode)
	checkFunc("", ef.AudioIP, ef.AudioURL, ef.AudioHTTPCode)
	return railgun.MsgPolicyNormal
}

func (s *Service) matchIP(ipInt int64) bool {
	for _, rs := range s.freeRecords {
		for _, r := range rs {
			if ipInt >= r.IPStartInt && ipInt <= r.IPEndInt {
				return true
			}
		}
	}
	return false
}

// nolint:unused
func (s *Service) checkBCDN(uri, plat string, ver int) bool {
	u, err := url.Parse(uri)
	if err != nil {
		return false
	}
	// if match, _ := regexp.MatchString("^(([0-9]{1,3}\\.){3}[0-9]{1,3}|((proxy|upcdn|upos|bfs|acache)-tf-|cn-).*\\.(acgvideo|bilivideo)\\.com)/|/live-bvc/", u.Host+u.Path); match {
	// 	log.Warn("checkBCDN uri:%v,plat:%v,ver:%v,result:%v", uri, plat, ver, match)
	// 	return true
	// }
	// nolint:gosimple
	if match := strings.Index(u.Host+u.Path, "/if5ax/") > -1; match {
		log.Warn("checkBCDN uri:%v,plat:%v,ver:%v,result:%v", uri, plat, ver, match)
		return true
	}
	// ios 5.57 9300
	// android 5.58 5580100
	if (plat == "iphone" && ver < 9300) || (plat == "android" && ver < 5580100) {
		if match, _ := regexp.MatchString("^([0-9]{1,3}\\.){3}[0-9]{1,3}:480/", u.Host+u.Path); match {
			log.Warn("checkBCDN uri:%v,plat:%v,ver:%v,result:%v", uri, plat, ver, match)
			return true
		}
	}
	return false
}

func (s *Service) JsAndWsAlarm(uri, ip, httpCode string) {
	if httpCode != "200" && httpCode != "206" {
		return
	}
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	info, err := s.dao.Info(context.Background(), ip)
	if err != nil {
		return
	}
	if info.GetInfo().GetCountry() == "保留地址" {
		return
	}
	if info.GetInfo().GetCountry() != "中国" {
		return
	}
	if err := ping(u, ip); err != nil {
		log.Error("%+v", err)
		return
	}
	duration, err := s.ac.Get("duration").Duration()
	if err != nil {
		log.Error("%+v", err)
		duration = time.Hour
	}
	lock.Lock()
	count := _ipCountm[u.Host+ip]
	// nolint:gosimple
	if time.Now().Sub(count.LastTime) > duration {
		count.Count = 0
	}
	count.LastTime = time.Now()
	count.Count++
	_ipCountm[u.Host+ip] = count
	lock.Unlock()
	// 累计到达次数告警
	threshold, err := s.ac.Get("threshold").Int()
	if err != nil {
		log.Error("%+v", err)
		threshold = 5
	}
	if count.Count >= threshold {
		log.Error("日志告警: %v ,域名: %v DNS解析IP: %v (%v,%v,%v)没有备案,http_code:%v,累计%v次", u.String(), u.Host, ip, info.GetInfo().GetProvince(), info.GetInfo().GetCity(), info.GetInfo().GetIsp(), httpCode, count.Count)
	}
}

func ping(uri *url.URL, ip string) error {
	host := uri.Host
	url := &url.URL{}
	*url = *uri
	url.Host = ip
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}
	req.Host = host
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Warn("ping url:%s,ip:%s,response:%+v", uri.String(), ip, resp)
	return nil
}
