package unicom

import (
	"context"

	log "go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/model"
	"go-gateway/app/app-svr/app-wall/interface/model/operator"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"
)

// UserIPInfo ip info
func (s *Service) UserIPInfo(c context.Context, ipStr string) (res *operator.IPInfo) {
	res = &operator.IPInfo{
		IP:       ipStr,
		Operator: "unknown",
	}
	// unicom
	if model.IsIPv4(ipStr) {
		if s.unciomIPState(model.InetAtoN(ipStr)) {
			res.Operator = "unicom"
			return
		}
	}
	// unicom
	info, err := s.locdao.InfoGRPC(c, ipStr)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	switch info.Isp {
	case "联通":
		res.Operator = "unicom"
	case "移动":
		res.Operator = "mobile"
	case "电信":
		res.Operator = "telecom"
	default:
		log.Warn("UserIPInfo ip:%s,isp:%s,country:%s,province:%s,city:%s", ipStr, info.GetIsp(), info.GetCountry(), info.GetProvince(), info.GetCity())
	}
	return
}

func (s *Service) UserIPOrder(ctx context.Context, ipStr, operator string) (*unicom.UnicomUserIP, string) {
	res := &unicom.UnicomUserIP{IPStr: ipStr, IsValide: false}
	if operator == "unicom" && model.IsIPv4(ipStr) {
		if s.unciomIPState(model.InetAtoN(ipStr)) {
			res.IsValide = true
			return res, operator
		}
	}
	info, err := s.locdao.InfoGRPC(ctx, ipStr)
	if err != nil {
		log.Error("UserIPOrder InfoGRPC(%v) error:%+v", ipStr, err)
		return res, ""
	}
	if (info.Isp == "电信" && operator == "telecom") ||
		(info.Isp == "移动" && operator == "mobile") ||
		(info.Isp == "联通" && operator == "unicom") {
		res.IsValide = true
		return res, operator
	}
	if info.Isp == "" {
		log.Error("UserIPOrder isp is empty ip:%s operator:%s res:%+v", ipStr, operator, info)
		return res, ""
	}
	log.Warn("UserIPOrder not match ip:%s operator:%s res:%+v", ipStr, operator, info)
	return res, ""
}

// UserIPInfoV2 old ip info
func (s *Service) UserIPInfoV2(c context.Context, ipStr string) (res *unicom.UnicomUserIP) {
	res = &unicom.UnicomUserIP{
		IPStr:    ipStr,
		IsValide: false,
	}
	// unicom
	if model.IsIPv4(ipStr) {
		if s.unciomIPState(model.InetAtoN(ipStr)) {
			res.IsValide = true
			return
		}
	}
	// unicom
	info, err := s.locdao.InfoGRPC(c, ipStr)
	if err != nil {
		log.Error("UserIPInfoV2 s.locdao.Info error(%v)", err)
		return
	}
	switch info.Isp {
	case "联通", "移动", "电信":
		res.IsValide = true
	default:
		log.Error("user_ip_info_v2 ip(%s) res(%+v)", ipStr, info)
	}
	return
}
