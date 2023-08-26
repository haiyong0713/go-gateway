package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/esports/service/internal/model"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

func teamCacheKey(id int64) string {
	return fmt.Sprintf(teamInfoMapCache, id)
}

func (d *dao) GetTeamsCache(ctx context.Context, teamIds []int64) (teamsInfoMap map[int64]*model.TeamModel, missIds []int64, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, teamId := range teamIds {
		key = teamCacheKey(teamId)
		args = args.Add(key)
	}
	if bss, err = redis.ByteSlices(d.redis.Do(ctx, "MGET", args...)); err != nil {
		log.Errorc(ctx, "GetTeamsCache d.redis.Do(ctx, MGET) error(%v)", err)
		return
	}
	teamsInfoMap = make(map[int64]*model.TeamModel)
	missIds = make([]int64, 0)
	for index, bs := range bss {
		team := new(model.TeamModel)
		if len(bs) == 0 {
			missIds = append(missIds, teamIds[index])
			continue
		}
		if err = json.Unmarshal(bs, team); err != nil {
			log.Error("GetTeamsCache json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		teamsInfoMap[team.ID] = team
	}
	return
}

func (d *dao) SetTeamsCache(ctx context.Context, teamInfoMap map[int64]*model.TeamModel) (err error) {
	if len(teamInfoMap) == 0 {
		return
	}
	var (
		bs       []byte
		keyID    string
		keyIDs   []string
		argsTeam = redis.Args{}
	)
	for _, v := range teamInfoMap {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("SetTeamsCache json.Marshal error(%v)", err)
			continue
		}
		keyID = teamCacheKey(v.ID)
		keyIDs = append(keyIDs, keyID)
		argsTeam = argsTeam.Add(keyID).Add(string(bs))
	}
	conn := d.redis.Conn(ctx)
	defer d.connClose(ctx, conn)
	if err = conn.Send("MSET", argsTeam...); err != nil {
		log.Error("SetTeamsCache conn.Send(MSET) error(%v)", err)
		return
	}
	for _, v := range keyIDs {
		tmpKey := v
		if err = conn.Send("EXPIRE", tmpKey, 86400); err != nil {
			return err
		}
	}
	err = conn.Flush()
	if err != nil {
		log.Errorc(ctx, "[Dao][SetTeamsCache][Flush][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) DeleteTeamCache(ctx context.Context, teamId int64) (err error) {
	redisKey := teamCacheKey(teamId)
	_, err = d.redis.Do(ctx, "del", redisKey)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][DeleteTeamCache][Error], err:%+v", err)
		return
	}
	return
}
