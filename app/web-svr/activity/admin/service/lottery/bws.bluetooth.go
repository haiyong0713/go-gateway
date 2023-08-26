package lottery

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/lottery"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

func (s *Service) AddBluetoothUps(c context.Context, bid int64, ups []*bwsmdl.BluetoothUp) error {
	tx, err := s.lotDao.BeginTran(c)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err := s.lotDao.DelBluetoothUp(tx, bid); err != nil {
		log.Error("%+v", err)
		return err
	}
	for _, u := range ups {
		if err := s.lotDao.InBluetoothUp(tx, bid, u.Mid, u.Key, u.Desc); err != nil {
			log.Error("%+v", err)
			return err
		}
	}
	return nil
}

func (s *Service) BluetoothUpList(c context.Context, param *lottery.BluetoothUpListParam) ([]*bwsmdl.BluetoothUp, int, error) {
	start := (param.Pn - 1) * param.Ps
	res, err := s.lotDao.BluetoothUpLimit(c, param.Bid, start, param.Ps)
	if err != nil {
		log.Error("s.lotDao.BluetoothUpLimit error(%v)", err)
		return nil, 0, err
	}
	count, err := s.lotDao.BluetoothUpCount(c, param.Bid)
	if err != nil {
		log.Error("s.lotDao.BluetoothUpCount error(%v)", err)
		return nil, 0, err
	}
	return res, count, nil
}

func (s Service) SaveBluetoothUp(c context.Context, param *lottery.EditBluetoothUpParam) error {
	if err := s.lotDao.UpBluetoothUp(c, param.ID, param.Mid, param.Key, param.Desc); err != nil {
		log.Error("s.lotDao.UpBluetoothUp error(%v)", err)
		return err
	}
	return nil
}

func (s Service) DelBluetoothUp(c context.Context, param *lottery.EditBluetoothUpParam) error {
	if err := s.lotDao.DelBluetoothUpByID(c, param.ID); err != nil {
		log.Error("s.lotDao.UpBluetoothUp error(%v)", err)
		return err
	}
	return nil
}
