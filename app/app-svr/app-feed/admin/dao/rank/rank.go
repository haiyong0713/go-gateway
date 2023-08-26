package rank

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	xtime "go-common/library/time"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/log"
	"go-common/library/xstr"

	"github.com/jinzhu/gorm"

	"go-gateway/app/app-svr/app-feed/admin/dataplat"
	"go-gateway/app/app-svr/app-feed/admin/model/rank"
)

const (
	_maxTagIDs = 20
)

// query在最终url中是普通sql格式
func (d *Dao) CallDataPlatHiveAPI(c context.Context, api string, query string, res *string) (err error) {
	var response = &dataplat.ResponseHive{
		JobStatusUrl: res,
	}

	var params = url.Values{}
	params.Add("query", query)
	if err = d.DataPlatClient.Post(c, api, params, response); err != nil {
		log.Error("fail to get response, err=%+v", err)
		return
	}

	if response.Code != http.StatusOK {
		err = fmt.Errorf("code:%d, msg:%s", response.Code, response.Msg)
		return
	}
	return
}

// int数据转换成string数据，用于拼接
func intArray2StringArray(in []int) (out []string) {
	out = make([]string, len(in))
	for i, v := range in {
		out[i] = strconv.Itoa(v)
	}
	return
}

// 新建一个榜单配置
func (d *Dao) InsertRankConfig(newConfig *rank.RankConfigReq, uname string) (err error) {
	tids := intArray2StringArray(newConfig.Tids)
	actIds := intArray2StringArray(newConfig.ActIds)
	blacklist := intArray2StringArray(newConfig.Blacklist)

	scoreConfig, _ := json.Marshal(newConfig.ScoreConfig)
	description, _ := json.Marshal(newConfig.Description)

	insertItem := &rank.RankConfig{
		Title:             newConfig.Title,
		STime:             newConfig.STime,
		ETime:             newConfig.ETime,
		Cycle:             newConfig.Cycle,
		PerUpdate:         86400,
		Tids:              strings.Join(tids, ","),
		ActIds:            strings.Join(actIds, ","),
		ArchiveStime:      newConfig.ArchiveStime,
		ArchiveEtime:      newConfig.ArchiveEtime,
		ArchiveSelectMode: newConfig.ArchiveSelectMode,
		ScoreConfig:       string(scoreConfig),
		Blacklist:         strings.Join(blacklist, ","),
		Cover:             newConfig.Cover,
		Description:       string(description),
		State:             0,
		CUser:             uname,
	}
	if err = d.DB.Create(insertItem).Error; err != nil {
		log.Error("dao InsertRankConfig Create() error(%v)", err)
		return
	}
	return
}

// 更新榜单配置
func (d *Dao) UpdateRankConfig(newConfig *rank.EditRankConfigReq, uname string) (err error) {
	tids := intArray2StringArray(newConfig.Tids)
	actIds := intArray2StringArray(newConfig.ActIds)
	blacklist := intArray2StringArray(newConfig.Blacklist)
	description, _ := json.Marshal(newConfig.Description)

	updateItem := &rank.RankConfig{
		Title:             newConfig.Title,
		STime:             newConfig.STime,
		ETime:             newConfig.ETime,
		Cycle:             newConfig.Cycle,
		PerUpdate:         86400,
		Tids:              strings.Join(tids, ","),
		ActIds:            strings.Join(actIds, ","),
		ArchiveStime:      newConfig.ArchiveStime,
		ArchiveEtime:      newConfig.ArchiveEtime,
		ArchiveSelectMode: newConfig.ArchiveSelectMode,
		Blacklist:         strings.Join(blacklist, ","),
		Cover:             newConfig.Cover,
		Description:       string(description),
	}

	if err = d.DB.Model(&rank.RankConfig{}).Where("id = ?", newConfig.ID).Updates(updateItem).Error; err != nil {
		log.Error("dao UpdateRankConfig Update() error(%v)", err)
		return
	}
	return
}

// 更新榜单状态
func (d *Dao) UpdateRankState(rankId, newState int) (err error) {
	updateItem := map[string]interface{}{
		"rank_status": newState,
	}

	if err = d.DB.Model(&rank.RankConfig{}).Where("id = ?", rankId).Updates(updateItem).Error; err != nil {
		log.Error("dao UpdateRankState Update() error(%v)", err)
		return
	}
	return
}

