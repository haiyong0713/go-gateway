package like

import (
	"context"
	xsql "database/sql"
	"fmt"
	audit "git.bilibili.co/bapis/bapis-go/aegis/strategy/service"
	fliapi "git.bilibili.co/bapis/bapis-go/filter/service"
	tunnelCommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"
	tunnelEcode "git.bilibili.co/bapis/bapis-go/platform/common/tunnel/ecode"
	tunnel "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
	videoup "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/tools/lib/function"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-gateway/app/web-svr/activity/interface/api"
)

const (
	_upActReserveRelationByOid         = "select `id`, `sid`, `mid`, `state`, `oid`, `type`, `audit`, `audit_channel` from up_act_reserve_relation where oid = ? and type = ? and state = ?"
	_upActReserveUpdateAuditAndChannel = "update up_act_reserve_relation set audit = ?,audit_channel = ?,state = ? where sid = ?"
	_queryMidByDate                    = "select sid, mid from up_act_reserve_relation where state = 100 and type = 1 and ctime > '%s' and ctime < '%s' limit %s offset %s"
	_upActReserveRelationBySid         = "select `id`, `sid`, `mid`, `type`, `state`, `live_plan_start_time`, `oid`, `audit`, `audit_channel`, `dynamic_id`, `lottery_type` from up_act_reserve_relation where sid = ?"
)

func (d *Dao) GetUpActReserveWaitingArcAudit(ctx context.Context, oid string) (res *like.UpActReserveRelation, err error) {
	row := d.db.QueryRow(ctx, _upActReserveRelationByOid, oid, int64(api.UpActReserveRelationType_Archive), int64(api.UpActReserveRelationState_UpReserveRelatedOnline))
	res = new(like.UpActReserveRelation)
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.State, &res.Oid, &res.Type, &res.Audit, &res.AuditChannel); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = fmt.Errorf("GetUpActReserveWaitingArcAudit d.db.Query _upActReserveRelationByOid rows.Scan error(%+v)", err)
		}
	}

	return
}

func (d *Dao) GetUpActReserveRelationInfoBySid(ctx context.Context, sid int64) (res *like.UpActReserveRelation, err error) {
	row := d.db.QueryRow(ctx, _upActReserveRelationBySid, sid)
	res = new(like.UpActReserveRelation)
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.Type, &res.State, &res.LivePlanStartTime, &res.Oid, &res.Audit, &res.AuditChannel, &res.DynamicID, &res.LotteryType); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = fmt.Errorf("GetUpActReserveRelationInfoBySid d.db.Query _upActReserveRelationBySid rows.Scan error(%+v)", err)
		}
	}

	return
}

func (d *Dao) UpActReserveUnionChangeState(ctx context.Context, sid int64, subjectState int64, audit int64, auditChannel int64, relationState int64) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(ctx, "recover() tx.Rollback() err(%+v)", err)
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(ctx, "tx.Rollback() error(%v)", err)
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "tx.Commit() error(%v)", err)
			return
		}
		return
	}()

	// act_subject
	if err = d.TXUpdateActSubjectState(tx, subjectState, sid); err != nil {
		return
	}
	// up_act_reserve_relation
	if err = d.TXUpdateUpActReserveAudit(tx, audit, auditChannel, relationState, sid); err != nil {
		return
	}

	return
}

