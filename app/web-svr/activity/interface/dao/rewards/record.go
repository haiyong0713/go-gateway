package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"go-gateway/app/web-svr/activity/interface/tool"
	"sort"
	"time"
)

//create table if not exists rewards_award_record_01 (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
//unique_id varchar(50)  NOT NULL DEFAULT '0' COMMENT '幂等ID',
//state tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 开始领取 1 领取完成 2 领取失败',
//activity_id int(11) unsigned NOT NULL DEFAULT '0' COMMENT '活动id',
//award_id int(11) unsigned NOT NULL DEFAULT '0' COMMENT '奖励id',
//award_type varchar(20)  NOT NULL DEFAULT '0' COMMENT '奖励类型',
//award_name varchar(50)  NOT NULL DEFAULT '0' COMMENT '奖励名称',
//award_config_content varchar(10000) NOT NULL DEFAULT "" COMMENT '奖励配置内容'
//ctime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//mtime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
//PRIMARY KEY (mid,unique_id),
//KEY ix_m_u_s (mid,unique_id,state))
//KEY ix_mtime (mtime))
//ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品用户发放记录表';

//create table if not exists rewards_award_fail_record (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
//unique_id varchar(50)  NOT NULL DEFAULT '0' COMMENT '幂等ID',
//error_msg varchar(2000)  NOT NULL DEFAULT '0' COMMENT '失败原因',
//retry_state tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 尚未重试 1 重试完成 ',
//ctime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//mtime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
//PRIMARY KEY (mid,unique_id),
//KEY ix_mtime (mtime))
//ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品发放失败记录表';

const (
	sql4GetAwardSendStatus = `
SELECT state 
FROM   rewards_award_record_%v 
WHERE  mid =? 
       AND activity_id =? 
       AND unique_id =? `

	sql4GetAwardSendStatusByMidAndUniqId = `
SELECT state 
FROM   rewards_award_record_%v 
WHERE  mid =?
       AND unique_id =? `

	sql4InsertAwardAddSendRecord = `
INSERT INTO rewards_award_record_%v 
            ( 
                        mid, 
                        activity_id, 
                        unique_id,
                        state,
                        award_id, 
                        award_type, 
                        award_name, 
                        award_config_content, 
                        business
            ) 
            VALUES 
            ( 
                        ?, 
                        ?, 
                        ?, 
                        ?, 
                        ?, 
                        ?, 
                        ?, 
                        ?, 
                        ?
            )
ON DUPLICATE KEY update state = ?
`

	sql4UpdateAwardRecordState = `
UPDATE rewards_award_record_%v 
SET    state=? 
WHERE  mid=? 
AND    activity_id=?
AND    unique_id=?
`

	sql4UpdateAwardRecordStateWithExtraInfo = `
UPDATE rewards_award_record_%v 
SET    state=? , extra_info=?
WHERE  mid=? 
AND    activity_id=?
AND    unique_id=?
`

	sql4InsertAwardFailRecord = `
INSERT INTO rewards_award_fail_record 
            (mid, 
             activity_id,
             award_id,
             unique_id,
             error_msg, 
             retry_state) 
VALUES     (?, 
            ?, 
            ?, 
            ?, 
            ?, 
            0) `
	sql4UpDateAwardFailRecord = `
UPDATE rewards_award_fail_record 
SET retry_state = ? 
WHERE mid = ?
AND activity_id =?
AND award_id=?
AND unique_id=?
`

	sql4GetAwardSendRecordByMidAndActivityId = `
SELECT   r.mid, 
         r.award_id,
         r.award_name, 
         c.activity_id,
         c.award_type, 
         c.icon_url, 
         r.ctime, 
         r.extra_info 
FROM     rewards_award_record_%v r, 
         rewards_award_config c 
WHERE    r.award_id=c.id 
AND      r.mid =? 
AND      r.activity_id IN (%s) 
ORDER BY r.ctime DESC
LIMIT    ?`

	sql4GetAwardSendCountByMidAndActivityId = `
SELECT count(*)
FROM   rewards_award_record_%v
WHERE  mid =? 
AND    activity_id =?`

	sql4InsertAwardAddress = `
INSERT INTO rewards_entity_award_addresses 
            (mid, 
             activity_id,
             activity_name, 
             address_id) 
VALUES     (?, 
            ?, 
            ?,  
            ?) `
	sql4GetAwardAddressCount = `
SELECT address_id 
FROM   rewards_entity_award_addresses 
WHERE  mid =? 
       AND activity_id =? LIMIT 1`
)

