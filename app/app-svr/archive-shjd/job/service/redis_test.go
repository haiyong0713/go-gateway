package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive-shjd/job/model"
	"go-gateway/app/app-svr/archive/service/api"

	. "github.com/glycerine/goconvey/convey"
)

func Test_Init_Stat_Redis_From_Arc(t *testing.T) {
	Convey("test init stat redis from archive service", t, func() {
		testAid := int64(1009600)
		ctx := context.TODO()
		stat := &api.Stat{
			Aid: testAid,
			Fav: 99,
		}
		err := s.updateArcRedis(ctx, stat, testAid) // 更新arc stat
		So(err, ShouldBeNil)
		conn := s.statRedis.Get(ctx)
		key := s.statPBKey(testAid)
		fmt.Println(key)
		conn.Do("DEL", key)
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		msg := &model.StatMsg{
			Aid:   testAid,
			Share: 200,
			Type:  model.TypeForShare,
		}
		s.initRedis(ctx, msg)
		statRds, err := s.getStatFromRedis(ctx, testAid)
		fmt.Printf("stat job redis: %+v", statRds)
		arcStat, err := getStatFromArcService(s, testAid, 0)
		fmt.Printf("arc service redis: %+v", arcStat)
		So(statRds, ShouldResemble, arcStat)
	})
}

func getStatFromArcService(s *Service, aid int64, idx int) (stat *api.Stat, err error) {
	conn := s.arcRedises[idx].Get(context.Background())
	defer conn.Close()
	key := s.statPBKey(aid)
	stat = new(api.Stat)
	res, err := redis.Bytes(conn.Do("get", key))
	if err != nil {
		if err == redis.ErrNil {
			return
		}
		return
	}
	if err = stat.Unmarshal(res); err != nil {
		log.Error("json unmarshal error")
		return
	}
	return
}
