package resource

import (
	"context"
	"encoding/json"
	"go-common/library/conf/env"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus/report"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/account"
	dao "go-gateway/app/app-svr/app-feed/admin/dao/resource"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/resource"
	"go-gateway/pkg/idsafe/bvid"
)

// AddCCOperateLog is
func AddCCOperateLog(operator string, operatorID int64, action string, cc *model.CustomConfig) {
	ccs, _ := json.Marshal(cc)
	//nolint:errcheck
	_ = report.Manager(&report.ManagerInfo{
		Uname:    operator,
		UID:      operatorID,
		Business: common.BusinessID,
		Type:     common.LogResourceCustomConfig,
		Oid:      cc.ID,
		Action:   action + string(ccs),
		Ctime:    time.Now(),
		// extra
		Index: []interface{}{
			cc.TP, cc.Oid,
		},
		Content: map[string]interface{}{
			"custom_config": cc,
		},
	})
}

// Service is
type Service struct {
	dao *dao.Dao
	acc *account.Dao
}

// New is
func New(c *conf.Config) *Service {
	s := &Service{
		dao: dao.New(c),
		acc: account.New(c),
	}
	if env.DeployEnv != env.DeployEnvPre {
		//nolint:biligowordcheck
		go s.ConsumeArchiveAuditResult()
	}
	return s
}

func (s *Service) ccListArchive(ctx context.Context, req *model.CCListReq) (*model.CCListArchiveReply, error) {
	//nolint:ineffassign,staticcheck
	ccl, err := s.dao.CCList(ctx, req)
	ccl.Page.Total, err = s.dao.CCListTotal(req)
	if err != nil {
		return nil, err
	}
	avids := make([]int64, 0)
	ccal := &model.CCListArchiveReply{
		Data: make([]*model.CustomConfigArchiveReply, 0, len(ccl.Data)),
		Page: ccl.Page,
	}
	for _, cc := range ccl.Data {
		ccar := &model.CustomConfigArchiveReply{
			CustomConfigReply: *cc,
		}
		ccal.Data = append(ccal.Data, ccar)
		avids = append(avids, cc.Oid)
	}
	if len(avids) > 0 {
		if archiveInfo, err := s.dao.GetArchiveInfo(ctx, avids); err != nil {
			log.Error("Failed to GetArchiveInfo: %d: %+v", avids, err)
			return nil, err
		} else {
			for _, cca := range ccal.Data {
				cca := cca
				if inf, ok := archiveInfo[cca.Oid]; !ok {
					continue
				} else {
					cca.CCArchive = model.CCArchive{
						Aid:   inf.Aid,
						Title: inf.Title,
						Mid:   inf.Author.Mid,
						State: int64(inf.State),
						Attr:  inf.AttributeV2,
					}
					cca.CCAuthor = model.CCAuthor{
						Mid:  inf.Author.Mid,
						Name: inf.Author.Name,
					}
				}
			}
		}
	}
	return ccal, nil
}

// CCAdd is
func (s *Service) CCAdd(ctx context.Context, req *model.CCAddReq) (rows int64, err error) {
	if req.TP == model.CustonConfigTPArchive {
		var sa *model.CCArchive
		sa, err = s.dao.SimpleArchvie(ctx, req.OidNum)
		if err != nil || sa == nil {
			log.Error("Failed to get simple archvie: %d: %+v", req.OidNum, err)
			return 0, err
		}
	}
	if rows, err = s.dao.CCAdd(ctx, req); err != nil {
		return rows, err
	}
	func() {
		cc, err := s.dao.GetCustomConfigBy(ctx, req.TP, req.OidNum)
		if err != nil {
			return
		}
		AddCCOperateLog(req.Operator, req.OperatorID, "新增 ", cc)
	}()
	return rows, nil
}

// CCUpdate is
func (s *Service) CCUpdate(ctx context.Context, req *model.CCUpdateReq) error {
	if req.TP == model.CustonConfigTPArchive {
		sa, err := s.dao.SimpleArchvie(ctx, req.OidNum)
		if err != nil || sa == nil {
			log.Error("Failed to get simple archvie: %d: %+v", req.OidNum, err)
			return err
		}
	}
	if req.AuditCode == 0 {
		if cc, err := s.dao.GetCustomConfig(ctx, req.ID); err != nil {
			log.Error("Failed to get GetCustomConfig: %d: %+v", req.ID, err)
			return err
		} else {
			req.AuditCode = cc.AuditCode
		}
	}
	if err := s.dao.CCUpdate(ctx, req); err != nil {
		return err
	}
	func() {
		cc, err := s.dao.GetCustomConfig(ctx, req.ID)
		if err != nil {
			return
		}
		AddCCOperateLog(req.Operator, req.OperatorID, "编辑 ", cc)
	}()
	return nil
}

func (s *Service) ccEnable(ctx context.Context, id int64, operator string, operatorID int64) error {
	if err := s.dao.CCUpdateState(ctx, id, model.CustomConfigStateEnable); err != nil {
		return err
	}
	func() {
		cc, err := s.dao.GetCustomConfig(ctx, id)
		if err != nil {
			log.Error("ccEnable error: %s", err.Error())
			return
		}
		AddCCOperateLog(operator, operatorID, "上线", cc)
	}()
	return nil
}

