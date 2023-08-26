package like

import (
	"context"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	guessBizDao "go-gateway/app/web-svr/activity/interface/dao/guess"
	"go-gateway/app/web-svr/activity/interface/model/guess"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const (
	memoryCacheKey4MainAndDetail = "guess_main_res"
)

var (
	hotMainDetailMap map[string]map[int64]*guess.MainRes
	hotMainMap       map[string][]*guess.MainID
)

func init() {
	hotMainDetailMap = make(map[string]map[int64]*guess.MainRes, 0)
	hotMainMap = make(map[string][]*guess.MainID, 0)
}

func (s *Service) GuessListV2(ctx context.Context, mid, oid, business int64) (rs *pb.GuessListReply, err error) {
	var (
		mainIDs []*guess.MainID
		mdRes   map[int64]*guess.MainRes
		mds     []*pb.GuessList
		rsTmp   map[int64][]*pb.GuessList
		ok      bool
	)
	rs = &pb.GuessListReply{}

	mdRes, mainIDs, ok = hotMainListAndDetailListByOIDAndBusiness(oid, business)
	if !ok {
		if mdRes, mainIDs, err = s.NewMDResult(ctx, business, oid, false); err != nil {
			log.Error("s.NewMDResult business(%d) oid(%d) error(%v)", business, oid, err)
			err = ecode.ActGuessesFail

			return
		}
		count := len(mdRes)
		if count == 0 {
			err = xecode.NothingFound

			return
		}
	}

	rsTmp = s.mdResult(ctx, mdRes, mid, mainIDs)
	for _, main := range mainIDs {
		mds = append(mds, rsTmp[main.ID]...)
	}

	rs.MatchGuess = mds

	return
}

func hotMainListAndDetailListByOIDAndBusiness(oID, business int64) (m map[int64]*guess.MainRes, list []*guess.MainID, ok bool) {
	status := tool.StatusOfHit
	list, ok = hotMainListByOIDAndBusiness(oID, business)
	m, ok = hotMainDetailListByOIDAndBusiness(oID, business)
	if !ok {
		status = tool.StatusOfMiss
	}

	tool.IncrMemoryCacheHitOrMissMetric(memoryCacheKey4MainAndDetail, status)

	return
}

func hotMainListByOIDAndBusiness(oID, business int64) (list []*guess.MainID, ok bool) {
	list = make([]*guess.MainID, 0)
	key := guess.GenHotMapKeyByOIDAndBusiness(oID, business)
	if d, tmpOK := hotMainMap[key]; tmpOK {
		ok = tmpOK
		for _, v := range d {
			list = append(list, v.DeepCopy())
		}
	}

	return
}

func hotMainDetailListByOIDAndBusiness(oID, business int64) (m map[int64]*guess.MainRes, ok bool) {
	m = make(map[int64]*guess.MainRes)

	key := guess.GenHotMapKeyByOIDAndBusiness(oID, business)
	if d, tmpOK := hotMainDetailMap[key]; tmpOK {
		ok = tmpOK
		for k, v := range d {
			m[k] = v.DeepCopy()
		}
	}

	return
}

func (s *Service) ASyncHotMainDetailList(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		_ = ticker.Stop
	}()

	for {
		select {
		case <-ticker.C:
			s.resetHotGuessRelations(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) resetHotGuessRelations(ctx context.Context) {
	if list, m, err := s.guessDao.HotMainResMap(ctx); err == nil {
		hotMainMap = guessBizDao.GenHotMainMapByMainIDList(list)
		hotMainDetailMap = m
	}
}
