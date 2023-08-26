package lottery

import (
	"context"
	"encoding/json"
	"go-common/library/cache"
	"go-common/library/log"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
)

// getMemberGroupMap ...
func (s *Service) getMemberGroupMap(c context.Context, sid string) (memberGroup map[int64]*l.MemberGroup, err error) {
	res, err := s.getMemberGroupByIds(c, sid)
	if err != nil {
		return nil, err
	}
	return s.turnMemberGroupBatchToMap(c, res), nil
}

// getMemberGroupByIds 获取用户组
func (s *Service) getMemberGroupByIds(c context.Context, sid string) (memberGroup []*l.MemberGroup, err error) {
	memberGroup, err = s.lottery.CacheMemberGroup(c, sid)
	if err != nil {
		err = nil
	}
	if len(memberGroup) != 0 {
		cache.MetricHits.Inc("LotteryMemberGroup")
		return
	}
	cache.MetricMisses.Inc("LotteryMemberGroup")
	resDao, err := s.lottery.RawMemberGroup(c, sid)
	if err != nil {
		return
	}
	memberGroup = s.turnMemberGroupDBToMemberGroup(c, resDao)
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheMemberGroup(c, sid, memberGroup)
	})
	return
}

func (s *Service) turnMemberGroupDBToMemberGroup(c context.Context, memberGroup []*l.MemberGroupDB) []*l.MemberGroup {
	res := make([]*l.MemberGroup, 0, len(memberGroup))
	for _, v := range memberGroup {
		var group []*l.Group
		err := json.Unmarshal([]byte(v.Group), &group)
		if err != nil {
			log.Errorc(c, "turnMemberGroupDBToMemberGroup v.Group(%s)", v.Group)
			continue
		}
		res = append(res, &l.MemberGroup{
			ID:    v.ID,
			Name:  v.Name,
			Group: group,
		})
	}
	return res
}
func (s *Service) turnMemberGroupBatchToMap(c context.Context, memberGroup []*l.MemberGroup) map[int64]*l.MemberGroup {
	res := make(map[int64]*l.MemberGroup)
	for _, v := range memberGroup {
		res[v.ID] = v
	}
	return res
}