const (
	AwardSentStateNotFound = -1
	AwardSentStateInit     = 0
	AwardSentStateStarting = 1
	AwardSentStateOK       = 2
	AwardSentStateFail     = 3
)

const (
	cacheKey4AwardSendStatus = "award_send_status_1215_%v_%v_%v"
)

func (d *Dao) getAwardSendStatusCacheKey(mid, activityId int64, uniqueId string) string {
	return fmt.Sprintf(cacheKey4AwardSendStatus, mid, activityId, uniqueId)
}

func (d *Dao) AddAwardSendStatusInCache(ctx context.Context, mid, activityId int64, uniqueId string, state int64) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	ttl := int64(15)
	if state == 1 { //state is send,  will never change back to not send, so make ttl longer
		ttl = tool.CalculateExpiredSeconds(1)
	}
	for i := 0; i < 3; i++ {
		_, err = conn.Do("SETEX", d.getAwardSendStatusCacheKey(mid, activityId, uniqueId), ttl, state)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "rewards d.AddAwardSendStatusInCache: %v", err)
	}
	return err
}

func (d *Dao) getAwardSendStatusInCache(ctx context.Context, mid, activityId int64, uniqueId string) (state int64, err error) {
	rc := component.GlobalBnjCache.Get(ctx)
	defer rc.Close()
	state, err = redis.Int64(rc.Do("GET", d.getAwardSendStatusCacheKey(mid, activityId, uniqueId)))
	return
}

func (d *Dao) getAwardSendStatusInDB(ctx context.Context, mid, activityId int64, uniqueId string) (state int64, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(sql4GetAwardSendStatus, userHit(mid)), mid, activityId, uniqueId)
	err = row.Scan(&state)
	if err == sql.ErrNoRows {
		state = AwardSentStateNotFound
		err = nil
	}
	return
}

// GetAwardSendStatusInDBByMidAndUniqId: 直接读取DB, 不支持大并发
func (d *Dao) GetAwardSendStatusInDBByMidAndUniqId(ctx context.Context, mid int64, uniqueId string) (state int64, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(sql4GetAwardSendStatusByMidAndUniqId, userHit(mid)), mid, uniqueId)
	err = row.Scan(&state)
	if err == sql.ErrNoRows {
		state = AwardSentStateNotFound
		err = nil
	}
	return
}

func (d *Dao) getAwardSendStatus(ctx context.Context, mid, activityId int64, uniqueId string) (state int64, err error) {
	var shouldUpdateCache bool
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.IsAwardAlreadySend error: %v", err)
		}
		if shouldUpdateCache {
			d.AddAwardSendStatusInCache(context.Background(), mid, activityId, uniqueId, state)
		}
	}()
	state, err = d.getAwardSendStatusInCache(ctx, mid, activityId, uniqueId)
	if err == nil {
		return
	}
	state, err = d.getAwardSendStatusInDB(ctx, mid, activityId, uniqueId)
	shouldUpdateCache = err == nil
	return
}

// IsAwardAlreadySend: 根据mid+uniqueId判断奖励是否发放,会读取缓存,极端情况下会出现不一致
func (d *Dao) IsAwardAlreadySend(ctx context.Context, mid, activityId int64, uniqueId string) (send bool, err error) {
	var s int64
	s, err = d.getAwardSendStatus(ctx, mid, activityId, uniqueId)
	if err != nil {
		return
	}
	send = s != AwardSentStateNotFound
	return
}

// IsAwardAlreadySendStrict: 根据mid+uniqueId判断奖励是否发放,不会读取缓存,准确性较高
func (d *Dao) IsAwardAlreadySendStrict(ctx context.Context, mid, activityId int64, uniqueId string) (send bool, err error) {
	var s int64
	s, err = d.getAwardSendStatusInDB(ctx, mid, activityId, uniqueId)
	if err != nil {
		return
	}
	send = s != AwardSentStateNotFound
	return
}

func (d *Dao) addAwardFailRecord(ctx context.Context, mid, activityId, awardId int64, uniqueId string, sendErr error) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "addAwardFailRecord for mid: %v, activityId: %v, uniqueId: %v,error: %v", mid, activityId, uniqueId, err)
		}
	}()
	for i := 0; i < 3; i++ {
		_, err = d.db.Exec(ctx, sql4InsertAwardFailRecord, mid, activityId, awardId, uniqueId, sendErr.Error())
		if err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	return err
}