// 查询所有的榜单配置
// 由于state字段真的可能等于0,所以-1代表对state字段无要求
func (d *Dao) QueryRankConfigList(id int, keyword string, state int, T int64, pn, ps int) (configList []*rank.RankConfig, count int, err error) {
	query := d.DB.Model(&rank.RankConfig{})

	if id != 0 {
		query = query.Where("id = ?", id)
	}

	// 当keyword看起来像个数字时,会将keyword解释为榜单id,并且加入查询条件.
	if keyword != "" {
		if id == 0 && isNum(keyword) {
			idFromKeyword, _ := strconv.ParseInt(keyword, 10, 64)
			query = query.Where("id = ? OR rank_name LIKE ?  OR c_user LIKE ?", idFromKeyword, "%"+keyword+"%", "%"+keyword+"%")
		} else {
			query = query.Where("rank_name LIKE ? OR c_user LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	if T != time.Unix(0, 0).Unix() {
		query = query.Where("stime <= ? and etime >= ?", time.Unix(T, 0), time.Unix(T, 0))
	}

	if state != -1 {
		query = query.Where("rank_status = ?", state)
	}

	if err = query.Order("ctime desc").Offset(ps * (pn - 1)).Limit(ps).Find(&configList).Error; err != nil {
		log.Error("QueryRankConfigList Find() error(%v)", err)
	}

	if err = query.Count(&count).Error; err != nil {
		log.Error("QueryRankConfigList Count() error(%v)", err)
		return
	}
	return
}

// check if a string looks like a number
func isNum(str string) bool {
	if _, err := strconv.ParseInt(str, 10, 64); err != nil {
		return false
	}
	return true
}

// 获取某个榜单下原始分数最新的一天的log_date
func (d *Dao) GetOriginalRankScoreListTime(rankId int) (logDate string, mtime time.Time, err error) {
	var (
		logDateTimeRes []*rank.RankScore
	)

	if err = d.DB.Model(&rank.RankScore{}).
		Where("rank_id = ?", rankId).
		Order("mtime desc").Limit(1).Find(&logDateTimeRes).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("GetOriginalRankScoreListTime Find() error(%v)", err)
			return
		}
	}

	if len(logDateTimeRes) > 0 {
		logDate = time.Unix(int64(logDateTimeRes[0].LogDate), 0).Format("2006-01-02")
		mtime = time.Unix(int64(logDateTimeRes[0].MTime), 0)
	} else {
		logDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		mtime = time.Unix(0, 0)
	}

	return
}

// 从rank_score获取原始分数
func (d *Dao) GetOriginalRankScoreList(rankId int, logDate string) (rankScoreList []*rank.RankScore, count int, err error) {
	query := d.DB.Model(&rank.RankScore{}).
		Where("rank_id = ?", rankId).
		Where("log_date = ?", logDate)
	if err = query.Count(&count).Error; err != nil {
		log.Error("GetOriginalRankScoreList Count() error(%v)", err)
		return
	}
	if count == 0 {
		return
	}
	if err = query.Order("score desc").
		Find(&rankScoreList).Error; err != nil {
		log.Error("GetOriginalRankScoreList Find() error(%v)", err)
		return
	}
	return
}

// 获取一些榜单的全部干预
func (d *Dao) GetRankInterventions(rankIdList []int) (interventionMap map[int64]*rank.RankArchiveIntervention, err error) {
	var interventionList []*rank.RankArchiveInterventionDB
	if err = d.DB.Model(&rank.RankArchiveInterventionDB{}).
		Where("rank_id in (?)", rankIdList).
		Find(&interventionList).Error; err != nil {
		log.Error("GetRankInterventions Find error(%v)", err)
		return
	}
	interventionMap = map[int64]*rank.RankArchiveIntervention{}
	for _, intervention := range interventionList {
		extra := &rank.ExtraScore{}
		extraArray := strings.Split(intervention.ExtraScore, ",")
		//nolint:gomnd
		switch len(extraArray) {
		case 3:
			{
				creative, _ := strconv.Atoi(extraArray[2])
				extra.Creative = creative
				reduction, _ := strconv.Atoi(extraArray[1])
				extra.Reduction = reduction
				complete, _ := strconv.Atoi(extraArray[0])
				extra.Complete = complete
			}
		case 2:
			{
				reduction, _ := strconv.Atoi(extraArray[1])
				extra.Reduction = reduction
				complete, _ := strconv.Atoi(extraArray[0])
				extra.Complete = complete
			}
		case 1:
			{
				complete, _ := strconv.Atoi(extraArray[0])
				extra.Complete = complete
			}
		}
		_ = json.Unmarshal([]byte(intervention.ExtraScore), extra)
		interventionMap[intervention.Avid] = &rank.RankArchiveIntervention{
			Avid:     intervention.Avid,
			RankId:   intervention.RankId,
			Rank:     intervention.RankPos,
			IsHidden: intervention.IsHidden,
			Extra:    extra,
		}
	}
	return
}

