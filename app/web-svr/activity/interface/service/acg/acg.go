package acg

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/acg"
)

func (s *Service) Task(c context.Context, mid int64) (res interface{}, err error) {
	conn := s.redis.Get(c)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", fmt.Sprintf("acg2020_mid_%d", mid)))
	if err == redis.ErrNil {
		return map[string]interface{}{
			"task": s.c.Acg2020.Task,
			"user": nil,
		}, nil
	}
	if err != nil {
		log.Errorc(c, "Task conn.Do(GET, %s error(%v)", fmt.Sprintf("acg2020_mid_%d", mid), err)
		return nil, err
	}
	one := &acg.UserTaskState{}
	if err := json.Unmarshal(reply, &one); err != nil {
		log.Errorc(c, "Task json.Unmarshal(%s) error(%v)", reply, err)
		return nil, err
	}
	return map[string]interface{}{
		"task": s.c.Acg2020.Task,
		"user": one,
	}, nil
}