func (d *Dao) updateAwardState(ctx context.Context, mid, activityId int64, uniqueId string, state int64, extraInfo map[string]string) (err error) {
	shouldUpdateCache := false
	defer func() {
		if err != nil {
			log.Errorc(ctx, "updateAwardState for mid: %v, activityId: %v, uniqueId: %v,error: %v", mid, activityId, uniqueId, err)
		}
		if shouldUpdateCache {
			_ = d.AddAwardSendStatusInCache(ctx, mid, activityId, uniqueId, state)
		}
	}()
	if len(extraInfo) == 0 {
		for i := 0; i < 3; i++ {
			_, err = d.db.Exec(ctx, fmt.Sprintf(sql4UpdateAwardRecordState, userHit(mid)), state, mid, activityId, uniqueId)
			if err == nil {
				break
			}
		}
	} else {
		extraInfoBs, _ := json.Marshal(extraInfo)
		for i := 0; i < 3; i++ {
			_, err = d.db.Exec(ctx, fmt.Sprintf(sql4UpdateAwardRecordStateWithExtraInfo, userHit(mid)), state, string(extraInfoBs), mid, activityId, uniqueId)
			if err == nil {
				break
			}
		}
	}

	shouldUpdateCache = err == nil
	return
}

// InitAwardSentRecord: 异步发放前插入init记录
// updateDB: 是否更新DB, 可避免消息队列丢失导致丢数据
// updateDB=true: 一致性高,容忍消息丢失
// updateDB=false: 性能高,需要提供额外的对账机制
func (d *Dao) InitAwardSentRecord(ctx context.Context, mid int64, uniqueId, business string,
	awardConfig *api.RewardsAwardInfo, updateDB bool) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "InitAwardSentRecord mid: %v, uniqueId: %v, error: %v", mid, uniqueId, err)
			return
		}
		//更新缓存中的状态
		_ = d.updateAwardState(ctx, mid, awardConfig.ActivityId, uniqueId, AwardSentStateInit, nil)
	}()
	if updateDB {
		for i := 0; i < 3; i++ {
			_, err = d.db.Exec(ctx, fmt.Sprintf(sql4InsertAwardAddSendRecord, userHit(mid)),
				mid, awardConfig.ActivityId, uniqueId, AwardSentStateInit, awardConfig.Id,
				awardConfig.Type, awardConfig.Name, awardConfig.JsonStr, business, AwardSentStateInit)
			if err == nil {
				break
			}
		}
	}

	return
}

func (d *Dao) SendAwardByFunc(ctx context.Context, mid int64, uniqueId, business string,
	awardConfig *api.RewardsAwardInfo, sendFunc func() (map[string]string, error)) (err error) {
	//检查发放状态
	state, err := d.getAwardSendStatusInDB(ctx, mid, awardConfig.ActivityId, uniqueId)
	if err != nil {
		log.Errorc(ctx, "SendAwardByFunc award already sent mid: %v,uniqueId: %v, error: %v", mid, uniqueId, err)
		err = ecode.RewardsAwardSendFail
		return
	}
	if state == AwardSentStateOK || state == AwardSentStateStarting {
		err = nil
		log.Infoc(ctx, "RewardsAwardAlreadySent for mid: %v, uniqueId: %v, business: %v, awardId: %v",
			mid, uniqueId, business, awardConfig.Id)
		return
	}
	//初始化发放记录
	_, err = d.db.Exec(ctx, fmt.Sprintf(sql4InsertAwardAddSendRecord, userHit(mid)),
		mid, awardConfig.ActivityId, uniqueId, AwardSentStateStarting, awardConfig.Id,
		awardConfig.Type, awardConfig.Name, awardConfig.JsonStr, business, AwardSentStateStarting)
	if err != nil {
		log.Errorc(ctx, "SendAwardByFunc add send record error mid: %v,uniqueId: %v,err: %v", mid, uniqueId, err)
		err = ecode.RewardsAwardSendFail
		return
	}
	var extraInfo map[string]string
	extraInfo, err = sendFunc()
	if err != nil {
		log.Errorc(ctx, "SendAwardByFunc mid: %v, uniqueId: %v, error: %v", mid, uniqueId, err)
		_ = d.updateAwardState(ctx, mid, awardConfig.ActivityId, uniqueId, AwardSentStateFail, nil)
		_ = d.addAwardFailRecord(ctx, mid, awardConfig.ActivityId, awardConfig.Id, uniqueId, err)
		err = ecode.RewardsAwardSendFail
		return
	}
	err = d.updateAwardState(ctx, mid, awardConfig.ActivityId, uniqueId, AwardSentStateOK, extraInfo)
	if err != nil {
		err = ecode.RewardsAwardSendFail
		return
	}
	go func() {
		_ = d.incrAwardCountCacheByMidAndActivity(ctx, mid, awardConfig.ActivityId, 1)
	}()
	if len(extraInfo) != 0 { //本次发放存在附属信息,删除缓存,等待回源读取
		_ = d.CacheDelUserAwardRecord(ctx, mid, awardConfig.ActivityId)
	}
	return
}

