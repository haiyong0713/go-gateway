package service

import (
	"context"
	"time"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/appstatic/admin/model"
)

const _chronosResID = 0

func (s *Service) UploadChronos(c context.Context, fileBody []byte) (*model.UploadReply, error) {
	var location string
	if err := func() error {
		fInfo, err := s.ParseFile(fileBody)
		if err != nil {
			return err
		}
		if location, err = s.UploadBfs(c, "", fInfo.Type, time.Now().Unix(), fileBody); err != nil {
			return err
		}
		fInfo.URL = location // 覆盖name
		if err = s.DB.Where("resource_id=?", _chronosResID).Where("url=?", location).
			FirstOrCreate(transFile(fInfo, _chronosResID)).Error; err != nil { // 保存md5和文件名到，重复上传的文件不存
			return err
		}
		return nil
	}(); err != nil {
		log.Error("UploadChronos err %+v", err)
		return nil, err
	}
	return &model.UploadReply{URL: location}, nil
}

func (s *Service) SaveRules(c context.Context, rules []*model.ChronosRule) error {
	eg := errgroup.WithCancel(c)

	eg.Go(func(c context.Context) error { // 保存本次提交到缓存
		return s.dao.SaveRawRules(c, rules)
	})

	md5Map := make(map[string]string)
	eg.Go(func(c context.Context) error { // 根据文件名找对应的md5信息
		filesNames := make([]string, 0)
		for _, v := range rules {
			filesNames = append(filesNames, v.File)
		}
		files := make([]*model.ResourceFile, 0)
		if err := s.DB.Model(&model.ResourceFile{}).Where("url IN (?)", filesNames).
			Where("is_deleted=0").Where("resource_id=?", _chronosResID).Find(&files).Error; err != nil {
			return err
		}
		for _, v := range files {
			md5Map[v.URL] = v.Md5
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Error("UploadChronos saveRules err %+v", err)
		return err
	}

	// 拼接app-player的结构并保存到app-player的redis中
	playerRules := make([]*model.PlayerRule, 0)
	for _, v := range rules {
		if v.File == "" { // 还没有上传文件的规则不下发
			continue
		}
		if v.Gray == 0 { // 尚未开始灰度的不下发
			continue
		}
		md5, ok := md5Map[v.File]
		if !ok { // 未找到对应md5的不下发
			continue
		}
		playerRules = append(playerRules, &model.PlayerRule{
			ChronosRule: *v,
			MD5:         md5,
		})
	}
	if len(playerRules) == 0 { // 无可用规则，不保存redis。日志报警打点，但是往下走让他清空缓存
		log.Error("UploadChronos PlayerRules_Empty")
	}
	if err := s.dao.SavePlayerRules(c, playerRules); err != nil {
		log.Error("UploadChronos saveRules err %+v", err)
		return err
	}
	return nil
}

func (s *Service) ListChronos(c context.Context) (data []*model.ChronosRule, err error) {
	return s.dao.RawRules(c)
}

func (s *Service) loadPackageInfoForAppView() {
	allPackageInfo, err := s.dao.FetchAllPackageByAppAndService()
	if err != nil {
		log.Error("loadPackageInfoForAppView error(%+v)", err)
		return
	}
	if err := s.dao.SavePackageInfoRulesToAppView(context.Background(), allPackageInfo); err != nil {
		log.Error("s.dao.SavePackageInfoRulesToAppView error(%+v) packageInfo(%+v)", err, allPackageInfo)
		return
	}
	log.Info("loadPackageInfoForAppView success info(%+v)", allPackageInfo)
}
