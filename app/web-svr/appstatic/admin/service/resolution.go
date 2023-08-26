package service

import (
	"path"

	"go-gateway/app/web-svr/appstatic/admin/model"

	"go-common/library/log"

	"github.com/pkg/errors"
)

func (s *Service) FetchDolbyWhiteList() ([]*model.DolbyWhiteList, error) {
	return s.dao.FetchDolbyWhiteList()
}

func fileHash(bfsPath string) string {
	fileNameAll := path.Base(bfsPath)
	fileSuffix := path.Ext(bfsPath)
	return fileNameAll[0 : len(fileNameAll)-len(fileSuffix)]
}

func (s *Service) AddDolbyWhiteList(in *model.DolbyWhiteList) error {
	in.BFSPathHash = fileHash(in.BFSPath)
	return s.dao.AddDolbyWhiteList(in)
}

func (s *Service) SaveDolbyWhiteList(in *model.DolbyWhiteList) error {
	in.BFSPathHash = fileHash(in.BFSPath)
	return s.dao.SaveDolbyWhiteList(in)
}

func (s *Service) DelDolbyWhiteList(id int64) error {
	return s.dao.DeleteDolbyWhiteList(id)
}

func (s *Service) FetchQnBlackList() ([]*model.QnBlackList, error) {
	return s.dao.FetchQnBlackList()
}

func (s *Service) AddQnBlackList(in *model.QnBlackList) error {
	return s.dao.AddQnBlackList(in)
}

func (s *Service) SaveQnBlackList(in *model.QnBlackList) error {
	return s.dao.SaveQnBlackList(in)
}

func (s *Service) DeleteQnBlackList(id int64) error {
	return s.dao.DeleteQnBlackList(id)
}

func (s *Service) AddLimitFree(in *model.LimitFreeInfo) error {
	//是否存在未下线的aid配置
	limitFreeExisted, err := s.dao.FetchLimitFreeByAid(in.Aid)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	if limitFreeExisted.ID != 0 {
		return errors.Errorf("already has limit free")
	}
	return s.dao.AddLimitFreeInfo(in)
}

func (s *Service) EditLimitFree(in *model.LimitFreeInfo) error {
	if in.ID == 0 {
		return errors.Errorf("error id")
	}
	return s.dao.EditLimitFreeInfo(in)
}

func (s *Service) LimitFreeList() ([]*model.LimitFreeInfo, error) {
	return s.dao.FetchLimitFreeList()
}

func (s *Service) DeleteLimitFree(id int64) error {
	return s.dao.DeleteLimitFreeInfo(id)
}
