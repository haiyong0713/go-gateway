package unicom

import (
	"context"
	"go-common/library/conf/env"

	log "go-common/library/log"
)

// loadUnicomIP load unicom ip
func (s *Service) loadUnicomIP() {
	unicomIP, err := s.dao.IPSync(context.TODO())
	if err != nil {
		log.Error("s.dao.IPSync error(%v)", err)
		return
	}
	s.unicomIpCache = unicomIP
	log.Info("loadUnicomIPCache success")
}

func (s *Service) loadUnicomPacks() {
	// 0下线，1上线，2预发
	states := []int{1}
	if env.DeployEnv == env.DeployEnvPre {
		states = []int{1, 2}
	}
	pack, err := s.dao.UserPacks(context.TODO(), states)
	if err != nil {
		log.Error("s.dao.UserPacks error(%v)", err)
		return
	}
	s.unicomPackCache = pack
}
