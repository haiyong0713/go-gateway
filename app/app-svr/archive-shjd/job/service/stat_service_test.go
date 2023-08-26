package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive-shjd/job/model"
	"go-gateway/app/app-svr/archive/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testAid  int64 = 100010
	testStat       = &api.Stat{
		Aid:     testAid,
		View:    80,
		Danmaku: 200,
		Reply:   0,
		Fav:     50,
		Coin:    0,
		Share:   0,
		NowRank: 0,
		HisRank: 0,
		Like:    0,
		DisLike: 0,
	}
)

func Test_Stat_Redis_Insert_And_Retrieve(t *testing.T) {
	Convey("testConnRedisAndInsert", t, func() {
		stat := testStat
		err := s.saveStatToRedis(context.Background(), stat)
		So(err, ShouldBeNil)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat, ShouldResemble, stat)
	})
}

func Test_Handle_Msg_For_New_Archive_Fav(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randFav := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Fav:  int32(randFav),
			Type: model.TypeForFav,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.Fav, ShouldEqual, randFav)
		//dbStat, err := s.dao.Stat(context.Background(), testAid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_Rank(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randRank := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:     testAid,
			HisRank: int32(randRank),
			Type:    model.TypeForRank,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.HisRank, ShouldEqual, randRank)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_Share(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randValue := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:   testAid,
			Share: int32(randValue),
			Type:  model.TypeForShare,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.Share, ShouldEqual, randValue)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_DM(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randVal := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			DM:   int32(randVal),
			Type: model.TypeForDm,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.Danmaku, ShouldEqual, randVal)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_Like(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randVal := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Like: int32(randVal),
			Type: model.TypeForLike,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.Like, ShouldEqual, randVal)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_Coin(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randVal := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Coin: int32(randVal),
			Type: model.TypeForCoin,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.Coin, ShouldEqual, randVal)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_Click(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randVal := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:   testAid,
			Click: int32(randVal),
			Type:  model.TypeForView,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.View, ShouldEqual, randVal)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}

func Test_Handle_Msg_For_New_Archive_Reply(t *testing.T) {
	Convey("Test_Flush_Redis_To_DB", t, func() {
		randVal := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:   testAid,
			Reply: int32(randVal),
			Type:  model.TypeForReply,
		}
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		s.statChan <- msg
		time.Sleep(time.Second * 2)
		queryStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(queryStat.Reply, ShouldEqual, randVal)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, queryStat)
	})
}
func Test_Handle_Msg_For_Existed_Archive_DM(t *testing.T) {
	testAid := int64(80086)
	Convey("Test_Handle_Msg_For_Existed_Archive", t, func() {
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		// ============== First Msg ================
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		randFav := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Fav:  int32(randFav),
			Type: model.TypeForFav,
		}
		s.statChan <- msg
		time.Sleep(time.Second * 1)
		firstRedisStat, err := s.getStatFromRedis(context.Background(), testAid)
		So(err, ShouldBeNil)
		So(firstRedisStat.Fav, ShouldEqual, randFav)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, firstRedisStat)
		// ==============Second Msg ==================
		randDm := rand.Int31n(9999)
		msg = &model.StatMsg{
			Aid:  testAid,
			DM:   int32(randDm),
			Type: model.TypeForDm,
		}
		s.statChan <- msg
		time.Sleep(time.Second)
		secondRedisStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(secondRedisStat.Danmaku, ShouldEqual, randDm)
		//dbStat, err = s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, firstRedisStat) // db数据依然和之前相同
		// =============== third msg ====================
		conn.Do("DEL", fmt.Sprintf("l:%d", testAid)) // 直接删去lock因此不用等待锁过期
		randLike := rand.Int31n(7777)
		msg = &model.StatMsg{
			Aid:  testAid,
			Like: int32(randLike),
			Type: model.TypeForLike,
		}
		s.statChan <- msg
		time.Sleep(time.Second)
		thirdRedisStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(thirdRedisStat.Like, ShouldEqual, randLike)
		//dbStat, err = s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, thirdRedisStat)
	})
}