func (s *Service) ccDisable(ctx context.Context, id int64, operator string, operatorID int64) error {
	if err := s.dao.CCUpdateState(ctx, id, model.CustomConfigStateDisable); err != nil {
		return err
	}
	func() {
		cc, err := s.dao.GetCustomConfig(ctx, id)
		if err != nil {
			log.Error("ccDisable error: %s", err.Error())
			return
		}
		AddCCOperateLog(operator, operatorID, "下线", cc)
	}()
	return nil
}

func (s *Service) ccAuditCodeUpdate(ctx context.Context, id int64, operator string, operatorID int64, auditCode int32) error {
	if err := s.dao.CCUpdateAuditCode(ctx, id, auditCode); err != nil {
		return err
	}
	func() {
		cc, err := s.dao.GetCustomConfig(ctx, id)
		if err != nil {
			return
		}
		AddCCOperateLog(operator, operatorID, "更改audit_code", cc)
	}()
	return nil
}

// CCLog is
func (s *Service) CCLog(ctx context.Context, req *model.CCLogReq) (*model.CCLogReply, error) {
	cc, err := s.dao.GetCustomConfig(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	logs, err := s.dao.SearchCCLog(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	reply := &model.CCLogReply{
		ID:      cc.ID,
		TP:      cc.TP,
		Oid:     cc.Oid,
		Logging: []*model.CCLog{},
	}
	for _, l := range logs.Result {
		cl := &model.CCLog{
			Operator:   l.Uname,
			OperatorID: l.UID,
			OperateAt:  l.Ctime,
			Operation:  l.Action,
		}
		reply.Logging = append(reply.Logging, cl)
	}
	return reply, nil
}

func asStateDesc(in string) string {
	switch in {
	case model.CustomConfigStatusExpired:
		return "已过期"
	case model.CustomConfigStatusStaging:
		return "待上线"
	case model.CustomConfigStatusOnline:
		return "在线"
	case model.CustomConfigStatusOffline:
		return "已下线"
	}
	return "未知"
}

// ConfigList is
func (s *Service) ConfigList(ctx context.Context, req *model.CCListReq) (*model.ConfigListReply, error) {
	res, err := s.ccListArchive(ctx, req)
	if err != nil {
		return nil, err
	}
	out := []*model.ConfigListItem{}
	for _, cc := range res.Data {
		bv, _ := bvid.AvToBv(cc.CCArchive.Aid)
		item := &model.ConfigListItem{
			ID:               cc.ID,
			Oid:              cc.Oid,
			BVID:             bv,
			Title:            cc.CCArchive.Title,
			Mid:              cc.CCArchive.Mid,
			Name:             cc.CCAuthor.Name,
			OState:           cc.CCArchive.State,
			Content:          cc.Content,
			URL:              cc.URL,
			Image:            cc.Image,
			ImageBig:         cc.ImageBig,
			HighlightContent: cc.HighlightContent,
			STime:            cc.STime.Time().Format("2006-01-02 15:04:05"),
			ETime:            cc.ETime.Time().Format("2006-01-02 15:04:05"),
			State:            cc.State,
			StateDesc:        asStateDesc(cc.Status),
			OriginType:       cc.OriginType,
			AuditCode:        cc.AuditCode,
		}
		out = append(out, item)
	}
	reply := &model.ConfigListReply{
		Data: out,
		Page: res.Page,
	}
	return reply, nil
}

// CCOpt is
func (s *Service) CCOpt(ctx context.Context, req *model.CCOptReq) error {
	switch req.State {
	case model.CustomConfigStateEnable:
		return s.ccEnable(ctx, req.ID, req.Operator, req.OperatorID)
	case model.CustomConfigStateDisable:
		return s.ccDisable(ctx, req.ID, req.Operator, req.OperatorID)
	}
	log.Warn("Unrecognized cc opt request: %+v", req)
	return nil
}

// GetConfig is
func (s *Service) GetConfig(ctx context.Context, req *model.GetConfigReq) (*model.ConfigListItem, error) {
	cc, err := s.dao.GetCustomConfig(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	reply := &model.ConfigListItem{
		ID:               cc.ID,
		Oid:              cc.Oid,
		Content:          cc.Content,
		URL:              cc.URL,
		Image:            cc.Image,
		ImageBig:         cc.ImageBig,
		HighlightContent: cc.HighlightContent,
		STime:            cc.STime.Time().Format("2006-01-02 15:04:05"),
		ETime:            cc.ETime.Time().Format("2006-01-02 15:04:05"),
		State:            cc.State,
		StateDesc:        asStateDesc(cc.ResolveStatusAt(now)),
		OriginType:       cc.OriginType,
		AuditCode:        cc.AuditCode,
	}
	if cc.TP == model.CustonConfigTPArchive {
		reply.BVID, _ = bvid.AvToBv(cc.Oid)
		func() {
			sa, err := s.dao.SimpleArchvie(ctx, cc.Oid)
			if err != nil {
				log.Error("Failed to get simple archvie: %d: %+v", cc.Oid, err)
				return
			}
			reply.Mid = sa.Mid
			reply.Title = sa.Title
			reply.OState = sa.State
		}()
		func() {
			ai, err := s.acc.Info3(ctx, reply.Mid)
			if err != nil {
				log.Error("Failed to get account info: %d: %+v", reply.Mid, err)
				return
			}
			reply.Name = ai.Name
		}()
	}
	return reply, nil
}
