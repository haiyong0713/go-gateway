package service

import (
	"context"
	"go-common/library/log"
)

func (service *Service) SyncActDomainCache() {
	ctx := context.Background()
	log.Infoc(ctx, "SyncActDomainCache start , sync_num:%v", service.c.ActDomainConfig.SyncNum)
	if err := service.dao.SyncActDomainCache(ctx, service.c.ActDomainConfig.SyncNum); err != nil {
		log.Errorc(ctx, "SyncActDomainCache:%v", err)
	}
}
