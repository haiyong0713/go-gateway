package web

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"

	egV2 "go-common/library/sync/errgroup.v2"
)

const (
	// 业务负责人
	_roleAdmin1 = 1
	// 业务二级负责人
	_rRoleAdmin2 = 2
	// 审核人
	_roleAdmin = 3
	// 普通用户
	_roleOrdinary = 0
	// 异常用户
	_roleNoUser = -1
)

// WebRcmdRole .
func (s *Service) WebRcmdRole(name, auth string, permis []string) (res *common.Role, err error) {
	var (
		roleAdminsGroupOne, roleAdminsGroupTwo, roleUser []*manager.PosRecUserMgt
	)
	res = &common.Role{
		Role: _roleNoUser,
	}
	if s.isAdmin(auth, permis) {
		res.Role = _roleAdmin
		if res.Group, err = s.managerDao.UserGroupByPids([]int{}); err != nil {
			log.Error("WebRcmdRole managerDao.UserGroupByPids name(%s) req(%v) err(%v)", name, []int64{}, err)
			return
		}
		return
	}
	if roleAdminsGroupOne, err = s.managerDao.UserRole(name, common.RoleAdmin1); err != nil {
		log.Error("WebRcmdRole managerDao.UserRole UserRole(%s) level(%d) err(%v)", name, common.RoleAdmin1, err)
		return
	}
	if len(roleAdminsGroupOne) > 0 {
		// 一级负责人
		res.Role = _roleAdmin1
		res.RoleGroup = roleAdminsGroupOne
		// 负责人所属组
		ids := s.RoleIDs(roleAdminsGroupOne)
		if res.Group, err = s.managerDao.UserGroupByPids(ids); err != nil {
			log.Error("WebRcmdRole managerDao.UserGroupByPids name(%s) req(%v) err(%v)", name, ids, err)
		}
		return
	}
	if roleAdminsGroupTwo, err = s.managerDao.UserRole(name, common.RoleAdmin2); err != nil {
		log.Error("WebRcmdRole managerDao.UserRole name(%s) level(%d) err(%v)", name, common.RoleAdmin2, err)
		return
	}
	if len(roleAdminsGroupTwo) > 0 {
		// 二级负责人
		res.Role = _rRoleAdmin2
		res.RoleGroup = roleAdminsGroupTwo
		// 二级负责人所属组
		ids := s.RoleIDs(roleAdminsGroupTwo)
		if res.Group, err = s.managerDao.UserGroupByPids(ids); err != nil {
			log.Error("WebRcmdRole managerDao.UserGroupByPids name(%s) req(%v) err(%v)", name, ids, err)
		}
		return
	}
	if roleUser, err = s.managerDao.UserRole(name, common.RoleOrdinary); err != nil {
		log.Error("WebRcmdRole managerDao.UserRole name(%s) level(%d) err(%v)", name, common.RoleOrdinary, err)
		return
	}
	if len(roleUser) > 0 {
		// 普通用户
		res.Role = _roleOrdinary
		res.RoleGroup = roleUser
		// 普通用户所属组
		ids := s.RoleIDs(roleUser)
		if res.Group, err = s.managerDao.UserGroupByPids(ids); err != nil {
			log.Error("WebRcmdRole managerDao.UserGroupByPids name(%s) req(%v) err(%v)", name, ids, err)
		}
		return
	}
	return
}

// RoleIDs .
func (s *Service) RoleIDs(req []*manager.PosRecUserMgt) (res []int) {
	for _, v := range req {
		res = append(res, v.Pid)
	}
	return
}

func (s *Service) isAdmin(auth string, permis []string) bool {
	for _, v := range permis {
		if v == auth {
			return true
		}
	}
	return false
}