func Test_Handle_Msg_For_Existed_Archive_Rank(t *testing.T) {
	testAid := int64(80086)
	Convey("Test_Handle_Msg_For_Existed_Archive", t, func() {
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		// ============== First Msg ================
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		randFav := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Fav:  int32(randFav),
			Type: model.TypeForFav,
		}
		s.statChan <- msg
		time.Sleep(time.Second * 1)
		firstRedisStat, err := s.getStatFromRedis(context.Background(), testAid)
		So(err, ShouldBeNil)
		So(firstRedisStat.Fav, ShouldEqual, randFav)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, firstRedisStat)
		// ==============Second Msg ==================
		randRank := rand.Int31n(9999)
		msg = &model.StatMsg{
			Aid:     testAid,
			NowRank: int32(randRank),
			HisRank: int32(randRank),
			Type:    model.TypeForRank,
		}
		s.statChan <- msg
		time.Sleep(time.Second)
		secondRedisStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(secondRedisStat.HisRank, ShouldEqual, randRank)
		//dbStat, err = s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, firstRedisStat) // db数据依然和之前相同
		// =============== third msg ====================
		conn.Do("DEL", fmt.Sprintf("l:%d", testAid)) // 直接删去lock因此不用等待锁过期
		randLike := rand.Int31n(7777)
		msg = &model.StatMsg{
			Aid:  testAid,
			Like: int32(randLike),
			Type: model.TypeForLike,
		}
		s.statChan <- msg
		time.Sleep(time.Second)
		thirdRedisStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(thirdRedisStat.Like, ShouldEqual, randLike)
		//dbStat, err = s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, thirdRedisStat)
	})
}

func Test_Handle_Msg_For_Existed_Archive_With_Fav_Coin_Rank(t *testing.T) {
	testAid := int64(80086)
	Convey("Test_Handle_Msg_For_Existed_Archive", t, func() {
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		// ============== First Msg ================
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		randFav := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Fav:  int32(randFav),
			Type: model.TypeForFav,
		}
		s.statChan <- msg
		time.Sleep(time.Second * 1)
		firstRedisStat, err := s.getStatFromRedis(context.Background(), testAid)
		So(err, ShouldBeNil)
		So(firstRedisStat.Fav, ShouldEqual, randFav)
		//dbStat, err := s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, firstRedisStat)
		// ==============Second Msg ==================
		randVal := rand.Int31n(9999)
		msg = &model.StatMsg{
			Aid:  testAid,
			Coin: int32(randVal),
			Type: model.TypeForCoin,
		}
		s.statChan <- msg
		time.Sleep(time.Second)
		secondRedisStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(secondRedisStat.Coin, ShouldEqual, randVal)
		//dbStat, err = s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, firstRedisStat) // db数据依然和之前相同
		// =============== third msg ====================
		conn.Do("DEL", fmt.Sprintf("l:%d", testAid)) // 直接删去lock因此不用等待锁过期
		randValMsg3 := rand.Int31n(7777)
		msg = &model.StatMsg{
			Aid:     testAid,
			HisRank: int32(randValMsg3),
			Type:    model.TypeForRank,
		}
		s.statChan <- msg
		time.Sleep(time.Second)
		thirdRedisStat, err := s.getStatFromRedis(context.TODO(), testAid)
		So(err, ShouldBeNil)
		So(thirdRedisStat.HisRank, ShouldEqual, randValMsg3)
		//dbStat, err = s.dao.Stat(context.Background(), testaid)
		//So(err, ShouldBeNil)
		//So(dbStat, ShouldResemble, thirdRedisStat)
	})
}

