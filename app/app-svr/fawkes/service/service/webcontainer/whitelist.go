package webcontainer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/ecode"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/empty"
	"k8s.io/apimachinery/pkg/util/sets"

	"go-gateway/app/app-svr/fawkes/service/api/app/webcontainer"
	"go-gateway/app/app-svr/fawkes/service/conf"
	webmdl "go-gateway/app/app-svr/fawkes/service/model/webcontainer"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// AddWhiteList 如果url已经存在，则会覆盖原来的配置
func (s *Service) AddWhiteList(ctx context.Context, req *webcontainer.AddWhiteListReq) (resp *empty.Empty, err error) {
	resp = &empty.Empty{}
	domains := strings.Split(req.Domain, ",")
	wm, err := s.fkDao.SelectWhitelistMapByDomain(ctx, domains)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("AddWhiteList domain:%s err:%v", req.Domain, err))
		return
	}
	var (
		update []*webmdl.WebWhiteList
		add    []string
	)
	for _, v := range domains {
		if _, ok := wm[v]; ok {
			update = append(update, wm[v])
		} else {
			add = append(add, v)
		}
	}
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("check domain status error:%v", err))
		log.Errorc(ctx, fmt.Sprintf("%v", err))
		return
	}
	if err = s.fkDao.BatchUpdateWhiteList(ctx, convert2Update(req, update)); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("AddWhiteList error:%v", err))
		log.Errorc(ctx, fmt.Sprintf("%v", err))
		return
	}
	if err = s.fkDao.BatchAddWhiteList(ctx, convert2Add(req, add)); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("AddWhiteList error:%v", err))
		log.Errorc(ctx, fmt.Sprintf("%v", err))
		return
	}
	return
}

func (s *Service) DelWhiteList(ctx context.Context, req *webcontainer.DelWhiteListReq) (resp *empty.Empty, err error) {
	resp = &empty.Empty{}
	_, err = s.fkDao.DelWhiteList(ctx, req.Id)
	if err != nil {
		log.Errorc(ctx, "DelWhiteList error: %v", err)
		return
	}
	return
}

func (s *Service) UpdateWhiteList(ctx context.Context, req *webcontainer.UpdateWhiteListReq) (resp *empty.Empty, err error) {
	resp = &empty.Empty{}
	_, err = s.fkDao.UpdateWhiteList(ctx, req)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("UpdateWhiteList error:%v", err))
		log.Errorc(ctx, "UpdateWhiteList error: %v", err)
		return
	}
	return
}

func (s *Service) GetWhiteList(ctx context.Context, req *webcontainer.GetWhiteListReq) (resp *webcontainer.GetWhiteListResp, err error) {
	var rows []*webmdl.WebWhiteList
	rows, err = s.fkDao.SelectWhiteList(ctx, req)
	if err != nil {
		log.Errorc(ctx, "UpdateWhiteList error: %v", err)
		return
	}
	resp = convert(rows)
	return
}

func (s *Service) WhiteListConfig(ctx context.Context, req *webcontainer.WhiteListConfigReq) (resp *webcontainer.WhiteListConfigResp, err error) {
	rows, err := s.fkDao.SelectWhiteListByActiveStatus(ctx, req.AppKey, true)
	var (
		activeDomains      []string
		allAbilityDomain   []string
		jsBridgeWhitelist  []string
		highlightWhitelist []string
		qrcodeWhitelist    []string
	)
	for _, v := range rows {
		activeDomains = append(activeDomains, v.Domain)
		f := sets.NewString(strings.Split(v.Feature, ",")...)
		if f.Has(string(webcontainer.Feature_JsBridge)) && f.Has(string(webcontainer.Feature_QrCode)) && f.Has(string(webcontainer.Feature_HighLight)) {
			// all
			allAbilityDomain = append(allAbilityDomain, v.Domain)
		} else {
			if f.Has(strconv.Itoa(int(webcontainer.Feature_JsBridge))) {
				jsBridgeWhitelist = append(jsBridgeWhitelist, v.Domain)
			}
			if f.Has(strconv.Itoa(int(webcontainer.Feature_HighLight))) {
				highlightWhitelist = append(highlightWhitelist, v.Domain)
			}
			if f.Has(strconv.Itoa(int(webcontainer.Feature_QrCode))) {
				qrcodeWhitelist = append(qrcodeWhitelist, v.Domain)
			}
		}
	}

	resp = &webcontainer.WhiteListConfigResp{
		H5AllAbilityWhitelist: regexGen(allAbilityDomain),
		H5JsbridgeWhitelist:   regexGen(jsBridgeWhitelist),
		H5HighlightWhitelist:  regexGen(highlightWhitelist),
		H5QrcodeWhitelist:     regexGen(qrcodeWhitelist),
		H5AlertWhitelist:      regexGen(activeDomains),
	}
	return
}