// WebRcmdCardList channel WebRcmdCard list
func (s *Service) WebRcmdCardList(lp *show.WebRcmdCardLP) (pager *show.WebRcmdCardPager, err error) {
	pager = &show.WebRcmdCardPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.showDao.DB.Model(&show.WebRcmdCard{})
	if lp.ID > 0 {
		w["id"] = lp.ID
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if lp.Title != "" {
		query = query.Where("title like ?", "%"+lp.Title+"%")
	}
	if lp.STime != "" {
		query = query.Where("ctime >= ?", lp.STime)
	}
	if lp.ETime != "" {
		query = query.Where("ctime <= ?", lp.ETime)
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("WebSvc.WebRcmdCardList count error(%v)", err)
		return
	}
	WebRcmdCards := make([]*show.WebRcmdCard, 0)
	if err = query.Where(w).Order("`id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&WebRcmdCards).Error; err != nil {
		log.Error("WebSvc.WebRcmdCardList Find error(%v)", err)
		return
	}
	pager.Item = WebRcmdCards
	return
}

// AddWebRcmdCard add channel WebRcmdCard
func (s *Service) AddWebRcmdCard(c context.Context, param *show.WebRcmdCardAP, name string, uid int64) (err error) {
	if err = s.showDao.WebRcmdCardAdd(param); err != nil {
		return
	}
	if err = util.AddWebRcmdCardLogs(name, uid, 0, common.ActionAdd, param); err != nil {
		log.Error("WebSvc.AddWebRcmdCard AddWebRcmdCardLogs error(%v)", err)
	}
	if err = util.AddLogs(common.LogWebRcmdCard, name, uid, 0, common.ActionAdd, param); err != nil {
		log.Error("WebSvc.AddWebRcmdCard AddLog error(%v)", err)
		return
	}
	return
}

// UpdateWebRcmdCard update channel WebRcmdCard
func (s *Service) UpdateWebRcmdCard(c context.Context, param *show.WebRcmdCardUP, name string, uid int64) (err error) {
	if err = s.showDao.WebRcmdCardUpdate(param); err != nil {
		return
	}
	if err = util.AddWebRcmdCardLogs(name, uid, param.ID, common.ActionEdit, param); err != nil {
		log.Error("WebSvc.UpdateWebRcmdCard AddWebRcmdCardLogs error(%v)", err)
	}
	if err = util.AddLogs(common.LogWebRcmdCard, name, uid, 0, common.ActionUpdate, param); err != nil {
		log.Error("WebSvc.UpdateWebRcmdCard AddLog error(%v)", err)
		return
	}
	return
}

// DeleteWebRcmdCard delete channel WebRcmdCard
func (s *Service) DeleteWebRcmdCard(id int64, name string, uid int64) (err error) {
	if err = s.showDao.WebRcmdCardDelete(id); err != nil {
		return
	}
	if err = util.AddWebRcmdCardLogs(name, uid, id, common.ActionDelete, id); err != nil {
		log.Error("WebSvc.DeleteWebRcmdCard AddWebRcmdCardLogs error(%v)", err)
	}
	if err = util.AddLogs(common.LogWebRcmdCard, name, uid, id, common.ActionDelete, id); err != nil {
		log.Error("WebSvc.DeleteWebRcmdCard AddLog error(%v)", err)
		return
	}
	return
}

// WebRcmdList Web list
func (s *Service) WebRcmdList(c context.Context, lp *show.WebRcmdLP) (pager *show.WebRcmdPager, err error) {
	pager = &show.WebRcmdPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.showDao.DB.Model(&show.WebRcmd{})
	if lp.ID != "" {
		w["card_value"] = lp.ID
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if lp.STime != "" {
		query = query.Where("stime >= ?", lp.STime)
	}
	if lp.ETime != "" {
		query = query.Where("etime <= ?", lp.ETime)
	}
	if lp.CardType != 0 {
		query = query.Where("card_type = ?", lp.CardType)
	}
	if lp.Partition != "" {
		query = query.Where("`partition` like ?", "%"+lp.Partition+"%")
	}
	if lp.Tag != "" {
		query = query.Where("tag like ?", "%"+lp.Tag+"%")
	}
	if lp.Avid != "" {
		query = query.Where("avid like ?", "%"+lp.Avid+"%")
	}
	if len(lp.GroupID) != 0 {
		query = query.Where("role_id in (?)", lp.GroupID)
	}
	cTimeStr := util.CTimeStr()
	if lp.Check != 0 {
		if lp.Check == common.Pass {
			// 已通过 未生效
			query = query.Where("`check` = ?", common.Pass)
			query = query.Where("stime > ?", cTimeStr)
		} else if lp.Check == common.Valid {
			// 已通过 已生效
			query = query.Where("`check` = ?", common.Pass)
			query = query.Where("stime <= ?", cTimeStr).Where("etime >= ?", cTimeStr)
		} else if lp.Check == common.InValid {
			// 已通过 已失效
			query = query.Where("(`check` = ? AND etime <= ?) OR (`check` = ?)", common.Pass, cTimeStr, common.InValid)
		} else {
			query = query.Where("`check` = ? ", lp.Check)
		}
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("WebSvc.WebList count error(%v)", err)
		return
	}
	Webs := make([]*show.WebRcmd, 0)
	if err = query.Where(w).Order("`id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&Webs).Error; err != nil {
		log.Error("WebSvc.WebList Find error(%v)", err)
		return
	}
	if len(Webs) > 0 {
		var (
			specialIDs []string
			gameIDs    []string
			avIDs      []string
			gIDs       []int
		)
		gIDMap := make(map[int]bool)
		for _, v := range Webs {
			if v.Check == common.Pass {
				c := time.Now().Unix()
				if (c >= v.Stime.Time().Unix()) && (c <= v.Etime.Time().Unix()) {
					v.Check = common.Valid
				} else if c > v.Etime.Time().Unix() && v.Check != common.InValid {
					v.Check = common.InValid
				}
			}
			if v.CardType == common.WebRcmdSpecial {
				// 特殊卡片 批量获取标题
				specialIDs = append(specialIDs, v.CardValue)
			} else if v.CardType == common.WebRcmdGame {
				// 游戏卡片 批量获取标题
				gameIDs = append(gameIDs, v.CardValue)
			} else if v.CardType == common.WebRcmdAV {
				avIDs = append(avIDs, v.CardValue)
				if v.BvID, err = bvav.AvStrToBvStr(v.CardValue); err != nil {
					v.BvID = err.Error()
					log.Error("WebSvc.WebList AvStrToBvStr(%s) error(%v)", v.CardValue, err)
					err = nil
					continue
				}
			}
			if _, ok := gIDMap[v.RoleId]; !ok {
				gIDs = append(gIDs, v.RoleId)
				gIDMap[v.RoleId] = true
			}
		}
		s.getCardInfo(c, Webs, specialIDs, gameIDs, avIDs, gIDs)
	}
	pager.Item = Webs
	return
}

//nolint:gocognit
func (s *Service) getCardInfo(c context.Context, webRcmds []*show.WebRcmd, specialIDs, gameIDs, avIDs []string, gIDs []int) {
	mapSpecial := make(map[string]*show.WebRcmdCard)
	Special := make([]*show.WebRcmdCard, 0)
	mapGame := make(map[string]*game.Game)
	mapArcs := make(map[string]*api.Arc)
	mapGroupInfo := make(map[int]*manager.PosRecUserMgt)
	eg := egV2.WithContext(c)
	mutex := sync.Mutex{}
	if len(specialIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			// 批量获取特殊卡片数据
			cardWhere := map[string]interface{}{
				"deleted": common.NotDeleted,
			}
			if err = s.showDao.DB.Model(&show.WebRcmdCard{}).Where(cardWhere).Where("id in (?)", specialIDs).Find(&Special).Error; err != nil {
				log.Error("showDao.DB Find(%v) error(%v)", specialIDs, err)
				err = nil
			}
			for _, v := range Special {
				mapSpecial[strconv.FormatInt(v.ID, 10)] = v
			}
			return
		})
	}
	if len(gameIDs) > 0 {
		for _, v := range gameIDs {
			gameID, _ := strconv.ParseInt(v, 10, 64)
			eg.Go(func(ctx context.Context) (err error) {
				var gameInfo *game.Game
				if gameInfo, err = s.GameDao.GamesPCInfo(ctx, gameID); err != nil || gameInfo == nil {
					log.Error("GamesPCInfo (%d) error(%v)", gameID, err)
					return nil
				}
				mutex.Lock()
				mapGame[strconv.FormatInt(gameInfo.ID, 10)] = gameInfo
				mutex.Unlock()
				return nil
			})
		}
	}
	if len(avIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			var avIDsTmp []int64

			//nolint:ineffassign,staticcheck
			arcs := make(map[int64]*api.Arc)
			for _, v := range avIDs {
				id, _ := strconv.ParseInt(v, 10, 64)
				avIDsTmp = append(avIDsTmp, id)
			}
			if arcs, err = s.arcDao.Arcs(ctx, avIDsTmp); err != nil {
				log.Error("WebSvc.arcDao.Arcs Find(%v) error(%v)", avIDsTmp, err)
				err = nil
			}
			for k, v := range arcs {
				mapArcs[strconv.FormatInt(k, 10)] = v
			}
			return
		})
	}
	if len(gIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			tmp, err := s.managerDao.UserGroupByPids(gIDs)
			if err != nil {
				log.Error("getCardInfo UserGroupByPids req(%v) err(%v)", gIDs, err)
				return nil
			}
			for _, v := range tmp {
				mapGroupInfo[v.ID] = v
			}
			return nil
		})
	}
	//nolint:errcheck
	eg.Wait()
	for _, web := range webRcmds {
		webCard := &show.WebRcmdCard{}
		var image string
		// 特殊卡片 批量获取标题
		if web.CardType == common.WebRcmdSpecial {
			if v, ok := mapSpecial[web.CardValue]; ok && v.Title != "" {
				webCard.Title = v.Title
				webCard.Image = v.Cover
				webCard.ReType = v.ReType
				webCard.ReValue = v.ReValue
				image = v.Image
			} else {
				webCard.Title = "！！！卡片信息获取失败！！！"
				webCard.Image = "！！！卡片信息获取失败！！！"
				image = "！！！卡片信息获取失败！！！"
			}
		} else if web.CardType == common.WebRcmdGame {
			if v, ok := mapGame[web.CardValue]; ok && v != nil {
				webCard.Title = v.Title
				webCard.Image = v.Image
				image = v.Image
			} else {
				webCard.Title = "！！！卡片信息获取失败！！！"
				webCard.Image = "！！！卡片信息获取失败！！！"
				image = "！！！卡片信息获取失败！！！"
			}
		} else if web.CardType == common.WebRcmdAV {
			if v, ok := mapArcs[web.CardValue]; ok && v.Title != "" {
				webCard.Title = v.Title
				webCard.Image = v.Pic
				image = v.Pic
			} else {
				webCard.Title = "！！！卡片信息获取失败！！！"
				webCard.Image = "！！！卡片信息获取失败！！！"
				image = "！！！卡片信息获取失败！！！"
			}
		}
		web.Card = webCard
		web.Image = image
		groupInfo, ok := mapGroupInfo[web.RoleId]
		if ok {
			web.GroupName = groupInfo.Name
		} else {
			web.GroupName = "-"
		}
	}
}

