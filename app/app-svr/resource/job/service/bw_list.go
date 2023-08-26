package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/job/dao"
	"go-gateway/app/app-svr/resource/job/model/common"
	"time"
)

func (s *Service) loadBWList() {
	ctx := context.Background()
	now := time.Now()
	// 分页，从第一页开始load数据，避免数据量过大
	pageNum := 0

	list, err := s.dao.GetModifiedBWListItemFromDB(ctx, pageNum)
	if err != nil {
		log.Error("loadBWListItem fail at %s, err is %v, pn is %v", now.Format("2006-01-02 15:04:05"), err.Error(), pageNum)
		return
	}

	for len(list) > 0 {
		log.Error("loadBWListItem success at %s, data length is %v, pn is %v", now.Format("2006-01-02 15:04:05"), len(list), pageNum)
		err = s.dao.SetModifiedBWListItemIntoRedis(ctx, list)
		if err != nil {
			log.Error("loadBWListItem fail at %s, data length is %v, pn is %v", now.Format("2006-01-02 15:04:05"), len(list), pageNum)
		}

		// len不为0的时侯，默认还有下一页
		if len(list) == dao.PageSize {
			pageNum += 1
			list, err = s.dao.GetModifiedBWListItemFromDB(ctx, pageNum)
			if err != nil {
				log.Error("loadBWListItem fail at %s, err is %v, pn is %v", now.Format("2006-01-02 15:04:05"), err.Error(), pageNum)
				break
			}
		} else {
			list = make([]*common.BWListItem, 0)
		}
	}
}