func (d *Dao) updateAwardFailRecordStateToOk(ctx context.Context, mid, activityId, awardId int64, uniqueId string) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "updateAwardFailRecordStateToOk for mid: %v, activityId: %v, uniqueId: %v,error: %v", mid, activityId, uniqueId, err)
		}
	}()
	for i := 0; i < 3; i++ {
		_, err = d.db.Exec(ctx, sql4UpDateAwardFailRecord, 1, mid, activityId, awardId, uniqueId)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	return err
}

func (d *Dao) RetrySendAwardByFunc(ctx context.Context, mid int64, uniqueId, business string,
	awardConfig *api.RewardsAwardInfo, sendFunc func() (map[string]string, error)) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "RetrySendAwardByFunc error: %v", err)
		}
	}()
	//检查发放状态
	state, err := d.getAwardSendStatusInDB(ctx, mid, awardConfig.ActivityId, uniqueId)
	if err != nil {
		log.Errorc(ctx, "RetrySendAwardByFunc award already sent mid: %v,uniqueId: %v", mid, uniqueId)
		return
	}
	if state == AwardSentStateNotFound {
		err = ecode.ActivityNoAward
		return
	}
	if state == AwardSentStateOK {
		_ = d.updateAwardFailRecordStateToOk(ctx, mid, awardConfig.ActivityId, awardConfig.Id, uniqueId)
		return
	}
	var extraInfo map[string]string
	extraInfo, err = sendFunc()
	if err != nil {
		_ = d.updateAwardState(ctx, mid, awardConfig.ActivityId, uniqueId, AwardSentStateFail, nil)
		return
	}
	err = d.updateAwardState(ctx, mid, awardConfig.ActivityId, uniqueId, AwardSentStateOK, extraInfo)
	if err != nil {
		return
	}
	_ = d.updateAwardFailRecordStateToOk(ctx, mid, awardConfig.ActivityId, awardConfig.Id, uniqueId)
	_ = d.incrAwardCountCacheByMidAndActivity(ctx, mid, awardConfig.ActivityId, 1)
	return
}

func getAwardRecordCacheKey(mid, activityId int64) string {
	return fmt.Sprintf("award_record_%v_%v", mid, activityId)
}

// resetUserAwardRecordCache: 向缓存中存储用户抽奖记录.
// 存储结构为:
// key-->mid:activityId
// value-->json.Marshal(records)
// records为空时,仍存储空数组`[]`到缓存, 避免频繁回源
func (d *Dao) resetUserAwardRecordCache(mid int64, activityIds []int64, res []*model.AwardSentInfo) {
	if len(activityIds) == 0 && len(res) == 0 {
		return
	}
	ctx := context.Background()
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	//根据activityId拆分数组
	tmpRes := make(map[int64][]*model.AwardSentInfo, 0)
	for _, r := range res {
		tmpR := r
		if tmpRes[tmpR.ActivityId] == nil {
			tmpRes[tmpR.ActivityId] = make([]*model.AwardSentInfo, 0)
		}
		tmpRes[tmpR.ActivityId] = append(tmpRes[tmpR.ActivityId], tmpR)
	}
	//查看是否存在空记录的activityId,构造缓存标记`{}`
	for _, aid := range activityIds {
		if tmpRes[aid] == nil {
			tmpRes[aid] = make([]*model.AwardSentInfo, 0)
		}
	}
	//存储记录到缓存中
	for aid, records := range tmpRes {
		bs, err := json.Marshal(records)
		if err != nil {
			log.Errorc(ctx, "resetUserAwardRecordCache set cache json.Marshal error: %v", err)
			continue
		}
		for i := 0; i < 3; i++ {
			_, err = conn.Do("SETEX", getAwardRecordCacheKey(mid, aid), tool.ExpiredRedisKeyAtDayEarly(), bs)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Errorc(ctx, "resetUserAwardRecordCache set cache error: %v", err)
			continue
		}
	}
}

