package rank

import (
	"context"
	"go-common/library/log"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank_v3"
)

// RankRule 用户已经获得的卡
func (d *Dao) RankRule(c context.Context, ruleID int64) (res *rankmdl.Rule, err error) {
	rule, err := d.GetRankRule(c, ruleID)
	if err != nil {
		log.Errorc(c, "d.GetRankRule err(%v)", err)
	}
	if rule != nil && err == nil {
		return rule, nil
	}
	rankRule, err := d.GetRuleByID(c, ruleID)
	if err != nil {
		log.Errorc(c, "d.GetRuleByID(c, %d) err(%v)", ruleID, err)
		return nil, err
	}

	err = d.AddRankRule(c, ruleID, rankRule)
	if err != nil {
		log.Errorc(c, " d.AddRankRule err(%v)", err)
	}
	return rankRule, nil
}

// RankBase 用户已经获得的卡
func (d *Dao) RankBase(c context.Context, baseID int64) (res *rankmdl.Base, err error) {
	rule, err := d.GetRankBase(c, baseID)
	if err != nil {
		log.Errorc(c, "d.GetRankByID err(%v)", err)
	}
	if rule != nil && err == nil {
		return rule, nil
	}
	rankBase, err := d.GetRankByID(c, baseID)
	if err != nil {
		log.Errorc(c, "d.GetRankByID(c, %d) err(%v)", baseID, err)
		return nil, err
	}

	err = d.AddRankBase(c, baseID, rankBase)
	if err != nil {
		log.Errorc(c, " d.AddRankBase err(%v)", err)
	}
	return rankBase, nil
}
