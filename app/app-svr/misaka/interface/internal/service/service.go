package service

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/misaka/interface/internal/dao"
	appmodel "go-gateway/app/app-svr/misaka/interface/internal/model/app"
	webmodel "go-gateway/app/app-svr/misaka/interface/internal/model/web"

	ipdb "github.com/ipipdotnet/ipdb-go"
)

var (
	_emptyCityInfo = map[string]string{}
)

// Service service.
type Service struct {
	ac  *paladin.Map
	dao *dao.Dao

	// new ip library
	v4 *ipdb.City
	v6 *ipdb.City
}

// New new a service and return.
func New() (s *Service) {
	var ac = new(paladin.TOML)
	if err := paladin.Watch("application.toml", ac); err != nil {
		panic(err)
	}
	s = &Service{
		ac:  ac,
		dao: dao.New(ac),
	}
	s.loadIP()
	return s
}

// AppReport is
func (s *Service) AppReport(ctx context.Context, body []byte) (code int, err error) {
	code = http.StatusOK
	data := new(appmodel.Data)
	if err = data.Unmarshal(body); err != nil {
		code = http.StatusBadRequest
		log.Error("data.Unmarshal error(%v)", err)
		return
	}
	if data.Size() == 0 {
		code = http.StatusBadRequest
		log.Error("data report is empty")
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	cityInfo := s.parseIP(ip)
	info := &appmodel.Info{
		IP:        ip,
		Data:      data,
		Country:   cityInfo["country_name"],
		Province:  cityInfo["region_name"],
		City:      cityInfo["city_name"],
		ISP:       cityInfo["isp_domain"],
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}
	if err = s.dao.PubApp(ctx, info); err != nil {
		log.Info("s.dao.PubApp(%+v) error(%+v)", info, err)
		code = http.StatusInternalServerError
	}
	return
}

// WebReport is
func (s *Service) WebReport(ctx context.Context, body []byte) (code int, err error) {
	code = http.StatusOK
	data := new(webmodel.Data)
	if err = json.Unmarshal(body, data); err != nil {
		code = http.StatusBadRequest
		log.Error("data.Unmarshal error(%v)", err)
		return
	}
	if data.Size() == 0 {
		code = http.StatusBadRequest
		log.Error("data report is empty")
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	cityInfo := s.parseIP(ip)
	info := &webmodel.Info{
		IP:        ip,
		Data:      data,
		Country:   cityInfo["country_name"],
		Province:  cityInfo["region_name"],
		City:      cityInfo["city_name"],
		ISP:       cityInfo["isp_domain"],
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}
	if err = s.dao.PubWeb(ctx, info); err != nil {
		log.Info("s.dao.PubWeb(%+v) error(%+v)", info, err)
		code = http.StatusInternalServerError
	}
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}

func (s *Service) parseIP(addr string) (cityInfo map[string]string) {
	ipv := net.ParseIP(addr)
	var err error
	if ip := ipv.To4(); ip != nil {
		if cityInfo, err = s.v4.FindMap(addr, "CN"); err != nil {
			log.Error("addr:%s parse error:%v", addr, err)
			return
		}
	} else if ip := ipv.To16(); ip != nil {
		if cityInfo, err = s.v6.FindMap(addr, "CN"); err != nil {
			log.Error("addr:%s parse error:%v", addr, err)
			return
		}
	}
	if cityInfo == nil {
		cityInfo = _emptyCityInfo
		return
	}
	// ex.: from 中国 台湾 花莲市 to 台湾 花莲市 ”“
	if cityInfo["region_name"] == "香港" || cityInfo["region_name"] == "澳门" || cityInfo["region_name"] == "台湾" {
		cityInfo["country_name"] = cityInfo["region_name"]
		cityInfo["region_name"] = cityInfo["city_name"]
		cityInfo["city_name"] = ""
	}
	// ex.: from 中国 中国 ”“ to 中国 ”“ ”“
	if cityInfo["country_name"] == cityInfo["region_name"] {
		cityInfo["region_name"] = ""
		cityInfo["city_name"] = ""
	} else if cityInfo["region_name"] == cityInfo["city_name"] {
		// ex.: from 中国 北京 北京 to 中国 北京 ”“
		cityInfo["city_name"] = ""
	}
	return
}

func (s *Service) loadIP() {
	v4file, err := s.ac.Get("ipv4File").String()
	if err != nil {
		panic("load ip v4 file error")
	}
	v4, err := ipdb.NewCity(v4file)
	if err != nil {
		panic("parse ip v4 file error")
	}
	s.v4 = v4

	v6file, err := s.ac.Get("ipv6File").String()
	if err != nil {
		panic("load ip v6 file error")
	}
	v6, err := ipdb.NewCity(v6file)
	if err != nil {
		panic("parse ip v6 file error")
	}
	s.v6 = v6
}