func (d *Dao) TXUpdateUpActReserveAudit(tx *sql.Tx, audit int64, auditChannel int64, relationState int64, sid int64) (err error) {
	if _, err = tx.Exec(_upActReserveUpdateAuditAndChannel, audit, auditChannel, relationState, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

func (d *Dao) GetBFSFileInfo(ctx context.Context, file string) (res *like.BFSFileInfo, err error) {
	res = &like.BFSFileInfo{}
	err = retry.WithAttempts(ctx, "GetBFSFileInfo", like.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		err = d.httpClientBFS.Get(ctx, file, "", url.Values{}, res)
		log.Infoc(ctx, "GetBFSFileInfo file(%+v) reply(%+v)", file, res)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "GetBFSFileInfo failed file(%+v) reply(%+v) error(%+v)", file, res, err)
		return
	}
	if res.Width == 0 || res.Height == 0 || res.FileSize == 0 {
		log.Errorc(ctx, "GetBFSFileInfo params illegal res(%+v)", res)
		return
	}
	return
}

func (d *Dao) GetArchiveInfo(ctx context.Context, aid int64) (res *videoup.ArchiveSimpleReply, err error) {
	res = new(videoup.ArchiveSimpleReply)
	err = retry.WithAttempts(ctx, "GetArchiveInfo", like.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		res, err = client.VideoClient.ArchiveSimple(ctx, &videoup.ArchiveSimpleReq{
			Aid: aid,
		})
		log.Infoc(ctx, "GetArchiveInfo client.VideoClient.ArchiveSimple aid(%+v) reply(%+v)", aid, res)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "GetArchiveInfo client.VideoClient.ArchiveSimple failed err(%v) aid(%+v)", err, aid)
		return
	}
	if res == nil || res.Arc == nil || res.Arc.Aid == 0 {
		log.Errorc(ctx, "GetArchiveInfo client.VideoClient.ArchiveSimple s.arcs nil reply(%+v)", res)
		return
	}
	return
}

func (d *Dao) FilterTitle(ctx context.Context, mid int64, content string) (level int64, err error, reply *fliapi.FilterV5Reply) {
	reply = &fliapi.FilterV5Reply{}
	req := &fliapi.FilterReq{
		Area:    "bullet_up",
		Mid:     mid,
		Message: content,
	}
	err = retry.WithAttempts(ctx, "FilterTitle", like.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		reply, err = client.FilterClient.FilterV5(ctx, req)
		log.Infoc(ctx, "FilterTitle client.FilterClient.FilterV5 req(%+v) reply(%+v)", req, reply)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "FilterTitle client.FilterClient.FilterV5 err(%+v) req(%+v) reply(%+v)", err, req, reply)
		return
	}

	// 区分社区 不送审 4 和 5 Source: 4, Desc: "运营规避"  Source: 5, Desc: "OGV运营"
	sourceLevels := []int64{0}
	if len(reply.Rules) > 0 {
		for _, v := range reply.Rules {
			if v.Source == 4 || v.Source == 5 {
				continue
			}
			if v.Level >= 0 {
				sourceLevels = append(sourceLevels, int64(v.Level))
			}
		}
	}

	// 不通source优先级整合
	if function.InInt64Slice(int64(like.SensitiveLevelIntercept20), sourceLevels) { // 标记打回
		level = like.SensitiveLevelIntercept20
	} else if function.InInt64Slice(int64(like.SensitiveLevelIntercept30), sourceLevels) { // 标记打回
		level = like.SensitiveLevelIntercept30
	} else if function.InInt64Slice(int64(like.SensitiveLevelIntercept40), sourceLevels) { // 标记打回
		level = like.SensitiveLevelIntercept40
	} else if function.InInt64Slice(int64(like.SensitiveLevelAudit), sourceLevels) { // 先审后发
		level = like.SensitiveLevelAudit
	} else if function.InInt64Slice(int64(like.SensitiveLevelPass), sourceLevels) { // 先发后审
		level = like.SensitiveLevelPass
	} else if function.InInt64Slice(int64(like.SensitiveLevelTest), sourceLevels) { // 测试标记
		level = like.SensitiveLevelTest
	} else if function.InInt64Slice(int64(like.SensitiveLevelNormal), sourceLevels) { // 通过
		level = like.SensitiveLevelNormal
	} else {
		err = fmt.Errorf("unknow type")
		log.Errorc(ctx, "FilterTitle err(%+v) req(%+v) reply(%+v)", err, req, reply)
	}

	return
}