// NameByTagID .
func (d *Dao) NameByTagID(ctx context.Context, tagIDs []int64) (tagsReply *tag.TagsReply, err error) {
	if tagsReply, err = d.tagClient.Tags(ctx, &tag.TagsReq{Tids: tagIDs}); err != nil {
		log.Error("[NameByTagID] d.tagClient.Tag() tag_id(%s) error(%v)", xstr.JoinInts(tagIDs), err)
		return
	}
	if tagsReply == nil || tagsReply.Tags == nil {
		log.Error("[NameByTagID] d.tagClient.Tag() tag_id(%s) nil reply", xstr.JoinInts(tagIDs))
		return
	}
	return
}

func (d *Dao) NamesByTagIDs(c context.Context, tagIDs []int64) (map[int64]*tag.Tag, error) {
	res := make(map[int64]*tag.Tag)
	pag := len(tagIDs)/_maxTagIDs + 1
	for i := 0; i < pag; i++ {
		maxIndex := (i + 1) * _maxTagIDs
		if maxIndex > len(tagIDs) {
			maxIndex = len(tagIDs)
		}
		tagTemp := tagIDs[i*_maxTagIDs : maxIndex]
		resTmp, err := d.NameByTagID(c, tagTemp)
		if err != nil {
			return nil, err
		}
		if len(resTmp.Tags) > 0 {
			for k, v := range resTmp.Tags {
				if _, ok := res[k]; !ok {
					res[k] = v
				}
			}
		}
	}
	return res, nil
}

// 检查干预排名是否冲突
func (d *Dao) RankArchiveInterventionConflictCheck(param *rank.RankArchiveIntervention) (err error) {
	// 取出所有稿件
	var RAI []rank.RankArchiveInterventionDB
	query := d.DB.Table("rank_intervention").Where("rank_id = ?", param.RankId)
	if err = query.Find(&RAI).Error; err != nil {
		log.Error("dao.rank.RankArchiveInterventionConflictCheck error(%v)", err)
		return
	}

	// 若存在位置相同,稿件不同的干预,说明排名重复了. 当然,排名默认为0是不算重复的
	for k, v := range RAI {
		if param.Rank != 0 && param.Rank == v.RankPos && param.Avid != v.Avid {
			errinfo := fmt.Sprintf("干预排名重复!\n稿件ID:%v\n指定排名:%v\n创建者:%v", RAI[k].Avid, RAI[k].RankPos, RAI[k].CUser)
			//nolint:staticcheck
			fmt.Printf(errinfo)
			err = ecode.Error(ecode.RequestErr, errinfo)
			return
		}

	}
	return

}

