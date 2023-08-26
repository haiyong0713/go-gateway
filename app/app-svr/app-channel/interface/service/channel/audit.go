package channel

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-channel/interface/model"
)

var (
	_auditRids = map[int8]map[int]struct{}{
		model.PlatIPad: {
			65537: {},
			13:    {},
			167:   {},
			65555: {},
			3:     {},
			129:   {},
			4:     {},
			36:    {},
			188:   {},
			177:   {},
			23:    {},
			11:    {},
			65551: {},
			65561: {},
		},
		model.PlatIPadHD: {
			65537: {},
			13:    {},
			167:   {},
			65555: {},
			3:     {},
			129:   {},
			4:     {},
			36:    {},
			188:   {},
			177:   {},
			23:    {},
			11:    {},
			65551: {},
			65561: {},
		},
		model.PlatIPhone: {
			13:    {},
			167:   {},
			65545: {},
			177:   {},
			65555: {},
			65541: {},
			65537: {},
			65560: {},
			3:     {},
			129:   {},
			4:     {},
			36:    {},
			188:   {},
			23:    {},
			11:    {},
			65546: {},
			65561: {},
			65553: {},
			65551: {},
			65539: {},
			65550: {},
			65543: {},
		},
		model.PlatIPhoneI: {
			// 一级
			13:    {},
			65541: {},
			3:     {},
			129:   {},
			4:     {},
			36:    {},
			188:   {},
			65561: {},
			65552: {},
			65556: {},
			65550: {},
		},
	}
)

// auditRegion region data list.
func (s *Service) auditRegion(mobiApp string, plat int8, build, rid int) (isAudit bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			if rids, ok := _auditRids[plat]; ok {
				if _, ok = rids[rid]; !ok {
					return true
				}
			}
		}
	}
	return false
}

func (s *Service) auditList(mobiApp string, _ int8, build int) (isAudit bool) {
	if plats, ok := s.auditCache[mobiApp]; ok {
		if _, ok = plats[build]; ok {
			return true
		}
	}
	return false
}

func (s *Service) loadAuditCache() {
	log.Info("cronLog start loadAuditCache")
	as, err := s.adt.Audits(context.TODO())
	if err != nil {
		log.Error("s.adt.Audits error(%v)", err)
		return
	}
	s.auditCache = as
}