func (s *Service) validateWebRcmd(p *show.WebRcmdUP) (err error) {
	if p.Partition == "" && p.Tag == "" && p.Avid == "" {
		err = fmt.Errorf("你还没有关联内容")
		return
	}
	w := map[string]interface{}{
		"deleted":  common.NotDeleted,
		"priority": p.Priority,
	}
	query := s.showDao.DB.Model(&show.WebRcmd{}).Where(w).
		Where("stime <= ?", p.Etime).
		Where("etime >= ?", p.Stime).
		Where("`check` not in (?)", []int{common.Rejecte, common.InValid})
	if p.ID != 0 {
		query = query.Where("id != ?", p.ID)
	}
	rcmds := []*show.WebRcmd{}
	if err = query.Find(&rcmds).Error; err != nil {
		log.Error("WebSvc.validateWebRcmd Find param(%v) error(%v)", p, err)
		return
	}
	if len(rcmds) == 0 {
		return
	}
	var partionMap, tagMap, avidMap map[string]struct{}
	partionMap = make(map[string]struct{})
	tagMap = make(map[string]struct{})
	avidMap = make(map[string]struct{})
	for _, rcmd := range rcmds {
		partionsTmp := strings.Split(rcmd.Partition, ",")
		for _, v := range partionsTmp {
			partionMap[v] = struct{}{}
		}
		tagsTmp := strings.Split(rcmd.Tag, ",")
		for _, v := range tagsTmp {
			tagMap[v] = struct{}{}
		}
		avidsTmp := strings.Split(rcmd.Avid, ",")
		for _, v := range avidsTmp {
			avidMap[v] = struct{}{}
		}
	}
	if p.Avid != "" {
		avids := strings.Split(p.Avid, ",")
		for _, avid := range avids {
			if _, ok := avidMap[avid]; ok {
				err = fmt.Errorf("该稿件[%s]已经有配置", avid)
				return
			}
		}
	}
	if p.Partition != "" {
		partions := strings.Split(p.Partition, ",")
		for _, partion := range partions {
			if _, ok := partionMap[partion]; ok {
				err = fmt.Errorf("该分区[%s]已经有配置", partion)
				return
			}
		}
	}
	if p.Tag != "" {
		tags := strings.Split(p.Tag, ",")
		for _, tag := range tags {
			if _, ok := tagMap[tag]; ok {
				err = fmt.Errorf("该tag[%s]已经有配置", tag)
				return
			}
		}
	}
	return
}

