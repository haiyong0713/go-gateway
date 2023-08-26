package ugctab

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	errgroup2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/resource/service/model"
	"strings"
	"sync"
)

func (d *Dao) GetEffectiveUgcTab(c context.Context) (ret []*model.UgcTabItem, err error) {
	// 查找缓存
	if len(d.tabCache) == 0 {
		err = ecode.NothingFound
		log.Error("Get dao.GetEffectiveUgcTab error: %s", err)
		return
	}
	ret = d.tabCache
	return ret, err
}

func (d *Dao) UpdateCache(c context.Context) (err error) {
	// 从mysql中取出数据并存储
	var tabs []*model.UgcTabItem
	if tabs, err = d.GetMysqlUgcTab(c); err != nil {
		log.Error("dao.updatetabCache error: %s", err.Error())
		return
	}
	lock := sync.Mutex{}
	eg := errgroup2.WithCancel(c)
	for i, tab := range tabs {
		avidFile := tab.AvidFile
		avid := tab.Avid
		index := i
		if avidFile != "" {
			eg.Go(func(c context.Context) error {
				var (
					avidMap map[string]bool
					avids   string
					e       error
				)
				if avidMap, avids, e = d.FetchAvidFromFile(avidFile); e != nil {
					log.Error("dao.updatetabCache with file error: %s, file path: %s", e.Error(), avidFile)
					return e
				}
				lock.Lock()
				tabs[index].AvidMap = avidMap
				tabs[index].Avid = avids
				log.Error("dao.updatetabCache with file success: len(%+v), file path: %s", len(strings.Split(avids, ",")), avidFile)
				lock.Unlock()
				return nil
			})
		} else if avid != "" {
			eg.Go(func(c context.Context) error {
				var (
					avidMap = make(map[string]bool)
					avidArr = strings.Split(avid, ",")
				)
				for _, avid := range avidArr {
					avidMap[avid] = true
				}
				lock.Lock()
				tabs[index].AvidMap = avidMap
				lock.Unlock()
				return nil
			})
		}
	}
	if err = eg.Wait(); err != nil {
		return
	}

	d.tabCache = tabs
	return
}
