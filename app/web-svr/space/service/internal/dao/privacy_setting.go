package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	pb "go-gateway/app/web-svr/space/service/api"
	"go-gateway/app/web-svr/space/service/internal/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	_privacySQL         = `SELECT id,privacy,status,modify_time FROM member_privacy%d WHERE mid=?`
	_privacyBatchAddSQL = `INSERT INTO member_privacy%d (mid,privacy,status) VALUES %s ON DUPLICATE KEY UPDATE status=VALUES(status)`
)

func privacyHit(mid int64) int64 {
	return mid % 10
}

func (d *dao) RawPrivacySetting(ctx context.Context, req *pb.PrivacySettingReq) ([]*model.MemberPrivacy, error) {
	var (
		res     []*model.MemberPrivacy
		newUser bool
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		rows, err := d.db.Query(ctx, fmt.Sprintf(_privacySQL, privacyHit(req.Mid)), req.Mid)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			r := &model.MemberPrivacy{}
			if err := rows.Scan(&r.ID, &r.Privacy, &r.Status, &r.ModifyTime); err != nil {
				return err
			}
			res = append(res, r)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(res) != 0 {
			return nil
		}
		res = []*model.MemberPrivacy{{ID: 0}}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		// 查询是否是新用户
		newUser = d.isNewUser(ctx, req.Mid)
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	for _, privacy := range res {
		privacy.NewUser = newUser
	}
	return res, nil
}

func (d *dao) cacheSFPrivacySetting(req *pb.PrivacySettingReq) string {
	return fmt.Sprintf("prvyset_%d", req.Mid)
}

func (d *dao) cacheExpire() int32 {
	rand.Seed(time.Now().UnixNano())
	return d.spaceExpire + rand.Int31n(d.cacheRand)
}

func (d *dao) CachePrivacySetting(ctx context.Context, req *pb.PrivacySettingReq) ([]*model.MemberPrivacy, error) {
	key := d.cacheSFPrivacySetting(req)
	ok, err := redis.Bool(d.redis.Do(ctx, "EXPIRE", key, d.cacheExpire()))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, err
	}
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	var res []*model.MemberPrivacy
	err = json.Unmarshal(data, &res)
	return res, nil
}

func (d *dao) AddCachePrivacySetting(ctx context.Context, req *pb.PrivacySettingReq, miss []*model.MemberPrivacy) error {
	key := d.cacheSFPrivacySetting(req)
	data, err := json.Marshal(miss)
	if err != nil {
		return err
	}
	_, err = d.redis.Do(ctx, "SETEX", key, d.cacheExpire(), data)
	return err
}

func (d *dao) UpdatePrivacySetting(ctx context.Context, req *pb.UpdatePrivacysReq) error {
	var (
		params []string
		args   []interface{}
	)
	for _, v := range req.Settings {
		var status int
		switch v.State {
		case pb.PrivacyState_closed:
			status = 0
		case pb.PrivacyState_opened:
			status = 1
		default:
			continue
		}
		params = append(params, "(?,?,?)")
		args = append(args, req.Mid)
		args = append(args, v.Option.String())
		args = append(args, status)
	}
	_, err := d.db.Exec(ctx, fmt.Sprintf(_privacyBatchAddSQL, privacyHit(req.Mid), strings.Join(params, ",")), args...)
	return err
}

func (d *dao) keyLivePlaybackWhitelist() string {
	return "live_back_list"
}

func (d *dao) CacheLivePlaybackWhitelist(ctx context.Context) (map[int64]struct{}, error) {
	key := d.keyLivePlaybackWhitelist()
	reply, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	var data map[int64]struct{}
	if err = json.Unmarshal(reply, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dao) isNewUser(ctx context.Context, mid int64) bool {
	timeValue, err := time.ParseInLocation("2006-01-02 15:04:05", d.settingNewUserTimePoint, time.Local)
	if err != nil {
		log.Error("日志告警 isNewUser时间错误,error:%+v", err)
		return false
	}
	hours := time.Since(timeValue).Hours()
	if hours <= 0 {
		return false
	}
	periods := fmt.Sprintf("0-%d", int64(math.Ceil(hours)))
	reply, err := d.accClient.CheckRegTime(ctx, &accgrpc.CheckRegTimeReq{Mid: mid, Periods: periods})
	if err != nil {
		log.Error("%+v", err)
		return false
	}
	return reply.GetHit()
}