// AddWebRcmd add Web recommand
func (s *Service) AddWebRcmd(c context.Context, param *show.WebRcmdAP, name string, uid int64) (err error) {
	userGroup, err := s.managerDao.UserGroupByName(name)
	if err != nil {
		return
	}
	if userGroup == nil {
		param.RoleId = 0
	} else {
		param.RoleId = userGroup.Pid
	}
	p := &show.WebRcmdUP{
		CardValue: param.CardValue,
		Partition: param.Partition,
		Tag:       param.Tag,
		Avid:      param.Avid,
		Priority:  param.Priority,
		Stime:     param.Stime,
		Etime:     param.Etime,
	}
	if err = s.validateWebRcmd(p); err != nil {
		return
	}
	if err = s.showDao.WebRcmdAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogWebRcmd, name, uid, 0, common.ActionAdd, param); err != nil {
		log.Error("WebSvc.AddWebRcmd AddLog error(%v)", err)
		return
	}
	return
}

// UpdateWebRcmd update Web
func (s *Service) UpdateWebRcmd(c context.Context, param *show.WebRcmdUP, name string, uid int64) (err error) {
	var (
		swValue *show.WebRcmd
	)
	p := &show.WebRcmdUP{
		CardValue: param.CardValue,
		Partition: param.Partition,
		Tag:       param.Tag,
		Avid:      param.Avid,
		Priority:  param.Priority,
		Stime:     param.Stime,
		Etime:     param.Etime,
		ID:        param.ID,
	}
	if err = s.validateWebRcmd(p); err != nil {
		return
	}
	if swValue, err = s.showDao.WebRcmdFindByID(param.ID); err != nil {
		log.Error("WebSvc.UpdateWebRcmd AddLog error(%v)", err)
		return
	}
	// 待审核&已通过&已生效-》编辑-》状态不变；其它-》编辑-》审待核
	cTime := time.Now().Unix()
	if (swValue.Check == common.Verify) ||
		(swValue.Check == common.Pass && swValue.Stime.Time().Unix() > cTime ||
			(swValue.Check == common.Pass && (cTime > swValue.Stime.Time().Unix() && cTime <= swValue.Stime.Time().Unix()))) {
		param.Check = swValue.Check
	} else {
		param.Check = common.Verify
	}
	if err = s.showDao.WebRcmdUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogWebRcmd, name, uid, 0, common.ActionUpdate, param); err != nil {
		log.Error("WebSvc.UpdateWebRcmd AddLog error(%v)", err)
		return
	}
	return
}