// AddSingleRecordCache: 添加单个中间记录到缓存中
func (d *Dao) AddSingleRecordCache(mid, activityId int64, info *api.RewardsSendAwardReply) (err error) {
	if activityId == 0 || info == nil {
		return
	}
	var bs []byte
	ctx := context.Background()
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		if err != nil {
			log.Errorc(ctx, "AddSingleRecordCache mid %v activityId %v error: %v", mid, activityId, err)
			//缓存可能已经污染,删除后返回等待下次访问主动回源DB
			_, errDel := conn.Do("DEL", getAwardRecordCacheKey(mid, activityId))
			if errDel != nil {
				log.Errorc(ctx, "AddSingleRecordCache mid %v activityId %v del cache error: %v", mid, activityId, errDel)
			}
		}
		_ = conn.Close()
	}()
	oldRecord := make([]*model.AwardSentInfo, 0)
	newRecord := make([]*model.AwardSentInfo, 0)

	for i := 0; i < 3; i++ {
		bs, err = redis.Bytes(conn.Do("GET", getAwardRecordCacheKey(mid, activityId)))
		if err == redis.ErrNil {
			err = nil
			return
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(bs, &oldRecord)
	if err != nil {
		return
	}

	newRecord = append(newRecord, info.ToSentInfo()) //append this info to head(sort by ctime desc)
	newRecord = append(newRecord, oldRecord...)
	if len(newRecord) > 100 {
		newRecord = newRecord[:100]
	}
	bs, err = json.Marshal(newRecord)
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ {
		_, err = conn.Do("SETEX", getAwardRecordCacheKey(mid, activityId), tool.CalculateExpiredSeconds(30), bs)
		if err == nil {
			break
		}
	}
	return
}

// GetAwardRecordByMidAndActivityIdsWithCache: 支持缓存的用户抽奖记录获取
// 执行流程:
// 1.从缓存中获取各个aid的中奖记录
// 2.如果有cache miss的aid,则从DB中获取
// 3.回种缓存
func (d *Dao) GetAwardRecordByMidAndActivityIdsWithCache(ctx context.Context, mid int64, activityIds []int64, limit int64) (res []*model.AwardSentInfo, err error) {
	res = make([]*model.AwardSentInfo, 0)
	if len(activityIds) == 0 {
		return
	}
	conn := component.GlobalBnjCache.Get(ctx)
	shouldUpdateCache := false
	cacheRes := make(map[int64][]*model.AwardSentInfo, 0)
	cacheMissingAids := make([]int64, 0)
	dbRes := make([]*model.AwardSentInfo, 0)
	defer func() {
		_ = conn.Close()
		if err != nil {
			log.Errorc(ctx, "GetAwardRecordByMidAndActivityIdWithCache mid %v, activityId %v, error: %v", mid, activityIds, err)
		}
		if shouldUpdateCache {
			d.resetUserAwardRecordCache(mid, cacheMissingAids, dbRes)
		}
	}()

	//从缓存中获取aid中奖记录
	for _, aid := range activityIds {
		bs, err := redis.Bytes(conn.Do("GET", getAwardRecordCacheKey(mid, aid)))
		if err == redis.ErrNil {
			cacheMissingAids = append(cacheMissingAids, aid)
			continue
		}
		record := make([]*model.AwardSentInfo, 0)
		err = json.Unmarshal(bs, &record)
		if err != nil {
			cacheMissingAids = append(cacheMissingAids, aid)
			continue
		}
		cacheRes[aid] = record
	}
	//将缓存中的记录拼装到结果集中
	for _, record := range cacheRes {
		res = append(res, record...)
	}
	//缓存全部命中,直接返回
	if len(cacheMissingAids) == 0 {
		sort.Slice(res, func(i, j int) bool {
			return res[i].SentTime > res[j].SentTime
		})
		return
	}
	//缓存miss的aid,通过DB进行查询
	dbRes, err = d.GetAwardRecordByMidAndActivityIdsFromDB(ctx, mid, cacheMissingAids, limit)
	//将DB中的记录拼装到结果集中
	res = append(res, dbRes...)
	shouldUpdateCache = err == nil
	return
}

func (d *Dao) GetAwardRecordByMidAndActivityIdsFromDB(ctx context.Context, mid int64, activityIds []int64, limit int64) (res []*model.AwardSentInfo, err error) {
	res = make([]*model.AwardSentInfo, 0)
	defer func() {
		if err != nil {
			tool.AddDBErrMetrics("award_record")
			log.Errorc(ctx, "GetAwardRecordByMidAndActivityIdsFromDB mid %v, activityId %v, error: %v", mid, activityIds, err)
		}
	}()
	tool.AddDBBackSourceMetrics("award_record")
	rows, err := d.db.Query(ctx, fmt.Sprintf(sql4GetAwardSendRecordByMidAndActivityId, userHit(mid), xstr.JoinInts(activityIds)), mid, limit)
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		extraInfo := ""
		t := &model.AwardSentInfo{ExtraInfo: make(map[string]string)}
		if err = rows.Scan(&t.Mid, &t.AwardId, &t.AwardName, &t.ActivityId, &t.Type, &t.IconUrl, &t.SentTime, &extraInfo); err != nil {
			return
		}
		if extraInfo != "" {
			err = json.Unmarshal([]byte(extraInfo), &t.ExtraInfo)
			if err != nil {
				return
			}
		}
		res = append(res, t)
	}
	err = rows.Err()
	return
}

