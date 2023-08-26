package ab_play

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/exp/ab"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"
)

var (
	blackListScene        = "player.player.vertical-switch.0.player"
	_landscapeNewUserFlag = ab.String("ugc_horizon_goto_story", "横屏视频切全屏新用户实验", "0")
	_landscapeOldUserFlag = ab.String("ugc_horizon_goto_story_alluser", "横屏视频切全屏老用户实验", "0")
)

type Dao struct {
	c *conf.Config
	// redis
	Redis *redis.Redis
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		Redis: redis.NewRedis(c.Redis.PlayStoryRedis),
	}
	return
}

func buildBlackListKey(buvid string) string {
	return fmt.Sprintf("%s-%s", buvid, blackListScene)
}

func (d *Dao) LandscapeStoryExp(ctx context.Context, buvid string, newUser bool) (bool, string) {
	flagStr := _landscapeOldUserFlag
	if newUser {
		flagStr = _landscapeNewUserFlag

	}
	flag := d.abtestRun(ctx, buvid, flagStr)
	if flag == "3" || flag == "4" {
		return true, d.c.Custom.StoryIcon
	}
	return false, ""
}

func (d *Dao) abtestRun(ctx context.Context, buvid string, flag *ab.StringFlag) string {
	t, ok := ab.FromContext(ctx)
	if !ok {
		return "0"
	}
	t.Add(ab.KVString("buvid", buvid))
	return flag.Value(t)
}

func (d *Dao) HitPlayBlackList(c context.Context, buvid string) bool {
	reply, err := redis.Bytes(d.Redis.Do(c, "GET", buildBlackListKey(buvid)))
	if err != nil {
		if err != redis.ErrNil {
			log.Error("HitPlayBlackList redis error(%+v), buvid(%s)", err, buvid)
		}
		return false
	}
	var value struct {
		Date      string `json:"log_date"`
		StoryDays int64  `json:"story_days"`
	}
	if err := json.Unmarshal(reply, &value); err != nil {
		log.Error("HitPlayBlackList unmarshal error(%+v), buvid(%s), reply(%+v)", err, buvid, string(reply))
		return false
	}
	daysAll := d.c.Custom.StoryDays
	switch value.Date {
	case "20220602", "20220603", "20220604", "20220605", "20220606", "20220607", "20220608":

		for _, v := range daysAll {
			if v == value.StoryDays {
				return false //清洗出黑名单
			}
		}
		return true //命中黑名单
	default:
	}
	return false
}
