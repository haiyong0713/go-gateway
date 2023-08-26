package feature

import (
	"context"
	"encoding/json"
	"strings"

	"go-common/library/log"

	commonMdl "go-gateway/app/app-svr/app-feed/admin/model/common"
	featureMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
)

func (s *Service) SwitchTV(c context.Context, req *featureMdl.SwitchTvListReq) (*featureMdl.SwitchTvListReply, error) {
	tmpRes, count, err := s.dao.SwitchTV(c, req, true)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var tvSwitchs []*featureMdl.SwitchTvItem
	for _, tmpRe := range tmpRes {
		i := new(featureMdl.SwitchTvItem)
		i.ID = tmpRe.ID
		i.Brand = tmpRe.Brand
		i.Chid = tmpRe.Chid
		i.Model = tmpRe.Model
		var sysVersion *featureMdl.SysVersion
		if err = json.Unmarshal([]byte(tmpRe.SysVersion), &sysVersion); err != nil {
			log.Warn("switch tv data invalid err %v", err)
			continue
		}
		i.SysVersion = sysVersion
		i.Config = strings.Split(tmpRe.Config, ",")
		i.Deleted = tmpRe.Deleted
		i.Ctime = tmpRe.Ctime
		i.Mtime = tmpRe.Mtime
		i.Description = tmpRe.Description
		tvSwitchs = append(tvSwitchs, i)
	}
	return &featureMdl.SwitchTvListReply{
		Page: &commonMdl.Page{
			Total: count,
			Num:   req.Pn,
			Size:  req.Ps,
		},
		List: tvSwitchs,
	}, nil
}

func (s *Service) SwitchTvSave(c context.Context, req *featureMdl.SwitchTvSaveReq) (string, error) {
	tmpRes, err := s.dao.SwitchTVAll(c)
	if err != nil {
		log.Error("%+v", err)
		return "冲突校验失败：已配数据查询失败", err
	}
	attrs := &featureMdl.SwitchTV{
		ID:          req.ID,
		Brand:       req.Brand,
		Chid:        req.Chid,
		Model:       req.Model,
		SysVersion:  req.SysVersion,
		Config:      req.Config,
		Description: req.Description,
	}
	for _, tmpRe := range tmpRes {
		if tmpRe.ID == req.ID {
			attrs.ID = tmpRe.ID
			attrs.Deleted = tmpRe.Deleted
			attrs.Ctime = tmpRe.Ctime
			continue
		}
	}
	if _, err = s.dao.SaveSwitchTv(c, attrs); err != nil {
		log.Error("s.dao.SaveSwitchTv(%+v) error(%+v)", attrs, req)
		return "", err
	}
	return "", nil
}

func (s *Service) SwitchTvDel(c context.Context, req *featureMdl.SwitchTvDelReq) error {
	attrs := map[string]interface{}{
		"deleted": 1,
	}
	if err := s.dao.UpdateSwitchTv(c, req.ID, attrs); err != nil {
		log.Error("s.dao.UpdateSwitchTv(%+v) error(%+v)", req, req)
		return err
	}
	return nil
}