// DomainStatusSync sync active_status
func (s *Service) DomainStatusSync(ctx context.Context, _ *empty.Empty) (resp *empty.Empty, err error) {
	allDomains, err := s.fkDao.SelectAllWhiteList(ctx)
	var domain []string
	domainMap := make(map[string]*webmdl.WebWhiteList)
	for _, v := range allDomains {
		domain = append(domain, v.Domain)
		domainMap[v.Domain] = v
	}
	statusMap, err := s.fkDao.DomainStatus(ctx, domain)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("fission.doaminstatus error:%v", err))
		log.Errorc(ctx, fmt.Sprintf("arg:%v, err:%v", domain, err))
		return
	}
	if err = s.fkDao.BatchUpdateWhiteList(ctx, convert2StatusUpdate(domain, domainMap, statusMap.Data)); err != nil {
		return
	}
	return
}

func regexGen(domain []string) string {
	return strings.Join(domain, "|")
}

func convert2StatusUpdate(domain []string, domainMap map[string]*webmdl.WebWhiteList, statusMap map[string]int64) []*webcontainer.UpdateWhiteListReq {
	var r []*webcontainer.UpdateWhiteListReq
	for _, v := range domain {
		item := &webcontainer.UpdateWhiteListReq{
			Id:             domainMap[v].Id,
			IsDomainActive: &types.BoolValue{Value: statusMap[v] == 0},
		}
		r = append(r, item)
	}
	return r
}

func sting2Feature(input string, splitter string) []webcontainer.Feature {
	strs := strings.Split(input, splitter)
	ary := make([]webcontainer.Feature, len(strs))
	for i := range ary {
		n, _ := strconv.ParseInt(strs[i], 10, 32)
		ary[i] = webcontainer.Feature(n)
	}
	return ary
}

func convert(rows []*webmdl.WebWhiteList) *webcontainer.GetWhiteListResp {
	var list []*webcontainer.WhiteListInfo
	for _, v := range rows {
		i := &webcontainer.WhiteListInfo{
			Id:             v.Id,
			AppKey:         v.AppKey,
			Title:          v.Title,
			Domain:         v.Domain,
			Reason:         v.Reason,
			IsThirdParty:   v.IsThirdParty,
			CometId:        v.CometId,
			Feature:        sting2Feature(v.Feature, ","),
			Effective:      v.Effective.Unix(),
			Expires:        v.Expires.Unix(),
			Ctime:          v.Ctime.Unix(),
			Mtime:          v.Mtime.Unix(),
			IsDomainActive: v.IsDomainActive,
			CometUrl:       conf.Conf.Comet.ProcessUrl + v.CometId,
		}
		list = append(list, i)
	}
	return &webcontainer.GetWhiteListResp{
		Whitelist: list,
	}
}

func convert2Add(req *webcontainer.AddWhiteListReq, add []string) []*webcontainer.AddWhiteListReq {
	var list []*webcontainer.AddWhiteListReq
	for _, v := range add {
		i := &webcontainer.AddWhiteListReq{
			AppKey:       req.AppKey,
			Title:        req.Title,
			Domain:       v,
			Reason:       req.Reason,
			IsThirdParty: req.IsThirdParty,
			CometId:      req.CometId,
			Feature:      req.Feature,
			Effective:    req.Effective,
			Expires:      req.Expires,
		}
		list = append(list, i)
	}
	return list
}

func convert2Update(req *webcontainer.AddWhiteListReq, update []*webmdl.WebWhiteList) []*webcontainer.UpdateWhiteListReq {
	var list []*webcontainer.UpdateWhiteListReq
	for _, v := range update {
		i := &webcontainer.UpdateWhiteListReq{
			Id:           v.Id,
			Title:        req.Title,
			Reason:       req.Reason,
			IsThirdParty: req.IsThirdParty,
			CometId:      req.CometId,
			Feature:      req.Feature,
			Effective:    &types.Int64Value{Value: v.Effective.Unix()},
			Expires:      &types.Int64Value{Value: v.Expires.Unix()},
		}
		list = append(list, i)
	}
	return list
}
