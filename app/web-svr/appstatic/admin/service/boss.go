package service

import (
	"context"
	"fmt"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/appstatic/admin/model"
)

// BossCdnPreload .
func (s *Service) BossCdnPreload(c context.Context, url string) (err error) {
	//只有正式环境才支持cdn预热
	if env.DeployEnv == env.DeployEnvProd || env.DeployEnv == env.DeployEnvPre {
		return s.dao.CdnDoPreload(c, []string{url})
	}
	return nil
}

// BossCdnStatus .
func (s *Service) BossCdnStatus(c context.Context, urls []string) (map[string]*model.CdnKsyun, error) {
	cdnQuery, err := s.dao.CdnPreloadQuery(c, urls)
	if err != nil {
		log.Error("BossCdnStatus req(%v) error(%s)", urls, err.Error())
		return nil, fmt.Errorf("CDN接口查询失败 错误信息(%s)", err.Error())
	}
	return cdnQuery, nil
}

// BossCdnPublishCheck checkout cdn preload status
func (s *Service) BossCdnPublishCheck(c context.Context, resID int64) error {
	if env.DeployEnv == env.DeployEnvUat {
		return nil
	}
	var (
		files []*model.ResourceFile
		urls  []string
	)
	where := map[string]interface{}{
		"is_deleted":  0,
		"resource_id": resID,
	}
	err := s.DB.Where(where).Where("url like ?", "%boss%").Find(&files).Error
	if err != nil {
		log.Error("BossCdnPublishCheck Find req(%d) error(%s)", resID, err)
		return err
	}
	if len(files) == 0 {
		return nil
	}
	for _, v := range files {
		urls = append(urls, v.URL)
	}
	cdnQuery, err := s.dao.CdnPreloadQuery(c, urls)
	if err != nil {
		if err == ecode.RequestErr {
			return fmt.Errorf("没有查询到CDN预热信息 url(%v)", urls)
		}
		log.Error("BossCdnPublishCheck req(%v) error(%s)", resID, err.Error())
		return fmt.Errorf("CDN接口查询失败 错误信息(%s)", err.Error())
	}
	for _, url := range urls {
		query, ok := cdnQuery[url]
		if !ok {
			return fmt.Errorf("没有查询到(%s)的CDN信息", url)
		}
		if query.Status == "failed" {
			return fmt.Errorf("url(%s)热更新失败res(%v)", url, query)
		}
		if query.Status == "success" && query.Progress == 100 {
			continue
		}
		return fmt.Errorf("url(%s)正在热更新中 res(%v)", url, query)
	}
	return nil
}