func (d *Dao) Go2Audit(ctx context.Context, sid int64, mid int64, title string, filter *fliapi.FilterV5Reply) (err error) {
	// 审核服务历史遗留问题 需要转换rpc结构体
	positions := make([]*audit.PosProto, 0)
	rules := make([]*audit.RuleProto, 0)
	if len(filter.Positions) > 0 {
		for _, v := range filter.Positions {
			positions = append(positions, &audit.PosProto{From: int64(v.From), To: int64(v.To)})
		}
	}
	if len(filter.Rules) > 0 {
		for _, v := range filter.Rules {
			rules = append(rules, &audit.RuleProto{
				Id:       v.ID,
				Mode:     int32(v.Mode),
				Rule:     v.Rule,
				Area:     v.Area,
				Key:      v.Key,
				Level:    int32(v.Level),
				Stime:    v.STime,
				Etime:    v.ETime,
				Comment:  v.Comment,
				Cid:      v.CID,
				Type:     v.Type,
				Source:   v.Source,
				RuleType: v.RuleType,
			})
		}
	}

	req := &audit.AegisProcessReq{
		BusinessId: d.c.UpActReserveAudit.BizID1,
		AddInfo: &audit.AegisAddInfo{
			BusinessId: d.c.UpActReserveAudit.BizID2,
			NetId:      d.c.UpActReserveAudit.NetID,
			Oid:        strconv.FormatInt(sid, 10),
			Mid:        mid,
			Content:    title,
			Filter: &audit.FilterReply{
				Result:    filter.Result,
				Level:     filter.Level,
				Positions: positions,
				Rules:     rules,
				Reason:    filter.Reason,
			},
		},
	}

	err = retry.WithAttempts(ctx, "Go2Audit", like.UpActReserveRelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = client.AuditClient.AegisProcess(ctx, req)
		log.Infoc(ctx, "client.AuditClient.AegisProcess req(%+v)", req)
		return err
	})

	if err != nil {
		log.Errorc(ctx, "client.AuditClient.AegisProcess err(%+v) req(%+v)", err, req)
		return
	}

	return
}

func (d *Dao) BuildPushCard(ctx context.Context, sid, uniqueId int64) (err error) {
	sTime := function.Now()
	eTime := sTime + 60*60*72

	var description string
	if uniqueId == like.NotifyMessageTypePushUpVerify14 {
		description = "提醒up主核销预约-14天时"
	} else if uniqueId == like.NotifyMessageTypePushUpVerify30 {
		description = "提醒up主核销预约-30天时"
	}

	subject, err := d.ActSubject(ctx, sid)
	if err != nil {
		return errors.Wrapf(err, "s.dao.ActSubject err")
	}
	if subject == nil || subject.ID == 0 {
		return errors.Errorf("s.dao.ActSubject nil subject(%+v)", subject)
	}

	req := &tunnel.UpsertCardMsgTemplateReq{
		BizId:        like.PlatformActivityBizID,
		UniqueId:     uniqueId,
		CardUniqueId: sid,
		TriggerType:  "time",
		StartTime:    time.Unix(sTime, 0).Format("2006-01-02 15:04:05"),
		EndTime:      time.Unix(eTime, 0).Format("2006-01-02 15:04:05"),
		CardContent: &tunnelCommon.MsgTemplateCardContent{
			NotifyCode: d.c.UpActReserveNotify.PushUpVerifyID,
			Params:     []string{subject.Name},
			SenderUid:  d.c.UpActReserveNotify.PushUpVerifySenderID,
			JumpUriConfigs: []*tunnelCommon.MsgUriPlatform{
				{
					AndroidUri: d.c.PushVerifyUriConfig.AndroidUri,
					IphoneUri:  d.c.PushVerifyUriConfig.IphoneUri,
					IpadUri:    d.c.PushVerifyUriConfig.IpadUri,
					WebUri:     d.c.PushVerifyUriConfig.WebUri,
					Text:       d.c.PushVerifyUriConfig.Text,
				},
				{
					AndroidUri: d.c.PushVerifyUriResetConfig.AndroidUri,
					IphoneUri:  d.c.PushVerifyUriResetConfig.IphoneUri,
					IpadUri:    d.c.PushVerifyUriResetConfig.IpadUri,
					WebUri:     d.c.PushVerifyUriResetConfig.WebUri,
					Text:       d.c.PushVerifyUriResetConfig.Text,
				},
			},
		},
		Description: description,
	}

	err = retry.WithAttempts(ctx, "UpsertCardMsgTemplate retry", like.UpActReserverelationRetry, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = client.TunnelClient.UpsertCardMsgTemplate(ctx, req)
		if ecode.EqualError(tunnelEcode.DuplicateCard, err) {
			err = nil
		}
		log.Infoc(ctx, "BuildPushCard client.TunnelClient.UpsertCardMsgTemplate req(%+v) err(%+v)", req, err)
		return err
	})

	if err != nil {
		err = errors.Wrapf(err, "BuildPushCard tunnel.UpsertCardMsgTemplate failed req(%+v)", req)
		return
	}

	return
}