// 编辑一条干预,没有则新增
func (d *Dao) RankArchiveEdit(param *rank.RankArchiveIntervention, uname string) (err error) {
	var (
		score []string
	)
	if param.Extra != nil {
		score = []string{strconv.Itoa(param.Extra.Complete), strconv.Itoa(param.Extra.Reduction), strconv.Itoa(param.Extra.Creative)}
	}

	scoreString := strings.Join(score, ",")
	interventionDB := rank.RankArchiveInterventionDB{
		RankId:     param.RankId,
		Avid:       param.Avid,
		IsHidden:   param.IsHidden,
		RankPos:    param.Rank,
		ExtraScore: scoreString,
		CUser:      uname,
	}

	// 检查有无记录
	var RAI []rank.RankArchiveInterventionDB
	query := d.DB.Model(&rank.RankArchiveInterventionDB{}).Where("rank_id = ?", param.RankId).Where("avid = ?", param.Avid)
	if err = query.Find(&RAI).Error; err != nil {
		log.Error("dao.rank.RankArchiveEdit error(%v)", err)
		return
	}

	// 同样的rankid和avid,不可能对应多条记录
	if len(RAI) > 1 {
		log.Error("dao.rank.RankArchiveEdit error(%v)", "more than one intervention!")
		err = ecode.Error(-1, "more than one intervention!")
		return
	}
	// 更新这条记录
	if len(RAI) == 1 {
		// 在使用结构体更新时,GORM会忽略为0的字段.但我们确实需要将某些值设置为0,因此要使用map更新.
		RAIMap := map[string]interface{}{}
		RAIMap["rank_id"] = param.RankId
		RAIMap["avid"] = param.Avid
		RAIMap["is_hidden"] = param.IsHidden
		RAIMap["rank_pos"] = param.Rank
		RAIMap["extra_score"] = scoreString
		RAIMap["c_user"] = uname
		if err = query.Updates(RAIMap).Error; err != nil {
			log.Error("[RankArchiveEdit]  d.DB.Update() error(%v)", err)
			return
		}

	}
	// 新建这条记录
	if len(RAI) == 0 {
		if err = d.DB.Create(&interventionDB).Error; err != nil {
			log.Error("[RankArchiveEdit]  d.DB.Create() error(%v)", err)
			return
		}
	}
	// todo 写日志
	return
}

// 在榜单中手动添加稿件
func (d *Dao) RankArchiveAdd(rankid int64, avid int64) (err error) {
	type AvManuallyAddList struct {
		Amal string `gorm:"column:av_manually_added"`
	}
	var (
		amal    AvManuallyAddList
		amals   []string
		avidStr = strconv.FormatInt(avid, 10)
	)

	// 从榜单配置中取出所有手动添加的稿件
	query := d.DB.Table("rank_config_list").Where("id = ?", rankid)
	if err = query.Select("av_manually_added").Find(&amal).Error; err != nil {
		log.Error("[RankArchiveAdd]  d.DB.Select() error(%v)", err)
		return
	}

	// 检查是否存在重复稿件
	amals = strings.Split(amal.Amal, ",")
	for _, v := range amals {
		if avidStr == v {
			log.Error("[RankArchiveAdd]  repeated avid error(%v)", err)
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("稿件%v已添加过，请勿重复添加!", v))
			return
		}
	}

	amals = append(amals, avidStr)

	// 插入刚刚添加的稿件
	if err = query.Update("av_manually_added", strings.Join(amals, ",")).Error; err != nil {
		d.DB.Rollback()
		log.Error("[RankArchiveAdd]  d.DB.Update() error(%v)", err)
		return
	}
	// todo 写日志
	return
}

// 榜单操作 上线/下线
func (d *Dao) RankOption(rankid int, state int) (err error) {
	query := d.DB.Table("rank_config_list").Where("id = ?", rankid)
	if err = query.Updates(map[string]interface{}{"rank_status": state}).Error; err != nil {
		log.Error("[RankArchiveAdd]  d.DB.Update() error(%v)", err)
		return
	}
	return
}

// 发榜
func (d *Dao) RankPublish(rankid int, avrank []string, username string) (err error) {
	if rankid == 0 {
		err = ecode.Error(ecode.RequestErr, "没有榜单ID")
		return
	}

	rankhistoryItemDB := rank.RankHistoryDB{
		RankId:     rankid,
		LogData:    xtime.Time(time.Now().Unix()),
		ScoreAvids: strings.Join(avrank, ","),
		CUser:      username,
	}
	query := d.DB.Table("rank_history")
	if err = query.Create(&rankhistoryItemDB).Error; err != nil {
		log.Error("dao RankPublish Create() error(%v)", err)
		return
	}
	if err = d.DB.Table("rank_config_list").
		Where("id = ?", rankid).
		Update("rank_history_id", rankhistoryItemDB.ID).Error; err != nil {
		log.Error("dao RankPublish Update() error(%v)", err)
		return
	}
	return
}

