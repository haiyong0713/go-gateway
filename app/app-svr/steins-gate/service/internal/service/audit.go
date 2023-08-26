package service

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func (s *Service) GraphAudit(c context.Context, param *model.AuditParam) (err error) {
	var (
		graph        *model.GraphAuditDB
		arc          *model.VideoUpView
		newResultGID int64
	)
	if graph, err = s.auditDao.GraphAuditByID(c, param.ID); err != nil {
		log.Error("%v", err)
		return
	}
	if graph == nil {
		err = ecode.NothingFound
		log.Error("GraphAudit ByID %d not found", param.ID)
		return
	}
	if param.WithNotify == 1 || param.State == model.GraphStatePass { // 需要通知up主 或者是 通过操作需要搬运到结果表我们需要根据mid进行灰度
		if arc, err = s.arcDao.VideoUpView(c, graph.Aid); err != nil {
			log.Error("%v", err)
			return
		}
	}
	if graph.ResultGID > 0 { // 已有结果，更新审核表+结果表
		if err = s.auditDao.GraphAudit(c, graph.Id, graph.ResultGID, param.State); err != nil {
			log.Error("%v", err)
			return
		}
		if param.State == model.GraphStateRepulse { // 结果表被驳回，删缓存
			if err = s.dao.DelGraphCache(c, graph.Aid); err != nil {
				log.Error("%v", err)
				return
			}
		}
	} else { // 从未出现在结果表
		if param.State == model.GraphStatePass { // 审核通过，graph数据从审核表迁移到结果表
			if newResultGID, err = s.dao.GraphAuditMigrate(c, graph, arc.Archive.Mid); err != nil {
				log.Error("%v", err)
				return
			}
			if err = s.auditDao.GraphAuditPass(c, newResultGID, graph.Id); err != nil { // 更新审核表状态和resultGID字段
				log.Error("%v", err)
				return
			}
		} else { // 审核拒绝
			if err = s.auditDao.GraphAuditRepulse(c, graph.Id); err != nil { // 更新审核表状态
				log.Error("%v", err)
				return
			}
		}
	}
	if param.WithNotify == 1 && (param.State == model.GraphStatePass || param.State == model.GraphStateRepulse) {
		if err = s.notifyUser(c, param, arc); err != nil {
			log.Error("%v", err)
			return
		}
	}
	return
}

func (s *Service) notifyUser(c context.Context, param *model.AuditParam, arc *model.VideoUpView) (err error) {
	var title, content string
	if param.State == model.GraphStatePass {
		title = s.c.Rule.PrivateMsg.PassTitle
		content = fmt.Sprintf(s.c.Rule.PrivateMsg.PassContent, arc.Archive.Title)
	} else { // param.State == model.GraphStateRepulse
		title = s.c.Rule.PrivateMsg.RejectTitle
		content = fmt.Sprintf(s.c.Rule.PrivateMsg.RejectContent, arc.Archive.Title, param.BuildMsg())
	}
	log.Warn("AuditPrivateMsg Gid %d Aid %d, State %d, Title %s, Content %s", param.ID, arc.Archive.Aid, param.State, title, content)
	if err = s.dao.SendMessage(c, []int64{arc.Archive.Mid}, s.c.Rule.PrivateMsg.MC, title, content); err != nil {
		log.Error("paramGid %d Err %v", param.ID, err)
		return
	}
	return

}
