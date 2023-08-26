package operator

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"go-common/library/log"

	client "git.bilibili.co/bapis/bapis-go/bilibili/app/wall/v1"
)

func (s *Service) RuleInfo(ctx context.Context) (*client.RulesReply, error) {
	rulesInfo := map[string]*client.RulesInfo{}
	for isp, rs := range s.c.Rule {
		var rules []*client.RuleInfo
		for _, r := range rs {
			rule := &client.RuleInfo{Tf: r.Tf, M: r.M, A: r.A, P: r.P, ABackup: r.ABackup}
			rules = append(rules, rule)
		}
		rulesInfo[isp] = &client.RulesInfo{RulesInfo: rules}
	}
	bs, err := json.Marshal(rulesInfo)
	if err != nil {
		log.Error("%+v", err)
	}
	mh := md5.Sum(bs)
	hashValue := hex.EncodeToString(mh[:])
	return &client.RulesReply{
		RulesInfo: rulesInfo,
		HashValue: hashValue,
	}, nil
}