// 结榜
func (d *Dao) RankTerminate(historyId int, tmcontent *rank.TernimateContent, avrank []string, username string) (err error) {
	if historyId == 0 {
		return
	}

	// 结榜配置以json形式存储
	//nolint:staticcheck
	tmcontentJson, err := json.Marshal(tmcontent.TmContent)
	rankhistoryItemDB := rank.RankHistoryDB{
		FinalRankConfig: string(tmcontentJson),
		CUser:           username,
		ScoreAvids:      strings.Join(avrank, ","),
	}

	query := d.DB.Table("rank_history")
	if err = query.Where("id = ?", historyId).Updates(&rankhistoryItemDB).Error; err != nil {
		log.Error("dao RankTerminate Create() error(%v)", err)
		return
	}
	stateChange := d.DB.Table("rank_config_list")
	// 榜单状态：0-未开始 1-进行中 2-已结束 3-已结榜. 如果结榜操作成功,无论状态是1还是2,都会变成3
	if err = stateChange.Where("id = ?", tmcontent.Id).Updates(map[string]interface{}{"rank_status": 3}).Error; err != nil {
		log.Error("dao RankTerminate Updates() error(%v)", err)
		return
	}
	return
}

// 获得当前榜单的history配置
func (d *Dao) FindRankHistoryConfig(historyId int) (historyConfig *rank.RankHistoryDB, err error) {
	if historyId == 0 {
		historyConfig = &rank.RankHistoryDB{}
		return
	}

	var (
		historyList []*rank.RankHistoryDB
	)
	if err = d.DB.Model(&rank.RankHistoryDB{}).
		Where("id = ?", historyId).
		Find(&historyList).Error; err != nil {
		log.Error("dao FindRankHistoryConfig Find() error(%v)", err)
		return
	}

	if len(historyList) > 0 {
		historyConfig = historyList[0]
	}

	return
}

// 获取今天零点之前的最近的一个榜单配置
func (d *Dao) FindLastRankHistoryConfig() (historyConfig *rank.RankHistoryDB, err error) {
	var (
		historyList []*rank.RankHistoryDB
	)
	currentTime := time.Now().Format("2006-01-02") + " 00:00:00"
	if err = d.DB.Model(&rank.RankHistoryDB{}).
		Where("ctime < ?", currentTime).
		Order("ctime desc").Limit(1).
		Find(&historyList).Error; err != nil {
		log.Error("dao FindRankHistoryConfig Find() error(%v)", err)
		return
	}

	if len(historyList) > 0 {
		historyConfig = historyList[0]
	}

	return
}

// 获取任务状态
func (d *Dao) GetTaskFlag(c context.Context, key string) (exist bool, status string, err error) {
	var (
		conn = d.Rds.Get(c)
	)
	defer conn.Close()

	if exist, err = redis.Bool(conn.Do("EXISTS", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(EXIST, %s) error(%v)", key, err)
		return
	}

	if !exist {
		return
	}

	if status, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(EXIST, %s) error(%v)", key, err)
		return
	}
	return
}

// 设置为任务状态
func (d *Dao) SetTaskFlag(c context.Context, key string, value string, expire int64) (err error) {
	if key == "" {
		return
	}

	var (
		conn = d.Rds.Get(c)
	)
	defer conn.Close()

	exist := false
	if exist, _, err = d.GetTaskFlag(c, key); err != nil {
		return
	}

	if !exist {
		if _, err = conn.Do("SET", key, value); err != nil {
			if err == redis.ErrNil {
				err = nil
				return
			}
			log.Error("conn.Do(SET, %s) error(%v)", key, err)
			return
		}
	}

	if _, err = conn.Do("EXPIRE", key, expire); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(EXPIRE, %s) error(%v)", key, err)
		return
	}

	return
}

// 删除标记
func (d *Dao) DeleteTaskFlag(c context.Context, key string) (err error) {
	if key == "" {
		return
	}

	var (
		conn = d.Rds.Get(c)
	)
	defer conn.Close()

	exist := false
	if exist, _, err = d.GetTaskFlag(c, key); err != nil {
		return
	}

	if exist {
		if _, err = conn.Do("DEL", key); err != nil {
			if err == redis.ErrNil {
				err = nil
				return
			}
			log.Error("conn.Do(DEL, %s) error(%v)", key, err)
			return
		}
	}

	return
}
