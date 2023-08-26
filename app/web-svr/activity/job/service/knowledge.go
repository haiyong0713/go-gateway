package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *Service) ActKnowledge() {
	ctx := context.Background()
	log.Infoc(ctx, "[ActKnowledge] Flush ActKnowledgeData Into Cache : time %s", time.Now().Format("2006-01-02 15:04"))

	s.actKnowledgeRunning.Lock()
	defer s.actKnowledgeRunning.Unlock()

	// 分赛道获取
	config := s.c.Knowledge
	sIDs := strings.Split(config.Sid, ",")
	// sid => lids的信息
	sidMapLids := make(map[int64][]*like.Like)
	for _, sid := range sIDs {
		var tmp []int64
		v, _ := strconv.ParseInt(sid, 10, 64)
		tmp = append(tmp, v)
		// likes表里面获取用户信息 id mid
		list, err := s.loadLikesList(ctx, tmp, 1)
		if len(list) == 0 {
			log.Errorc(ctx, "[ActKnowledge] Flush ActKnowledgeData Empty Data list:%v", list)
			return
		}
		if err != nil {
			log.Errorc(ctx, "[ActKnowledge] Flush ActKnowledgeData Into Cache From DB Err : %v", err)
			return
		}
		sidMapLids[v] = list
	}
	upVote := make(map[int64]like.LIDWithVotes)
	if len(sidMapLids) > 0 {
		sidUserList := make(map[int64][]int64)
		// 整合lid获取票数
		for sid, list := range sidMapLids {
			for _, v := range list {
				sidUserList[sid] = append(sidUserList[sid], v.ID)
			}
			res, err := s.GetTotalLikesByLID(ctx, sidUserList[sid])
			if err != nil {
				log.Errorc(ctx, "[ActKnowledge] GetTotalLikesByLID Err err:%v sid:%v", err, sid)
				return
			}
			// 将sidUserList每一个up主 拿到自己的票数
			for _, v := range list {
				each := &like.LIDWithVote{
					ID:  v.ID,
					Wid: v.Wid,
				}
				vote, ok := res[v.ID]
				if ok {
					each.Vote = vote.Like
				} else {
					each.Vote = 0
				}
				upVote[sid] = append(upVote[sid], each)
			}
			// 排序
			sort.Sort(upVote[sid])
			for k, v := range upVote[sid] {
				v.Order = int64(k + 1)
			}
		}
	}

	// 按照sid维度来刷缓存
	for sid, v := range upVote {
		str, err := json.Marshal(v)
		if err != nil {
			log.Errorc(ctx, "[ActKnowledge] FlushVoteCacheBySid Before json.Marshal Error err:%v", err)
			continue
		}
		log.Infoc(ctx, "[ActKnowledge] FlushVoteCacheBySid redis result %s", string(str))
		err = s.dao.FlushVoteCacheBySid(ctx, sid, string(str))
		if err != nil {
			continue
		}
	}
}

func (s *Service) GetTotalLikesByLID(ctx context.Context, lids []int64) (res map[int64]*like.Extend, err error) {
	res, err = s.dao.RawLikeExtendByLids(ctx, lids)
	if err != nil {
		log.Errorc(ctx, "Get RawLikeExtend From DB Err : %v", err)
		return
	}
	return
}