func (d *Dao) GetDynamicLotteryInfo(ctx context.Context, sid string, typ int64) (text string, jumpUrl string, err error) {
	params := url.Values{}

	bizType, err := d.GetDynamicLotteryBizID(typ)
	if err != nil {
		err = errors.Wrapf(err, "d.GetDynamicLotteryBizID err type(%+v)", typ)
		return
	}

	params.Set("business_id", sid)
	params.Set("business_type", strconv.FormatInt(bizType, 10))

	rsp := new(struct {
		Code int `json:"code"`
		Data struct {
			FirstPrizeCmt    string `json:"first_prize_cmt"`
			SecondPrizeCmt   string `json:"second_prize_cmt"`
			ThirdPrizeCmt    string `json:"third_prize_cmt"`
			LotteryDetailUrl string `json:"lottery_detail_url"`
		} `json:"data"`
	})
	if err = retry.WithAttempts(ctx, "GetDynamicLotteryPrizeInfo", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		err = d.httpClient.Get(ctx, d.dynamicLotteryPrizeInfo, metadata.String(ctx, metadata.RemoteIP), params, &rsp)
		log.Infoc(ctx, "d.httpClient.Get dynamicLotteryPrizeInfo req(%+v) res(%+v)", params, rsp)
		if err != nil {
			return err
		}
		if rsp.Code != ecode.OK.Code() {
			return fmt.Errorf("d.httpClient.Get response code != 0 params(%+v) result(%+v)", params, rsp)
		}
		return err
	}); err != nil {
		err = errors.Wrapf(err, "d.httpClient.Get err params(%+v)", params)
		return
	}

	if rsp.Data.FirstPrizeCmt == "" {
		err = fmt.Errorf("rsp.Data.FirstPrizeCmt empty rsp(%+v)", rsp)
		return
	}

	text = rsp.Data.FirstPrizeCmt
	if rsp.Data.SecondPrizeCmt != "" {
		text += "、" + rsp.Data.SecondPrizeCmt
	}
	if rsp.Data.ThirdPrizeCmt != "" {
		text += "、" + rsp.Data.ThirdPrizeCmt
	}

	jumpUrl = rsp.Data.LotteryDetailUrl

	return
}

func (d *Dao) GetDynamicLotteryBizID(typ int64) (bizID int64, err error) {
	if typ == int64(api.UpActReserveRelationType_Archive) {
		bizID = like.DynamicLotteryArcBizID
		return
	}
	if typ == int64(api.UpActReserveRelationType_Live) {
		bizID = like.DynamicLotteryLiveBizID
		return
	}
	err = fmt.Errorf("illegal type")
	return
}

func (d *Dao) QueryDateInterval(ctx context.Context, stime, etime string, limit, offset string) (res map[int64]*like.UpActReserveRelation, err error) {
	dateIntervalSQL := fmt.Sprintf(_queryMidByDate, stime, etime, limit, offset)
	rows, err := d.db.Query(ctx, dateIntervalSQL)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "d.db.Query,sql:%v", dateIntervalSQL)
		}
		return
	}
	res = make(map[int64]*like.UpActReserveRelation)
	defer rows.Close()
	for rows.Next() {
		var mid, sid int64
		if err = rows.Scan(&sid, &mid); err != nil {
			err = errors.Wrapf(err, "rows.Scan err")
			return
		}
		res[sid] = &like.UpActReserveRelation{Mid: mid, Sid: sid}
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "d.QueryDateInterval rows.Err() error(%v)", err)
	}
	return
}