// DeleteWebRcmd delete Web
func (s *Service) DeleteWebRcmd(id int64, name string, uid int64) (err error) {
	if err = s.showDao.WebRcmdDelete(id); err != nil {
		return
	}
	if err = util.AddLogs(common.LogWebRcmd, name, uid, id, common.ActionDelete, id); err != nil {
		log.Error("WebSvc.DeleteWebRcmd AddLog error(%v)", err)
		return
	}
	return
}

// OptionWebRcmd option Web
func (s *Service) OptionWebRcmd(id int64, opt string, name string, uid int64, isBatchOpt int) (err error) {
	up := &show.WebRcmdOption{}
	if opt == common.OptionOnline {
		up.Check = common.Pass
	} else if opt == common.OptionHidden {
		up.Check = common.InValid
	} else if opt == common.OptionPass {
		up.Check = common.Pass
	} else if opt == common.OptionReject {
		up.Check = common.Rejecte
	} else {
		err = fmt.Errorf("参数不合法")
		return
	}
	up.ID = id
	if err = s.showDao.WebRcmdOption(up); err != nil {
		return
	}
	logParam := map[string]interface{}{
		"id":    id,
		"opt":   opt,
		"up":    up,
		"batch": isBatchOpt,
	}
	if err = util.AddLogs(common.LogWebRcmd, name, uid, id, common.ActionOpt, logParam); err != nil {
		log.Error("WebSvc.OptionWebRcmd AddLog error(%v)", err)
		return
	}
	return
}

type BatchOptionWebRcmdResItem struct {
	OK     int    `json:"ok"`
	ID     int64  `json:"id"`
	Reason string `json:"reason"`
}

// 批量操作
func (s *Service) BatchOptionWebRcmd(ids []int64, opt string, name string, uid int64) (res []*BatchOptionWebRcmdResItem, err error) {
	if opt == common.OptionBatchPass {
		for _, id := range ids {
			e := s.OptionWebRcmd(id, common.OptionPass, name, uid, 1)
			if e != nil {
				res = append(res, &BatchOptionWebRcmdResItem{
					OK:     0,
					ID:     id,
					Reason: err.Error(),
				})
			} else {
				res = append(res, &BatchOptionWebRcmdResItem{
					OK:     1,
					ID:     id,
					Reason: "操作成功",
				})
			}
		}
	} else if opt == common.OptionBatchReject {
		for _, id := range ids {
			e := s.OptionWebRcmd(id, common.OptionReject, name, uid, 1)
			if e != nil {
				res = append(res, &BatchOptionWebRcmdResItem{
					OK:     0,
					ID:     id,
					Reason: e.Error(),
				})
			} else {
				res = append(res, &BatchOptionWebRcmdResItem{
					OK:     1,
					ID:     id,
					Reason: "操作成功",
				})
			}
		}
	}
	return
}
