package service

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/database/taishan"
	"go-common/library/log"
)

const _maxTaiShan = 500

func (s *Service) taishanBatchSet(c context.Context, kvMap map[string][]byte) {
	if len(kvMap) == 0 {
		return
	}

	var records []*taishan.Record
	for key, value := range kvMap {
		records = append(records, &taishan.Record{
			Key:     []byte(key),
			Columns: []*taishan.Column{{Value: value}},
		})
	}

	if err := s.batchSetTaishan(c, records); err != nil {
		log.Error("taishanBatchSet kvMap(%+v) err(%+v)", kvMap, err)
	}
}

func (s *Service) batchSetTaishan(c context.Context, records []*taishan.Record) error {
	for i := 0; i < len(records); i += _maxTaiShan {
		var partRecord []*taishan.Record
		if i+_maxTaiShan > len(records) {
			partRecord = records[i:]
		} else {
			partRecord = records[i : i+_maxTaiShan]
		}
		if err := s._batchSetTaish(c, partRecord); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) _batchSetTaish(c context.Context, records []*taishan.Record) error {
	req := &taishan.BatchPutReq{
		Table: s.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: s.Taishan.tableCfg.Token,
		},
	}
	req.Records = records
	resp, err := s.Taishan.client.BatchPut(c, req)
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

func (s *Service) delTaishan(c context.Context, keys []*taishan.Record) error {
	for i := 0; i < len(keys); i += _maxTaiShan {
		var partRecord []*taishan.Record
		if i+_maxTaiShan > len(keys) {
			partRecord = keys[i:]
		} else {
			partRecord = keys[i : i+_maxTaiShan]
		}
		if err := s._batchDelTaishan(c, partRecord); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) _batchDelTaishan(c context.Context, records []*taishan.Record) error {
	req := &taishan.BatchDelReq{
		Table: s.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: s.Taishan.tableCfg.Token,
		},
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

//func (s *Service) initAllArcTaishan() {
//	maxID, _ := strconv.ParseInt(os.Getenv("TAISHANMAXID"), 10, 64)
//	conn := s.redis.Get(context.Background())
//	id, _ := redis.Int64(conn.Do("GET", "TAISHANMAXID"))
//	conn.Close()
//	if id > 0 {
//		maxID = id
//	}
//	log.Warn("initAllArcTaishan start at id(%d)", maxID)
//	for i := 0; i < 16; i++ {
//		go func() {
//			for {
//				conn := s.redis.Get(context.Background())
//				id, err := redis.Int64(conn.Do("DECR", "TAISHANMAXID"))
//				conn.Close()
//				if err != nil {
//					log.Error("%+v", err)
//					continue
//				}
//				if id <= 0 {
//					log.Warn("taishan all arc init complete")
//					break
//				}
//				for {
//					time.Sleep(10 * time.Millisecond)
//					err := func() error {
//						ctx := context.Background()
//						aid, err := s.resultDao.IDToAid(ctx, id)
//						if err != nil {
//							if err == sql.ErrNoRows {
//								return nil
//							}
//							return err
//						}
//						if aid == 0 {
//							log.Error("日志报警 不可能存在的id(%d)对应aid=0的情况发生了", id)
//							return nil
//						}
//						// upVideoCache 内部有 retry 逻辑，此处不接收err，方法内重试
//						s.upVideoCache(ctx, aid)
//						arc, err := s.resultDao.RawArc(ctx, aid)
//						if err != nil || arc == nil {
//							return err
//						}
//						vs, err := s.resultDao.RawVideos(ctx, aid)
//						if err != nil {
//							return err
//						}
//						var delTai []*taishan.Record
//						for _, v := range vs {
//							delTai = append(delTai, &taishan.Record{Key: []byte(arcmdl.VideoKey(aid, v.Cid))})
//						}
//						err = s.delTaishan(ctx, delTai)
//						if err != nil {
//							return err
//						}
//						if err = s.setVideosPageCache(ctx, aid, vs); err != nil {
//							return err
//						}
//						log.Info("initAllArcTaishan id(%d) success", aid)
//						return nil
//					}()
//					if err != nil {
//						log.Error("initAllArcTaishan error:%+v", err)
//						continue
//					}
//					break
//				}
//			}
//		}()
//	}
//}
