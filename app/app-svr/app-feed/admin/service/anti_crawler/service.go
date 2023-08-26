package anti_crawler

import (
	"context"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	antidao "go-gateway/app/app-svr/app-feed/admin/dao/anti_crawler"
	feadao "go-gateway/app/app-svr/app-feed/admin/dao/feature"
	model "go-gateway/app/app-svr/app-feed/admin/model/anti_crawler"
	feamdl "go-gateway/app/app-svr/app-feed/admin/model/feature"

	"github.com/robfig/cron"
)

// Service audit service.
type Service struct {
	dao    *antidao.Dao
	feaDao *feadao.Dao
	cache  *fanout.Fanout
	cron   *cron.Cron
}

// New new a audit service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:    antidao.New(c),
		feaDao: feadao.New(c),
		cache:  fanout.New("cache"),
		cron:   cron.New(),
	}
	s.initCron()
	return
}

func (s *Service) initCron() {
	s.cron.AddFunc("@every 5m", func() {
		if err := s.businessConfigproc(context.Background()); err != nil {
			log.Error("%+v", err)
		}
	})
	s.cron.Start()
}

func (s *Service) Close() {
	s.cron.Stop()
	s.cache.Close()
}

func (s *Service) UserLog(ctx context.Context, v *model.UserLogParam) ([]*model.InfocMsg, error) {
	data, err := s.dao.UserLog(ctx, v.Buvid, v.Mid, v.ReqHost, v.Path, v.Stime, v.Etime, v.Page, v.PerPage)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return []*model.InfocMsg{}, nil
	}
	return data, nil
}

func (s *Service) BusinessConfigList(ctx context.Context, req *model.BusinessConfigListReq) ([]*model.WList, error) {
	data, err := func() ([]*model.WList, error) {
		if req.Value == "" {
			return s.dao.WListAllCache(ctx)
		}
		v, err := s.dao.WListCache(ctx, req.Value)
		if err != nil {
			return nil, err
		}
		return []*model.WList{v}, nil
	}()
	if err != nil {
		return nil, err
	}
	if data == nil {
		return []*model.WList{}, nil
	}
	return data, nil
}

func (s *Service) BusinessConfigUpdate(ctx context.Context, req *model.BusinessConfigUpdateReq) error {
	data := &model.WList{
		Value:    req.Value,
		Forever:  req.Forever,
		Deadline: req.Datetime,
	}
	if err := s.dao.SetWListCache(ctx, data); err != nil {
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.businessConfigSet(ctx); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) BusinessConfigDelete(ctx context.Context, req *model.BusinessConfigDeleteReq) error {
	if err := s.dao.DelWListCache(ctx, req.Value); err != nil {
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.businessConfigSet(ctx); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) businessConfigproc(ctx context.Context) error {
	const key = "w_list_key"
	locked, err := s.dao.TryLock(ctx, key, 60)
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer func() {
		if err1 := s.dao.UnLock(ctx, key); err1 != nil {
			log.Error("%+v", err1)
		}
	}()
	return s.businessConfigSet(ctx)
}

func (s *Service) businessConfigSet(ctx context.Context) error {
	req := &feamdl.BusinessConfigListReq{
		TreeID:  999999,
		KeyName: "common.mogul",
	}
	reply, _, err := s.feaDao.SearchBusinessConfig(ctx, req, false, false)
	if err != nil {
		return err
	}
	var (
		oldVals map[string]struct{}
		config  *feamdl.BusinessConfig
	)
	func() {
		if len(reply) == 0 {
			return
		}
		config = reply[0]
		vs := strings.Split(config.Config, ",")
		vals := map[string]struct{}{}
		for _, v := range vs {
			vals[v] = struct{}{}
		}
		oldVals = vals
	}()
	data, err := s.dao.WListAllCache(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	var vals []string
	for _, val := range data {
		delete(oldVals, val.Value)
		if val.Forever == 1 {
			vals = append(vals, val.Value)
			continue
		}
		t := time.Unix(val.Deadline, 0)
		if now.Before(t) {
			vals = append(vals, val.Value)
		}
	}
	var missVals []*model.WList
	for val := range oldVals {
		vals = append(vals, val)
		missVals = append(missVals, &model.WList{
			Value:    val,
			Forever:  1,
			Deadline: 0,
		})
	}
	newCfg := strings.Join(vals, ",")
	if err := s.businessConfigSave(ctx, config, newCfg); err != nil {
		return err
	}
	if err := s.dao.SetWListCache(ctx, missVals...); err != nil {
		return err
	}
	return nil
}

func (s *Service) businessConfigSave(ctx context.Context, c *feamdl.BusinessConfig, newCfg string) error {
	param := &feamdl.BusinessConfig{
		ID:            c.ID,
		TreeID:        c.TreeID,
		KeyName:       c.KeyName,
		Config:        newCfg,
		Description:   c.Description,
		Relations:     c.Relations,
		Creator:       c.Creator,
		CreatorUID:    c.CreatorUID,
		Modifier:      c.Modifier,
		ModifierUID:   c.ModifierUID,
		State:         c.State,
		Ctime:         c.Ctime,
		WhiteListType: c.WhiteListType,
		WhiteList:     c.WhiteList,
	}
	log.Info("businessConfigSave %+v", param)
	_, err := s.feaDao.BusinessConfigSave(ctx, param)
	return err
}
