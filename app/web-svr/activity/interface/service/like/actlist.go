package like

import (
	"context"
	"go-common/library/log"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
)

func (s *Service) ListDomain(ctx context.Context, pN, pS int) (list []*lmdl.Record, err error) {
	var (
		value      interface{}
		start, end int
	)

	if value, err = s.gCache.Get(_typeDomain); err != nil {
		log.Errorc(ctx, "ListDomain gCache Get err:%v", err)
		return
	}
	if list, ok := value.([]*lmdl.Record); ok {
		if start = (pN - 1) * pS; start < 0 {
			start = 0
		}

		if end = start + pS; end > len(list) {
			end = len(list)
		}
		return list[start:end], nil
	}
	return
}

func (s *Service) SearchDomain(ctx context.Context, domainName string) (record *lmdl.Record, err error) {
	var (
		value interface{}
	)
	if value, err = s.gCache.Get(_typeDomain); err != nil {
		log.Errorc(ctx, "SearchDomain gCache Get err:%v", err)
		return nil, err
	}
	if list, ok := value.([]*lmdl.Record); ok {
		for _, v := range list {
			if v.FirstDomain == domainName || (v.SecondDomain == domainName && v.SecondDomain != "") {
				return v, nil
			}
		}
	}
	return
}
