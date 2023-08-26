package like

import (
	"context"
	"go-common/library/log"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
)

var (
	MemSubjectRule = make(map[int64][]*likemdl.SubjectRule)
)

func (d *Dao) MemorySubjectRulesBySid(c context.Context, sid int64) (res []*likemdl.SubjectRule, err error) {
	if res, ok := MemSubjectRule[sid]; ok {
		return res, nil
	}
	return d.SubjectRulesBySid(c, sid)
}

func (d *Dao) AddMemorySubjectRulesBySid(c context.Context, sids []int64) (err error) {
	tmp := make(map[int64][]*likemdl.SubjectRule)
	for _, sid := range sids {
		rule, err := d.SubjectRulesBySid(c, sid)
		if err == nil {
			tmp[sid] = rule
		} else {
			log.Errorc(c, "AddMemorySubjectRulesBySid sid[%d] err[%v]", sid, err)
			if rule, ok := MemSubjectRule[sid]; ok {
				tmp[sid] = rule
			}
		}
	}
	MemSubjectRule = tmp
	return nil
}