// 按照如下次序发送8次消息，保证每次消息都更新了stat redis和arc redis
func Test_Message_Fav_Coin_Rank_Share_Rely_View_DM_Like(t *testing.T) {
	Convey("Test_Message_Fav_Coin_Rank_Share_Rely_View_DM_Like", t, func() {
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		if _, err := conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid))); err != nil {
			t.Errorf("conn del error, err %+v", err)
			t.Fail()
		}
		if _, err := conn.Do("DEL", "l:"+strconv.Itoa(int(testAid))); err != nil {
			t.Errorf("conn del error, err %+v", err)
			t.Fail()
		}
		msgs := [8]model.StatMsg{}
		vals := [8]int{}
		for i := 0; i < 8; i++ {
			randVal := int(rand.Int31n(8888))
			vals[i] = randVal
			msgs[i] = model.StatMsg{
				Aid: testAid,
			}
			switch i {
			case 0:
				msgs[i].Fav = int32(randVal)
				msgs[i].Type = model.TypeForFav
			case 1:
				msgs[i].Coin = int32(randVal)
				msgs[i].Type = model.TypeForCoin
			case 2:
				msgs[i].HisRank = int32(randVal)
				msgs[i].Type = model.TypeForRank
			case 3:
				msgs[i].Share = int32(randVal)
				msgs[i].Type = model.TypeForShare
			case 4:
				msgs[i].Reply = int32(randVal)
				msgs[i].Type = model.TypeForReply
			case 5:
				msgs[i].Click = int32(randVal)
				msgs[i].Type = model.TypeForView
			case 6:
				msgs[i].DM = int32(randVal)
				msgs[i].Type = model.TypeForDm
			case 7:
				msgs[i].Like = int32(randVal)
				msgs[i].Type = model.TypeForLike
			default:
				t.Fail()
			}
		}
		for i := 0; i < 8; i++ {
			s.statChan <- &msgs[i]
			time.Sleep(time.Second * 3)
			statRedis, err := s.getStatFromRedis(context.TODO(), testAid)
			So(err, ShouldBeNil)
			arcRedis, err := getStatFromArcService(s, testAid, 0)
			So(err, ShouldBeNil)
			So(arcRedis, ShouldResemble, statRedis)
			switch i {
			// Fav_Coin_Rank_Share_Rely_View_DM_Like
			case 0:
				So(statRedis.Fav, ShouldEqual, vals[i])
			case 1:
				So(statRedis.Coin, ShouldEqual, vals[i])
			case 2:
				So(statRedis.HisRank, ShouldEqual, vals[i])
			case 3:
				So(statRedis.Share, ShouldEqual, vals[i])
			case 4:
				So(statRedis.Reply, ShouldEqual, vals[i])
			case 5:
				So(statRedis.View, ShouldEqual, vals[i])
			case 6:
				So(statRedis.Danmaku, ShouldEqual, vals[i])
			case 7:
				So(statRedis.Like, ShouldEqual, vals[i])
			default:
				t.Fail()
			}
		}
	})
}

func Test_Archive_Service_Is_Updated_No_Matter_What(t *testing.T) {
	testAid := int64(39939)
	Convey("test archive service redis is updated no matter what", t, func() {
		conn := s.statRedis.Get(context.Background())
		defer conn.Close()
		// ============== First Msg ================
		conn.Do("DEL", s.statPBKey(testAid)+strconv.Itoa(int(testAid)))
		conn.Do("DEL", "l:"+strconv.Itoa(int(testAid)))
		randFav := rand.Int31n(8888)
		msg := &model.StatMsg{
			Aid:  testAid,
			Fav:  int32(randFav),
			Type: model.TypeForFav,
		}
		s.statChan <- msg
		time.Sleep(time.Second * 1)
		for idx := range s.arcRedises {
			arcStat, err := getStatFromArcService(s, testAid, idx)
			log.Info("arcStat %+v of group %d on first msg", arcStat, idx)
			So(err, ShouldBeNil)
			So(arcStat.Fav, ShouldEqual, randFav)
		}
		// ==============Second Msg =================
		randRank := rand.Int31n(9999)
		msg = &model.StatMsg{
			Aid:     testAid,
			NowRank: int32(randRank),
			HisRank: int32(randRank),
			Type:    model.TypeForRank,
		}
		s.statChan <- msg
		time.Sleep(time.Second * 1)
		for idx := range s.arcRedises {
			arcStat, err := getStatFromArcService(s, testAid, idx)
			log.Info("arcStat %+v of group %d", arcStat, idx)
			So(err, ShouldBeNil)
			So(arcStat.HisRank, ShouldEqual, randRank)
		}
	})
}
