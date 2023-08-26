package dao

import (
	"context"
	"github.com/pkg/errors"
	"net/url"
	"strconv"
	"time"

	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	controlGRPC "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	moralrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"
	sysMsgGRPC "git.bilibili.co/bapis/bapis-go/system-msg/interface"
	"github.com/jinzhu/gorm"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/actionlog"

	"go-gateway/app/web-svr/space/admin/model"
)

const (
	_fansURL      = "/x/admin/space/fans"
	_actionLog    = "/x/admin/search/log"
	_vipInfo      = "/internal/v1/user"
	_bfsMove      = "/x/internal/upload/file/move"
	_notifySend   = "/api/notify/send.user.notify.do"
	_accountBlock = "/x/internal/block/block"
	_creditBlock  = "/x/internal/credit/blocked/info/add"
	_delMoral     = "/x/internal/member/moral/update"
	_purgeCache   = "/api/member/purgeCache"
	//
	_clearCacheTopPhoto = "/x/internal/space/topphoto/cache/clear"
)

/*============================================头图审核==============================================*/

// GetMemberUploadTopPhotoByPage 分页获取头图审核信息
func (d *Dao) GetMemberUploadTopPhotoByPage(params *model.MemberUploadTopPhotoSearchParams, pn int, ps int) (list []*model.MemberUploadTopPhoto, total int, err error) {
	list = make([]*model.MemberUploadTopPhoto, 0)
	query := d.DB.Model(&model.MemberUploadTopPhoto{}).Where("deleted = ? AND status = ?", model.MEMBER_UPLOAD_TOP_PHOTO_NOTDELETED, model.TOP_PHOTO_UNPASS)

	if params.UploadTimeStart != "" {
		query = query.Where("upload_date >= ?", params.UploadTimeStart)
	}

	if params.UploadTimeEnd != "" {
		query = query.Where("upload_date <= ?", params.UploadTimeEnd)
	}

	if len(params.MIDs) != 0 {
		query = query.Where("mid IN (?) ", params.MIDs)
	}

	if params.PlatFrom != 0 {
		query = query.Where("platfrom = ?", params.PlatFrom)
	}

	if err = query.Order("id ASC").Count(&total).Limit(ps).Offset((pn - 1) * ps).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return
}

// GetPhotoByIDs 根据ID返回
func (d *Dao) GetPhotoByIDs(ids []int64) (res []*model.MemberUploadTopPhoto, err error) {
	query := d.DB.Model(&model.MemberUploadTopPhoto{}).Where("id IN (?)", ids)
	if err = query.Find(&res).Error; err != nil && err != xecode.NothingFound {
		return nil, errors.Wrapf(err, "TopPhoto Dao: GetPhotoByIDs  (%+v)", ids)
	}
	return res, nil

}

// GetLastPhoto 获取上一张通过的头图
func (d *Dao) GetLastPhoto(mid int64, platfrom []int, id int64) (res *model.MemberUploadTopPhoto, err error) {
	if mid == 0 {
		return
	}

	res = &model.MemberUploadTopPhoto{}
	query := d.DB.Model(&model.MemberUploadTopPhoto{}).
		Order("modify_time DESC").
		Where("mid = ? AND status = ? AND platfrom IN (?) AND id != ?", mid, model.TOP_PHOTO_PASSED, platfrom, id)
	if err = query.First(res).Error; err != nil && err != xecode.NothingFound {
		return nil, errors.Wrapf(err, "Dao: GetLastPhoto (%d, %+v, %d)", mid, platfrom, id)
	}

	return res, nil

}

