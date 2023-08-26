package service

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/taishan"
	"go-common/library/log"

	"github.com/pkg/errors"
)

func (s *Service) setTaishan(c context.Context, key, val []byte) error {
	req := &taishan.PutReq{
		Table: s.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: s.Taishan.tableCfg.Token,
		},
	}
	req.Record = &taishan.Record{Key: key, Columns: []*taishan.Column{{Value: val}}}
	resp, err := s.Taishan.client.Put(c, req)
	if err != nil {
		return err
	}
	if resp.GetStatus() == nil {
		return errors.New("response status is invalid")
	}
	if resp.GetStatus().ErrNo != 0 {
		return errors.Errorf("key: %+v, errno: %+v, errmsg: %+v", string(key), resp.Status.ErrNo, resp.Status.Msg)
	}
	return nil
}

func (s *Service) delTaishan(c context.Context, keys [][]byte) error {
	req := &taishan.BatchDelReq{
		Table: s.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: s.Taishan.tableCfg.Token,
		},
	}
	records := make([]*taishan.Record, 0)
	for _, k := range keys {
		records = append(records, &taishan.Record{Key: k})
	}
	req.Records = records
	resp, err := s.Taishan.client.BatchDel(c, req)
	if err != nil {
		return err
	}
	if resp.AllSucceed {
		return nil
	}
	var errs []error
	for _, v := range resp.Records {
		if v.Status.ErrNo != 0 {
			errs = append(errs, errors.Errorf("key: %+v, errno: %+v, errmsg: %+v", string(v.Key), v.Status.ErrNo, v.Status.Msg))
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("%+v", errs)
	}
	return nil
}

// nolint:gocognit
func (s *Service) initAllArcTaishan() {
	maxID, _ := strconv.ParseInt(os.Getenv("TAISHANMAXID"), 10, 64)
	conn := s.rds.Get(context.Background())
	id, _ := redis.Int64(conn.Do("GET", "TAISHANMAXID"))
	conn.Close()
	if id > 0 {
		maxID = id
	}
	log.Warn("initAllArcTaishan start at id(%d)", maxID)
	for i := 0; i < 20; i++ {
		// nolint:biligowordcheck
		go func() {
			for {
				conn := s.rds.Get(context.Background())
				id, err := redis.Int64(conn.Do("DECR", "TAISHANMAXID"))
				conn.Close()
				if err != nil {
					log.Error("%+v", err)
					continue
				}
				if id <= 0 {
					log.Warn("taishan all arc init complete")
					break
				}
				for {
					time.Sleep(5 * time.Millisecond)
					err := func() error {
						aid, err := s.dao.IDToAid(context.Background(), id)
						if err != nil {
							if err == sql.ErrNoRows {
								return nil
							}
							return err
						}
						if aid == 0 {
							log.Error("日志告警 不可能存在的id(%d)对应aid=0的情况发生了", id)
							return nil
						}
						arc, ip, err := s.dao.Archive(context.Background(), aid)
						if err != nil || arc == nil {
							log.Error("s.dao.Archive err(%+v) or arc not exist(%d)", err, aid)
							return err
						}
						s.transIpv6ToLocation(context.Background(), arc, ip)
						if err = s.setArcCache(context.Background(), arc); err != nil {
							log.Error("s.setArcCache err(%+v) aid(%d)", err, aid)
							return err
						}
						vs, err := s.dao.Videos(context.Background(), arc.Aid)
						if err != nil {
							log.Error("s.dao.Videos err(%+v) aid(%d)", err, aid)
							return err
						}
						if err = s.setVideosPageCache(context.Background(), aid, vs); err != nil {
							log.Error("s.setVideosPageCache err(%+v) aid(%d)", err, aid)
							return err
						}
						if err = s.setSimpleArcCache(context.Background(), arc, vs); err != nil {
							log.Error("s.setSimpleArcCache err(%+v) aid(%d)", err, aid)
							return err
						}
						for _, v := range vs {
							s.UpdateVideoCache(context.Background(), aid, v.Cid)
						}
						log.Info("initAllArcTaishan aid(%d) success", aid)
						return nil
					}()
					if err != nil {
						log.Error("%+v", err)
						continue
					}
					break
				}
			}
		}()
	}

}
