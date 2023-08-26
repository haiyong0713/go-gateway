package domain

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	mdomain "go-gateway/app/web-svr/activity/admin/model/domain"
	"go-gateway/app/web-svr/activity/ecode"
)

func (service *Service) syncCache(ctx context.Context, record *mdomain.Record) (rows int64, err error) {
	if record == nil || record.Id <= 0 {
		return
	}

	//强行设置一个递增版本号
	record.Mtime = (xtime.Time)(time.Now().Unix())
	if err = service.dao.CacheDomain(ctx, record); err == nil {
		// 更新status
		rows, err = service.dao.UpdateStatus(ctx, DomainStatusSync, record.Id)
	}
	log.Infoc(ctx, "AddDomain sync redis:effect_rows:%v , err:%v", rows, err)
	return
}

func (service *Service) syncFawkes(ctx context.Context, forceUpdate bool) (err error) {
	var (
		interRecords, fawsRcords []mdomain.Record
	)

	if interRecords, err = service.dao.GetDomainList(ctx, service.c.ActDomainConf.DefaultPageNo,
		service.c.ActDomainConf.DefaultPageSize); err != nil {
		log.Errorc(ctx, "syncFawkes GetDomainList from activity-interface :%v", err)
		return
	}
	sort.Slice(interRecords, func(i, j int) bool {
		return interRecords[i].Id < interRecords[j].Id
	})
	var updateKeys, appkeys []string
	appkeys = strings.Split(service.c.ActDomainConf.FawkesConf.AppKey, ",")

	for _, v := range appkeys {
		if !forceUpdate {
			if fawsRcords, err = service.dao.GetFawkesConfig(ctx, v); err != nil {
				log.Errorc(ctx, "syncFawkes GetFawkesConfig from fawkes :%v", err)
				continue
			}
			sort.Slice(fawsRcords, func(i, j int) bool {
				return fawsRcords[i].Id < fawsRcords[j].Id
			})
			// reflect.DeepEqual  深度比较 , 前提是两个slice，已经排序了
			if reflect.DeepEqual(interRecords, fawsRcords) {
				log.Infoc(ctx, "syncFawkes FawkesConfig == GetDomainList , exist!")
				continue
			}
			log.Infoc(ctx, "syncFawkes FawkesConfig compare app_key:%v , GetDomainList:%v , GetFawkesConfig:%v", v, interRecords, fawsRcords)
		}
		updateKeys = append(updateKeys, v)
	}
	return service.dao.AddFawkesConfig(ctx, interRecords, updateKeys)
}

func (service *Service) syncAll(ctx context.Context, record *mdomain.Record) (err error) {
	if _, err = service.syncCache(ctx, record); err == nil {
		err = service.syncFawkes(ctx, true)
	}
	if err != nil {
		log.Errorc(ctx, "syncAll cache or fawkes failed:%v", err)
	}
	return
}

// AddDomain 增加
func (service *Service) AddDomain(ctx context.Context, param *mdomain.AddDomainParam) (id int64, err error) {

	id, err = service.dao.AddRecord(ctx, param)
	if err != nil || id <= 0 {
		log.Warnc(ctx, "AddDomain failed : %v , %v , %v ", param, id, err)
		return -1, ecode.ActivityDomainAddError
	}

	if err1 := service.syncAll(ctx, &mdomain.Record{
		Id:           id,
		ActName:      param.ActName,
		PageLink:     param.PageLink,
		FirstDomain:  param.FirstDomain,
		SecondDomain: param.SecondDomain,
		Stime:        param.Stime,
		Etime:        param.Etime,
		Ctime:        (xtime.Time)(time.Now().Unix()),
		Mtime:        (xtime.Time)(time.Now().Unix()),
	}); err1 != nil {
		log.Warnc(ctx, "AddDomain syncAll err :%v", err1)
	}

	return
}

func (service *Service) EditDomain(ctx context.Context, param *mdomain.Record) (rows int64, err error) {
	rows, err = service.dao.UpdateRecord(ctx, param, DomainStatusInit)
	if err != nil || rows <= 0 {
		return
	}
	if err1 := service.syncAll(ctx, param); err1 != nil {
		log.Warnc(ctx, "EditDomain syncAll err :%v", err1)
	}
	return
}

func (service *Service) StopDomain(ctx context.Context, aid int64) (rows int64, err error) {
	rows, err = service.dao.UpdateEtime(ctx, DomainStatusInit, aid)
	if err != nil || rows <= 0 {
		return
	}

	if list, _, err := service.dao.Search(ctx, aid, "", 1, 0); err == nil && len(list) == 1 {
		record := list[0]
		record.Etime = (xtime.Time)(time.Now().Unix())
		if err1 := service.syncAll(ctx, record); err1 != nil {
			log.Warnc(ctx, "StopDomain syncAll err :%v", err1)
		}
	}
	return
}

func (service *Service) SearchDomain(ctx context.Context, param *mdomain.Search) (list []*mdomain.Record, total int, err error) {
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}
	return service.dao.Search(ctx, param.Id, param.ActName, param.PageSize, offset)
}

func (service *Service) SyncScript(ctx context.Context, syncNum int) (rows int64, err error) {
	list, err := service.dao.SyncFailedList(ctx, syncNum)
	log.Infoc(ctx, "SyncScript start, DB sync failed num:%v , err:%v", len(list), err)
	if err != nil {
		return
	}

	for _, value := range list {
		if effect, err := service.syncCache(ctx, value); err == nil {
			rows += effect
		}
	}

	// 本次操作之后，DB和redis已经保持一致了
	if len(list) == 0 || (len(list) < syncNum && rows == int64(len(list))) {
		if err = service.syncFawkes(ctx, false); err != nil {
			log.Errorc(ctx, "SyncScript 2 syncFawkes  err:%v", err)
		}
	}
	return
}
