package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_tableInformationRecommendCard = "information_recommend_card"
	_updateSQL                     = "UPDATE `information_recommend_card` SET `card_type`=?,`card_id`=?,`stime`=?,`etime`=?,`card_pos`=?,`is_cover`=?,`pos_index`=?,`apply_reason`=?,`cover_img`=?,audit_status=?,offline_status=? WHERE id = ?"
)

// RecommendCardByID
func (d *Dao) RecommendCardByID(c context.Context, id int64) (card *show.RecommendCard, err error) {
	card = &show.RecommendCard{}
	w := map[string]interface{}{
		"is_deleted": common.NotDeleted,
	}
	if err = d.DB.Table(_tableInformationRecommendCard).Where("id = ?", id).Where(w).First(&card).Error; err != nil {
		log.Error("dao.RecommendCardByID error(%v)", err)
	}
	return
}

// RecommendCardList
func (d *Dao) RecommendCardList(c context.Context, params *show.RecommendCardListReq) (list []*show.RecommendCard, count int, err error) {
	w := map[string]interface{}{
		"is_deleted": common.NotDeleted,
	}
	query := d.DB.Table(_tableInformationRecommendCard)
	if params.CardType > 0 {
		w["card_type"] = params.CardType
	}
	if params.AvID > 0 {
		w["card_id"] = params.AvID
	}
	cTimeStr := util.CTimeStr()
	// 状态筛选
	switch params.Status {
	case show.StatusToAudit:
		query = query.Where("audit_status = ? AND offline_status = ? AND etime >= ?", show.AuditStatusToAudit, show.OnlineStatus, cTimeStr)
	case show.StatusAuditPass:
		query = query.Where("audit_status = ? AND offline_status = ? AND stime >= ?", show.AuditStatusPass, show.OnlineStatus, cTimeStr)
	case show.StatusAuditReject:
		query = query.Where("audit_status = ? AND offline_status = ? AND etime >= ?", show.AuditStatusReject, show.OnlineStatus, cTimeStr)
	case show.StatusOffline:
		query = query.Where("offline_status = ? OR etime < ?", show.OfflineStatus, cTimeStr)
	case show.StatusOnline:
		query = query.Where("audit_status = ? AND offline_status = ? AND stime <= ? AND etime >= ?", show.AuditStatusPass, show.OnlineStatus, cTimeStr, cTimeStr)
	}

	if params.Uname != "" {
		query = query.Where("uname like ?", "%"+params.Uname+"%")
	}
	if params.Stime > 0 {
		query = query.Where("stime >= ?", params.Stime.Time().Format("2006-01-02 15:04:05"))
	}
	if params.Etime > 0 {
		query = query.Where("etime <= ?", params.Etime.Time().Format("2006-01-02 15:04:05"))
	}
	if err = query.Where(w).Count(&count).Error; err != nil {
		log.Error("showDao.RecommendCardList count error(%v)", err)
		return
	}
	if count == 0 {
		list = make([]*show.RecommendCard, 0)
		return
	}
	if err = query.Where(w).Order("`id` DESC").Offset((params.Pn - 1) * params.Ps).Limit(params.Ps).Find(&list).Error; err != nil {
		log.Error("showDao.RecommendCardList Find error(%v)", err)
		return
	}
	return
}

// AddRecommendCard
func (d *Dao) AddRecommendCard(c context.Context, params *show.RecommendCardAddReq) (err error) {
	if err = d.DB.Table(_tableInformationRecommendCard).Create(params).Error; err != nil {
		log.Error("showDao.AddRecommendCard Create error(%v), params(%v)", err, params)
	}
	return
}

// IntervalCheckRecommendCard
func (d *Dao) IntervalCheckRecommendCard(c context.Context, params *show.RecommendCardIntervalCheckReq) (overlap bool, err error) {
	var (
		count       int
		overlapItem = &show.RecommendCard{}
	)
	query := d.DB.Table(_tableInformationRecommendCard).
		Where("card_pos = ? AND pos_index = ? AND audit_status != ? AND offline_status != ? AND id != ? AND "+
			"NOT (stime > ? OR etime < ?)",
			params.CardPos, params.PosIndex, show.AuditStatusReject, show.OfflineStatus, params.ID,
			params.Etime.Time().Format("2006-01-02 15:04:05"), params.Stime.Time().Format("2006-01-02 15:04:05"))
	if err = query.Count(&count).Error; err != nil {
		log.Error("dao.DupCheckRecommendCard Count error(%v), params(%v)", err, params)
		return
	}
	if count == 0 {
		overlap = false
		return
	}
	overlap = true
	if err = query.First(&overlapItem).Error; err != nil {
		log.Error("dao.DupCheckRecommendCard First error(%v), params(%v)", err, params)
		return
	}
	log.Info("overlapItem: %+v", overlapItem)
	return
}

// ModifyRecommendCard
func (d *Dao) ModifyRecommendCard(c context.Context, params *show.RecommendCardModifyReq) (err error) {
	if err = d.DB.Table(_tableInformationRecommendCard).Exec(_updateSQL, params.CardType, params.CardID, params.Stime, params.Etime,
		params.CardPos, params.IsCover, params.PosIndex, params.ApplyReason, params.CoverImg, params.AuditStatus, params.OfflineStatus, params.ID).Error; err != nil {
		log.Error("showDao.ModifyRecommendCard Update error(%v), params(%v)", err, params)
	}
	return
}

// ModifyRecommendCard
func (d *Dao) DeleteRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	up := map[string]interface{}{
		"is_deleted": common.Deleted,
	}
	if err = d.DB.Table(_tableInformationRecommendCard).Where("id = ?", params.ID).Update(up).Error; err != nil {
		log.Error("showDao.DeleteRecommendCard Update error(%v), params(%v)", err, params)
	}
	return
}

// OfflineRecommendCard
func (d *Dao) OfflineRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	up := map[string]interface{}{
		"offline_status": show.OfflineStatus,
	}
	if err = d.DB.Table(_tableInformationRecommendCard).Where("id = ?", params.ID).Update(up).Error; err != nil {
		log.Error("showDao.OfflineRecommendCard Update error(%v), params(%v)", err, params)
	}
	return
}

// PassRecommendCard
func (d *Dao) PassRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	up := map[string]interface{}{
		"audit_status": show.AuditStatusPass,
	}
	if err = d.DB.Table(_tableInformationRecommendCard).Where("id = ?", params.ID).Update(up).Error; err != nil {
		log.Error("showDao.PassRecommendCard Update error(%v), params(%v)", err, params)
	}
	return
}

// RejectRecommendCard
func (d *Dao) RejectRecommendCard(c context.Context, params *show.RecommendCardOpReq) (err error) {
	up := map[string]interface{}{
		"audit_status": show.AuditStatusReject,
	}
	if err = d.DB.Table(_tableInformationRecommendCard).Where("id = ?", params.ID).Update(up).Error; err != nil {
		log.Error("showDao.RejectRecommendCard Update error(%v), params(%v)", err, params)
	}
	return
}
