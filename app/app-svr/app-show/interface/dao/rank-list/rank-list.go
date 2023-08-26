package ranklist

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-show/interface/conf"
	model "go-gateway/app/app-svr/app-show/interface/model/rank-list"

	"github.com/pkg/errors"
)

func keyRankLinkMeta(id int64) string {
	return fmt.Sprintf("rank_list_meta_%d", id)
}

// Dao is
type Dao struct {
	redis  *redis.Redis
	actRPC actgrpc.ActivityClient
}

func New(c *conf.Config) *Dao {
	actRPC, err := actgrpc.NewClient(c.ActivityGRPC)
	if err != nil {
		panic(err)
	}
	return &Dao{
		redis:  redis.NewRedis(c.Redis.Recommend.Config),
		actRPC: actRPC,
	}
}

// RankMeta is
func (d *Dao) RankMeta(ctx context.Context, id int64) (*model.Meta, error) {
	out := &model.Meta{}

	b, err := redis.Bytes(d.redis.Do(ctx, "GET", keyRankLinkMeta(id)))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return nil, errors.WithStack(err)
	}

	sort.Slice(out.FinalRank, func(i, j int) bool {
		return out.FinalRank[i].Position < out.FinalRank[j].Position
	})
	return out, nil
}

// UpActivityArchive is
func (d *Dao) UpActivityArchive(ctx context.Context, mid int64, actIDs []int64) []int64 {
	lock := sync.Mutex{}
	aidCtr := map[int64][]int64{}

	eg := errgroup.WithContext(ctx)
	for _, actID := range actIDs {
		actID := actID
		eg.Go(func(ctx context.Context) error {
			req := &actgrpc.ListActivityArcsReq{
				Sid: actID,
				Mid: mid,
			}
			reply, err := d.actRPC.ListActivityArcs(ctx, req)
			if err != nil {
				log.Error("Failed to list activity arcs: %+v: %+v", req, err)
				return nil
			}
			lock.Lock()
			aidCtr[actID] = reply.Aid
			lock.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute errgroup: %+v", err)
		return nil
	}

	sort.Slice(actIDs, func(i, j int) bool {
		return actIDs[i] > actIDs[j]
	})
	out := []int64{}
	for _, actID := range actIDs {
		out = append(out, aidCtr[actID]...)
	}
	return out
}