func (d *Dao) RePassPhotoEditByID(id int64, path string, tx *gorm.DB) (err error) {
	if id == 0 {
		return xecode.RequestErr
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	query := tx.Model(&model.MemberUploadTopPhoto{}).Where("id = ?", id)
	up := map[string]interface{}{
		"status":   model.TOP_PHOTO_PASSED,
		"img_path": path,
	}
	return query.Update(up).Update("deleted", model.MEMBER_UPLOAD_TOP_PHOTO_NOTDELETED).Error

}

func (d *Dao) EditPhotoByID(id int64, up map[string]interface{}, tx *gorm.DB) (err error) {
	if id == 0 {
		return xecode.RequestErr
	}
	if len(up) == 0 {
		return
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	query := tx.Model(&model.MemberUploadTopPhoto{}).Where("id = ?", id)
	return query.Update(up).Error
}

// GetNowValidPhoto	获取当前生效的头图记录
func (d *Dao) GetNowValidPhoto(mid int64, platfrom int, id int64) (res *model.MemberUploadTopPhoto, err error) {
	if mid == 0 {
		return
	}

	var condi = map[string]interface{}{
		"mid":      mid,
		"platfrom": platfrom,
		"status":   model.TOP_PHOTO_PASSED,
		"deleted":  model.MEMBER_UPLOAD_TOP_PHOTO_NOTDELETED,
	}
	res = &model.MemberUploadTopPhoto{}
	query := d.DB.Model(&model.MemberUploadTopPhoto{}).
		Order("upload_date DESC").
		Where(condi).
		Where("id != ?", id)

	if err = query.First(res).Error; err != nil {
		if err == xecode.NothingFound {
			err = nil
		} else {
			err = errors.Wrapf(err, "Dao: GetNowValidPhoto (%d, %d, %d)", mid, platfrom, id)
		}
		return nil, err
	}

	return res, nil

}

// GetBackTimes 获取驳回次数
func (d *Dao) GetBackTimes(mids []int64) (backTimes map[int64]int, err error) {
	if len(mids) == 0 {
		return nil, nil
	}

	res := make([]model.TopPhotoBackTimes, 0)
	backTimes = make(map[int64]int)
	query := d.DB.Model(&model.MemberUploadTopPhoto{}).Select("mid, count(1) as back_times").
		Where("status = ? AND mid IN (?)", model.TOP_PHOTO_NOTPASSED, mids).Group("mid")
	if err = query.Scan(&res).Error; err != nil {
		return nil, errors.Wrapf(err, "Dao: GetBackTimes (%+v)", mids)
	}

	for _, r := range res {
		backTimes[r.Mid] = r.BackTimes
	}

	return
}

// GetFans 获取粉丝数
func (d *Dao) GetFans(mids []int64) (fans map[int64]int64, err error) {
	if len(mids) == 0 {
		return nil, nil
	}

	fans = make(map[int64]int64)
	var fansReply *relationGRPC.StatsReply
	var req = &relationGRPC.MidsReq{Mids: mids}

	if fansReply, err = d.relationClient.Stats(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: GetFans grpc Stats Error %+v", err)
		return nil, err
	}

	for _, mid := range mids {
		fans[mid] = fansReply.StatReplyMap[mid].Follower
	}

	return
}

// GetAccountInfosByMIDs 根据MID查询账号信息
func (d *Dao) GetAccountInfosByMIDs(mids []int64) (reply map[int64]*midrpc.Info, err error) {
	if len(mids) == 0 {
		return nil, nil
	}
	var (
		infosReply *midrpc.InfosReply
		req        = &midrpc.MidsReq{Mids: mids}
	)

	if infosReply, err = d.midClient.Infos3(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: GetAccountInfosByMIDs GRPC Infos3 Error %+v", err)
		return nil, err
	}

	reply = infosReply.Infos

	return
}

// GetAccountCertification 根据Mid查询账号认证信息
func (d *Dao) GetAccountCertification(mids []int64) (reply map[int64]*midrpc.Card, err error) {
	if len(mids) == 0 {
		return nil, nil
	}

	var (
		cardsReply *midrpc.CardsReply
		req        = &midrpc.MidsReq{Mids: mids}
	)

	if cardsReply, err = d.midClient.Cards3(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: GetAccountCertification GRPC Cards3 Error %+v", err)
		return nil, err
	}

	reply = cardsReply.Cards

	return

}

// 获取 vip相关信息
func (d *Dao) GetVipInfo(mid int64) (info *midrpc.VipInfo, err error) {
	var (
		req       = &midrpc.MidReq{Mid: mid}
		cardReply *midrpc.CardReply
	)

	if cardReply, err = d.midClient.Card3(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: GetVipInfo grpc Card3 Error %+v", err)
		return nil, err
	}

	info = &cardReply.Card.Vip

	return

}

/*==================================================================================================*/
/*============================================行为日志==============================================*/

// GetActionLog 获取行为日志
func (d *Dao) GetActionLog(ids string, pn int, ps int) (data *model.LogSearchResRawData, err error) {
	query := url.Values{}
	query.Set("appid", "log_audit")
	query.Set("business", "241")
	query.Set("order", "ctime")
	query.Set("oid", ids)
	query.Set("pn", strconv.FormatInt(int64(pn), 10))
	query.Set("ps", strconv.FormatInt(int64(ps), 10))
	res := &model.LogSearchResRaw{}
	if err = d.http.Get(context.Background(), d.actionLogURL, "", query, res); err != nil {
		log.Error("TopPhoto Dao: GetActionLog HTTPGet Error %+v", err)
		return nil, err
	}
	data = res.Data
	return
}

// AddAuditLog 添加行为日志
func (d *Dao) AddAuditLog(params *model.AuditLogInitParams) (err error) {
	mInfo := &actionlog.ManagerInfo{
		Business: params.Business,                                // 业务 id, 请填写 info 中对应的业务 id
		Uname:    params.UName,                                   //全匹配, 默认:审核人员内网name, 业务方可自定义
		UID:      params.UID,                                     //全匹配, 默认:审核人员内网uid, 业务方可自定义
		Type:     params.Type,                                    //全匹配, 默认:操作对象的类型, 业务方可自定义
		Oid:      params.OID,                                     //全匹配, 默认:操作对象的id, 业务方可自定义
		Action:   params.Action,                                  //全匹配, 默认:具体操作类型，如打回, 业务方可自定义
		Ctime:    params.CTime,                                   // 可以时间排序
		Index:    params.Index,                                   // 为预留自定义字段, 根据传入的数据格式 string转化为 str_0~str_9(全匹配),
		Content:  map[string]interface{}{"json": params.Content}, // 数据只展示, 不参与搜索, 在 es 中保存为一个 json 字符串
	}

	// 异步请求, batchSize 条数(默认:10)据或每隔 workInterval 时间(默认:1s)上传一次数据,
	// 应用奔溃,可能造成数据丢失
	// 上传三次数据未成功, 会丢弃数据, 造成数据的丢失
	err = actionlog.AsyncManager(mInfo)
	return
}

/*==================================================================================================*/
/*============================================用户头图==============================================*/

func (d *Dao) DeleteByParams(mid int64, platfrom int, tx *gorm.DB) (err error) {
	if mid == 0 {
		return xecode.Error(xecode.RequestErr, "DeleteByParams参数为空")
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	query := tx.Model(&model.MemberTopPhoto{MID: mid})
	if err = query.Delete(model.MemberTopPhoto{}, "mid = ? AND platfrom = ?", mid, platfrom).Error; err != nil {
		log.Error("TopPhoto Dao: DeleteByParams (%d) Error %+v", mid, err)
		return err
	}
	return

}

func (d *Dao) EditByMidToNotActivated(mid int64, tx *gorm.DB) (err error) {
	if mid == 0 {
		return xecode.Error(xecode.RequestErr, "EditByMidToNotActivated参数为空")
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	query := tx.Model(&model.MemberTopPhoto{MID: mid}).Where("mid = ?", mid)
	if err = query.Update("is_activated", 0).Error; err != nil {
		log.Error("TopPhoto Dao: EditByMid (%d) Error %+v", mid, err)
		return err
	}

	return

}

func (d *Dao) AddTopPhoto(toAdd *model.MemberTopPhoto, tx *gorm.DB) (err error) {
	if toAdd == nil {
		return xecode.Error(xecode.RequestErr, "AddTopPhoto参数为空")
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	query := tx.Model(&model.MemberTopPhoto{MID: toAdd.MID})
	if err = query.Create(toAdd).Error; err != nil {
		log.Error("Dao AddTopPhoto Error %+v", err)
		return err
	}

	return
}

func (d *Dao) GetByMidAndSid(mid int64, sid int64) (mtp *model.MemberTopPhoto, err error) {
	mtp = &model.MemberTopPhoto{}
	query := d.DB.Model(&model.MemberTopPhoto{MID: mid}).Where("mid = ? AND sid = ?", mid, sid)
	if err = query.Find(mtp).Error; err != nil {
		if err == xecode.NothingFound {
			err = nil
		} else {
			log.Error("TopPhoto Dao: GetByParams (%d, %d) Error %+v", mid, sid, err)
			return nil, err
		}
	}
	return
}

func (d *Dao) EditByParams(mid int64, updateMap map[string]interface{}, condition map[string]interface{}, tx *gorm.DB) (err error) {
	if mid == 0 {
		return xecode.Error(xecode.RequestErr, "EditByParams参数为空")
	}
	if len(updateMap) == 0 {
		return
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	query := tx.Model(&model.MemberTopPhoto{MID: mid}).Where(condition)
	if err = query.Update(updateMap).Error; err != nil {
		return errors.Wrapf(err, "Dao: EditByParams (%d, %+v, %+v)", mid, updateMap, condition)
	}
	return
}

/*==================================================================================================*/
/*============================================bfs==============================================*/

// BFSMove 图片移动
func (d *Dao) BFSMove(imgPath string, uname string, bucket string) (res *model.BfsRes, err error) {
	if imgPath == "" || bucket == "" {
		return
	}

	var (
		query = url.Values{}
	)
	query.Set("urls", imgPath)
	query.Set("username", uname)
	query.Set("bucket", bucket)

	res = &model.BfsRes{}

	if err = d.http.Post(context.Background(), d.bfsMoveURL, "", query, res); err != nil {
		return nil, errors.Wrapf(err, "Dao: BFSMove (%s, %s)", imgPath, bucket)
	}

	return

}

// PurgeCache 删除缓存
func (d *Dao) PurgeCache(mid int64) (res *model.ResRaw, err error) {
	query := url.Values{}

	query.Set("mid", strconv.FormatInt(mid, 10))
	query.Set("modifiedAttr", "purgePhotoCache")

	res = &model.ResRaw{}

	if err = d.http.Get(context.Background(), d.purgeCacheURL, "", query, res); err != nil {
		return nil, errors.Wrapf(err, "Dao: PurgeCache (%d) http.Get", mid)
	}

	return
}

/*==================================================================================================*/
/*============================================系统通知==============================================*/

func (d *Dao) SendNotify(param *model.NotifySendInit) (reply *sysMsgGRPC.AsyncSendUserNotifyResp, err error) {
	var (
		req = &sysMsgGRPC.SendUserNotifyReq{
			Mc:       param.Mc,
			Title:    param.Title,
			DataType: int32(param.DataType),
			Context:  param.Context,
			MidList:  []uint64{uint64(param.MIDList)},
		}
	)

	if reply, err = d.systemMsgClient.AsyncSendUserNotify(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: SendNotify grpc AsyncSendUserNotify Error %+v", err)
		return nil, err
	}

	return reply, nil
}

/*==================================================================================================*/
/*============================================账户封禁==============================================*/

func (d *Dao) BlockAccount(param *model.AccountBlockInit) (err error) {
	if param.Comment == "" {
		param.Comment = param.Reason
	}

	var (
		expire = time.Now().Unix() + int64(param.Duration)
		req    = &controlGRPC.AddControlRoleReq{
			Mid:          []int64{param.MID},
			ControlRole:  []string{"silence"},
			Business:     "space_banner",
			OperatorId:   param.OpID,
			OperatorName: param.Operator,
			Comment:      param.Comment,
			Reason:       param.Reason,
			IsNotify:     param.Notify == 1,
			IsExpirable:  param.Action == 1,
			ExpireAt:     expire,
		}
	)

	if _, err = d.controlClient.AddControlRole(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: BlockAccount grpc AddControlRole Error %+v", err)
		return err
	}

	return
}

func (d *Dao) CreditInfoAdd(param *model.BlockInfoAdd) (res *model.ResRaw, err error) {
	query := url.Values{}
	query.Set("mid", strconv.FormatInt(param.MID, 10))
	query.Set("blocked_days", strconv.Itoa(param.BlockedDays))
	query.Set("blocked_forever", strconv.Itoa(param.BlockedForever))
	query.Set("blocked_remark", param.BlockedRemark)
	query.Set("moral_num", strconv.Itoa(param.MoralNum))
	query.Set("origin_type", strconv.Itoa(param.OriginType))
	query.Set("punish_time", strconv.FormatInt(param.PunishTime, 10))
	query.Set("punish_type", strconv.Itoa(param.PunishType))
	query.Set("reason_type", strconv.Itoa(param.ReasonType))
	query.Set("oper_id", strconv.FormatInt(param.OperID, 10))
	query.Set("operator_name", param.OperatorName)

	res = &model.ResRaw{}

	if err = d.http.Post(context.Background(), d.creditBlockInfoURL, "", query, res); err != nil {
		log.Error("TopPhoto Dao: CreditInfoAdd HTTPPost Error %+v", err)
		return nil, err
	}

	return
}

func (d *Dao) DelMoral(param *model.DelMoralParam) (err error) {
	var (
		req = &moralrpc.UpdateMoralReq{
			Mid:        param.MID,
			Delta:      int64(param.Delta),
			Origin:     int64(param.Origin),
			Reason:     param.Reason,
			ReasonType: int64(param.ReasonType),
			Operator:   param.Operator,
			Remark:     param.Remark,
			IsNotify:   param.IsNotify == 1,
		}
	)

	if _, err = d.moralClient.AddMoral(context.Background(), req); err != nil {
		log.Error("TopPhoto Dao: DelMoral grpc AddMoral Error %+v", err)
		return err
	}

	return
}

// 清除缓存
func (d *Dao) ClearCacheTopPhoto(c context.Context, mid int64) error {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	if err := d.http.Post(c, d.clearCacheTopPhotoURL, "", params, &res); err != nil {
		return errors.Wrapf(err, "mid:%d", mid)
	}
	if res.Code != 0 {
		return errors.Wrap(xecode.Int(res.Code), d.clearCacheTopPhotoURL+"?"+params.Encode())
	}
	return nil
}
