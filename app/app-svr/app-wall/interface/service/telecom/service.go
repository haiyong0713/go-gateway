package telecom

import (
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	seqDao "go-gateway/app/app-svr/app-wall/interface/dao/seq"
	telecomDao "go-gateway/app/app-svr/app-wall/interface/dao/telecom"
)

const _telecomKey = "telecom"

type Service struct {
	c                  *conf.Config
	dao                *telecomDao.Dao
	seqdao             *seqDao.Dao
	flowPercentage     int
	smsTemplate        string
	smsMsgTemplate     string
	smsFlowTemplate    string
	smsOrderTemplateOK string
	telecomArea        map[string]struct{}
	// prom
	pHit  *prom.Prom
	pMiss *prom.Prom
	// cache
	cache *fanout.Fanout
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                  c,
		dao:                telecomDao.New(c),
		seqdao:             seqDao.New(c),
		flowPercentage:     c.Telecom.FlowPercentage,
		smsTemplate:        c.Telecom.SMSTemplate,
		smsMsgTemplate:     c.Telecom.SMSMsgTemplate,
		smsFlowTemplate:    c.Telecom.SMSFlowTemplate,
		smsOrderTemplateOK: c.Telecom.SMSOrderTemplateOK,
		telecomArea:        map[string]struct{}{},
		// prom
		pHit:  prom.CacheHit,
		pMiss: prom.CacheMiss,
		// cache
		cache: fanout.New("cache", fanout.Buffer(10240)),
	}
	s.loadTelecomArea(c)
	return
}

func (s *Service) loadTelecomArea(c *conf.Config) {
	areas := make(map[string]struct{}, len(c.Telecom.Area))
	for _, v := range c.Telecom.Area {
		for _, area := range v {
			if _, ok := areas[area]; !ok {
				areas[area] = struct{}{}
			}
		}
	}
	s.telecomArea = areas
	log.Info("loadTelecomArea success")
}
