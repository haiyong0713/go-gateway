package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/stat/job/model"

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
		// 设置DB的值为90
		stat.Fav = 90
		s.dao.Update(ctx, stat)
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
		s.initRedisAndDB(ctx, msg)
		statRds, err := s.getStatFromRedis(ctx, testAid)
		fmt.Printf("stat job redis: %+v", statRds)
		arcStat, err := getStatFromArcService(s, testAid, 0)
		fmt.Printf("arc service redis: %+v", arcStat)
		So(statRds, ShouldResemble, arcStat)
	})
}

func Test_Redis_Int64Map(t *testing.T) {
	Convey("Test_Not_Delete", t, func() {
		var testNotExist int64 = 99999999999999
		stat, err := s.getStatFromRedis(context.TODO(), testNotExist)
		So(err, ShouldResemble, fmt.Errorf("redis doesn't exsit key %s when execute HGETALL", s.statPBKey(testNotExist)))
		emptyStat := &api.Stat{}
		So(stat, ShouldResemble, emptyStat)
	})
}
