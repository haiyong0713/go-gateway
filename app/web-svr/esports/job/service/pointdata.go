package service

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/log"
	mdlesp "go-gateway/app/web-svr/esports/job/model"
)

func (s *Service) bind(match *mdlesp.ContestData) (err error) {
	var res []byte
	if s.clientID == "" || match.MatchID == 0 {
		return
	}
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("match_id", strconv.FormatInt(match.MatchID, 10))
	params.Set("key", s.c.Leidata.Key)
	for {
		if res, err = s.dao.ThirdPost(context.Background(), params); err == nil && string(res) == "1" {
			break
		}
		if time.Now().Unix() > match.Stime {
			break
		}
		time.Sleep(time.Minute)
	}
	if err != nil || string(res) != "1" {
		log.Error("bind s.dao.ThirdPost error(%+v) res(%s) clientID(%s) matchID(%d)", err, string(res), s.clientID, match.MatchID)
		return
	}
	log.Info("bind success res(%s) clientID(%s) matchID(%d)", string(res), s.clientID, match.MatchID)
	return
}

func (s *Service) matchs(ctx context.Context) (err error) {
	var (
		matchs    []*mdlesp.ContestData
		startTime time.Time
		isBind    bool
	)
	for {
		for i := 0; i < _tryTimes; i++ {
			if matchs, err = s.dao.ContPoints(context.Background()); err != nil {
				log.Errorc(ctx, "s.dao.ContPoints error(%+v)", err)
				time.Sleep(time.Millisecond * 100)
				continue
			}
			break
		}
		for _, match := range matchs {
			if match.Stime > 0 {
				startTime = time.Unix(match.Stime, 0).Add(time.Duration(s.c.Leidata.BindTime))
				if time.Now().Unix() < startTime.Unix() {
					continue
				}
			}
			if match.Etime > 0 && time.Now().Unix() > match.Etime {
				continue
			}
			if match.MatchID > 0 {
				if _, ok := s.matchIDs.Data[match.MatchID]; !ok {
					isBind = true
					log.Infoc(ctx, "matchs add matchid(%d) cid(%d) cliendID(%s)", match.MatchID, match.CID, s.clientID)
					s.matchIDs.Lock()
					s.matchIDs.Data[match.MatchID] = match
					s.matchIDs.Unlock()
				}
			}
		}
		if isBind {
			if s.clientID != "" {
				for _, match := range s.matchIDs.Data {
					tmpM := match
					log.Warnc(ctx, "matchs init match_id(%d) clientID(%s)", tmpM.MatchID, s.clientID)
					go s.bind(tmpM)
				}
			}
		}
		time.Sleep(time.Duration(s.c.Leidata.RecentSleep))
	}
}