func getAwardCountCacheKey(mid, activityId int64) string {
	return fmt.Sprintf("award_count_%v_%v", mid, activityId)
}

// incrAwardCountCacheByMidAndActivity: 增加缓存中的活动中间次数
func (d *Dao) incrAwardCountCacheByMidAndActivity(ctx context.Context, mid, activityId, incrCount int64) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
		if err != nil {
			log.Errorc(ctx, "incrAwardCountCacheByMidAndActivity")
		}
	}()
	for i := 0; i < 3; i++ {
		_, err = conn.Do("INCRBY", getAwardCountCacheKey(mid, activityId), incrCount)
		if err == nil {
			break
		}
	}
	return
}

func (d *Dao) GetAwardCountByMidAndActivityIdWithCache(ctx context.Context, mid, activityId int64) (count int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	shouldUpdateCache := false
	defer func() {
		if err != nil {
			log.Errorc(ctx, "GetAwardCountByMidAndActivityIdWithCache mid %v, activityId %v, error: %v", mid, activityId, err)
		}
		if shouldUpdateCache {
			_, errSet := conn.Do("SETEX", getAwardCountCacheKey(mid, activityId), tool.ExpiredRedisKeyAtDayEarly(), count)
			if errSet != nil {
				log.Errorc(ctx, "GetAwardCountByMidAndActivityIdWithCache set cache error: %v", errSet)
			}
		}
		_ = conn.Close()
	}()
	count, err = redis.Int64(conn.Do("GET", getAwardCountCacheKey(mid, activityId)))
	if err == nil {
		return
	}
	count, err = d.GetAwardCountByMidAndActivityIdFromDB(ctx, mid, activityId)
	shouldUpdateCache = err == nil
	return
}

func (d *Dao) GetAwardCountByMidAndActivityIdFromDB(ctx context.Context, mid, activityId int64) (count int64, err error) {
	defer func() {
		if err != nil {
			tool.AddDBErrMetrics("award_count")
			log.Errorc(ctx, "GetAwardCountByMidAndActivityIdFromDB mid %v, activityId %v, error: %v", mid, activityId, err)
		}
	}()
	row := d.db.QueryRow(ctx, fmt.Sprintf(sql4GetAwardSendCountByMidAndActivityId, userHit(mid)), mid, activityId)
	err = row.Scan(&count)
	return
}

// 查看奖励的地址是否已填写
func (d *Dao) IsActivityAddressExists(ctx context.Context, mid, activityId int64) (addressId int64, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "IsActivityAddressExists error: mid: %v, activityId: %v", mid, activityId)
		}
	}()
	row := d.db.QueryRow(ctx, sql4GetAwardAddressCount, mid, activityId)
	err = row.Scan(&addressId)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

// 领取到实体奖励后, 用户填写地址
func (d *Dao) AddActivityAddress(ctx context.Context, mid int64, info *api.RewardsAwardInfo, addressId int64) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "AddActivityAddress error: info: %+v, addressId: %v", info, addressId)
		}
	}()
	_, err = d.db.Exec(ctx, sql4InsertAwardAddress, mid, info.ActivityId, info.ActivityName, addressId)
	return err
}

// 删除用户的中奖记录缓存
func (d *Dao) CacheDelUserAwardRecord(ctx context.Context, mid, activityId int64) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	_, err = conn.Do("DEL", getAwardRecordCacheKey(mid, activityId))
	if err != nil {
		log.Errorc(ctx, "CacheDelUserAwardRecord mid %v activityId %v error: %v", mid, activityId)
	}
	return
}
