package rank

import (
	"fmt"
	go_common_library_time "go-common/library/time"
	"strconv"
	"time"
)

const (
	// RankTypeUp  up主维度
	RankTypeUp = 1
	// RankTypeArchive  稿件维度
	RankTypeArchive = 2
	// InterventionTypeWhite 白名单
	InterventionTypeWhite = 1
	// InterventionTypeBlack 黑名单
	InterventionTypeBlack = 2
	// InterventionObjectUp 干预对象up主
	InterventionObjectUp = 1
	// InterventionObjectArchive 干预对象稿件
	InterventionObjectArchive = 2
	// SIDSourceAid 视频数据源
	SIDSourceAid = 1
	// StatisticsTimeTypeNature 自然日
	StatisticsTimeTypeNature = 1
	// StatisticsTimeTypeCircle 统计周期
	StatisticsTimeTypeCircle = 2
	// StatisticsTimeState 有效
	StatisticsTimeState = 1
	// AutoPublishYes 自动发布
	AutoPublishYes = 1
	// AutoPublishNo 不自动发布
	AutoPublishNo = 0
	// SnapshotStateNotNormal 不录入
	SnapshotStateNotNormal = 0
	// SnapshotStateNormal 录入
	SnapshotStateNormal = 1
	// OidStateNotNormal 不录入
	OidStateNotNormal = 0
	// OidStateNormal 录入
	OidStateNormal = 1
	// NeedSnapshot 需要快照
	NeedSnapshot = true
	// NeedNotSnapshot 不需要快照
	NeedNotSnapshot = false
	// NeedDiffScore 需要diff
	NeedDiffScore = true
	// NeedNotDiffScore 不需要diff
	NeedNotDiffScore = false
)

const (
	// RankAttributeAll 总榜
	RankAttributeAll = uint(0)
	// RankAttributeDay 日榜
	RankAttributeDay = uint(1)
	// RankAttributeWeek 周榜
	RankAttributeWeek = uint(2)
	// RankAttributeMonth 月榜
	RankAttributeMonth = uint(3)
)

const (
	// RankLogStateDone 计算完成
	RankLogStateDone = 1
	// RankLogStateLearn 训练中
	RankLogStateLearn = 2
)

// Rank 排行榜
type Rank struct {
	ID                   int64                       `form:"id" json:"id"`
	SID                  int64                       `form:"sid" json:"sid"`
	SIDSource            int                         `form:"sid_source" json:"sid_source"`
	Ratio                string                      `form:"ratio" json:"ratio"`
	RatioStruct          *Ratio                      `json:"_"`
	RankType             int                         `form:"rank_type" json:"rank_type"`
	RankAttribute        int64                       `form:"rank_attribute" json:"rank_attribute"`
	Top                  int64                       `form:"top" json:"top"`
	IsAuto               int64                       `form:"is_auto" json:"is_auto"`
	IsShowScore          int64                       `form:"is_show_score" json:"is_show_score"`
	State                int64                       `form:"state" json:"state"`
	LastBatch            int64                       `form:"last_batch" json:"last_batch"`
	LastAttribute        int                         `form:"last_attribute" json:"last_attribute"`
	StatisticsTime       string                      `form:"statistics_time" json:"statistics_time"`
	StatisticsTimeStruct *StatisticsTime             `form:"_s" json:"_s"`
	Stime                go_common_library_time.Time `form:"stime" json:"stime"`
	Etime                go_common_library_time.Time `form:"etime" json:"etime"`
	Ctime                go_common_library_time.Time `json:"ctime"`
	Mtime                go_common_library_time.Time `json:"mtime"`
}

// Log 日志
type Log struct {
	ID            int64                       `form:"id" json:"id" validate:"min=1"`
	RankID        int64                       `form:"rank_id" json:"rank_id"`
	Batch         int64                       `form:"batch" json:"batch"`
	RankAttribute int64                       `form:"rank_attribute" json:"rank_attribute"`
	State         int64                       `form:"state" json:"state"`
	Ctime         go_common_library_time.Time `json:"ctime"`
	Mtime         go_common_library_time.Time `json:"mtime"`
}

// StatisticsTime 处理时间
type StatisticsTime struct {
	Day   *OperateTime `json:"day"`
	Week  *OperateTime `json:"week"`
	Month *OperateTime `json:"month"`
	All   *OperateTime `json:"all"`
}

// CronBatch ...
type CronBatch struct {
	Cron string
	Type uint
}

// IsAutoPublish 是否自动发布
func (r *Rank) IsAutoPublish() bool {
	if r.IsAuto == AutoPublishYes {
		return true
	}
	return false
}

// GetBatch 返回batch
func (r *Rank) GetBatch(stime int64) int {
	hour, _, _, _, day, year, month := getDay(stime)
	lastBatchStr := fmt.Sprintf("%d%02d%02d%02d", year, month, day, hour)
	lastBatch, _ := strconv.Atoi(lastBatchStr)
	return lastBatch

}

// GetRankAttributeType ...
func (r *Rank) GetRankAttributeType(attributeType uint) int {
	return (1 << attributeType)
}

// GetRankAttributeType ...
func GetRankAttributeType(attributeType uint) int {
	return (1 << attributeType)
}

// GetLastBatch 获得上次的batch
func (r *Rank) GetLastBatch(attributeType uint) (lastBatch int) {
	staticsTime := r.StatisticsTimeStruct
	var lastBatchStr string
	now := time.Now()
	if staticsTime == nil {
		return 0
	}
	switch attributeType {
	case RankAttributeAll:
		if staticsTime.All != nil && staticsTime.All.State == 1 {
			lastDay := now.AddDate(0, 0, -1)
			lastBatch = r.GetBatch(lastDay.Unix())
		}
	case RankAttributeDay:
		if staticsTime.Day != nil && staticsTime.Day.State == 1 {
			lastDay := now.AddDate(0, 0, -1)
			lastBatch = r.GetBatch(lastDay.Unix())
		}
	case RankAttributeWeek:
		if staticsTime.Week != nil && staticsTime.Day.State == 1 {
			lastDay := now.AddDate(0, 0, -7)
			lastBatch = r.GetBatch(lastDay.Unix())
		}
	case RankAttributeMonth:
		if staticsTime.Month != nil && staticsTime.Day.State == 1 {
			if staticsTime.Month.Type == StatisticsTimeTypeNature {
				hour, _, _, _, day, year, month := getDay(now.Unix())
				lastBatchStr = fmt.Sprintf("%d%02d%02d%02d", year, month-1, day, hour)
				lastBatch, _ = strconv.Atoi(lastBatchStr)
			} else {
				lastDay := now.AddDate(0, 0, -30)
				lastBatch = r.GetBatch(lastDay.Unix())
			}
		}
	}
	return

}
func getDay(stime int64) (int, int, int, int, int, int, int) {
	hour := time.Unix(stime, 0).Hour()
	minute := time.Unix(stime, 0).Minute()
	second := time.Unix(stime, 0).Second()
	week := int(time.Unix(stime, 0).Weekday())
	day := int(time.Unix(stime, 0).Day())
	year := time.Unix(stime, 0).Year()
	month := int(time.Unix(stime, 0).Month())
	return hour, minute, second, week, day, year, month
}

func (r *Rank) isRightHourCron(hour int) bool {
	if hour >= 0 && hour <= 23 {
		return true
	}
	return false
}

func (r *Rank) isRightMinuteCron(minute int) bool {
	if minute >= 0 && minute <= 59 {
		return true
	}
	return false
}

func (r *Rank) isRightSecondCron(second int) bool {
	if second >= 0 && second <= 59 {
		return true
	}
	return false
}

func (r *Rank) isRightWeekCron(week int) bool {
	if week >= 0 && week <= 6 {
		return true
	}
	return false
}

func (r *Rank) isRightDayCron(day int) bool {
	if day >= 1 && day <= 31 {
		return true
	}
	return false
}

func (r *Rank) isRightMonthCron(month int) bool {
	if month >= 1 && month <= 12 {
		return true
	}
	return false
}

// GetStatisticsCron 获取统计cron时间
func (r *Rank) GetStatisticsCron() (cron []*CronBatch) {
	cron = make([]*CronBatch, 0)
	now := time.Now().Unix()
	if r.StatisticsTimeStruct == nil {
		return
	}
	staticsTime := r.StatisticsTimeStruct
	if staticsTime.Day != nil && staticsTime.Day.State == 1 {
		day := staticsTime.Day
		hour, minute, second, _, _, _, _ := getDay(int64(day.Stime))
		cronDay := fmt.Sprintf("%d %d %d * * *", second, minute, hour)
		if int64(day.Stime) < now && r.isRightHourCron(hour) && r.isRightMinuteCron(minute) && r.isRightSecondCron(second) {
			cron = append(cron, &CronBatch{Cron: cronDay, Type: RankAttributeDay})
		}
	}
	if staticsTime.Week != nil && staticsTime.Week.State == 1 {
		week := staticsTime.Week
		w := int(time.Unix(int64(week.Stime), 0).Weekday())
		hour, minute, second, w, _, _, _ := getDay(int64(week.Stime))
		if week.Type == StatisticsTimeTypeNature {
			w = week.UpdateTime
		}
		cronDay := fmt.Sprintf("%d %d %d * * %d", second, minute, hour, w)
		if int64(week.Stime) < now && r.isRightHourCron(hour) && r.isRightMinuteCron(minute) && r.isRightSecondCron(second) && r.isRightWeekCron(w) {
			cron = append(cron, &CronBatch{Cron: cronDay, Type: RankAttributeWeek})
		}
	}

	if staticsTime.Month != nil && staticsTime.Month.State == 1 {
		month := staticsTime.Month
		hour, minute, second, _, d, _, _ := getDay(int64(month.Stime))
		if month.Type == StatisticsTimeTypeNature {
			d = month.UpdateTime
		}
		cronDay := fmt.Sprintf("%d %d %d %d * *", second, minute, hour, d)
		if int64(month.Stime) < now && r.isRightHourCron(hour) && r.isRightMinuteCron(minute) && r.isRightSecondCron(second) && r.isRightDayCron(d) {
			cron = append(cron, &CronBatch{Cron: cronDay, Type: RankAttributeMonth})
		}
	}
	if staticsTime.All != nil && staticsTime.All.State == 1 {
		all := staticsTime.All
		hour, minute, second, _, _, _, _ := getDay(int64(all.Stime))
		cronDay := fmt.Sprintf("%d %d %d * * *", second, minute, hour)
		if int64(all.Stime) < now && r.isRightHourCron(hour) && r.isRightMinuteCron(minute) && r.isRightSecondCron(second) {
			cron = append(cron, &CronBatch{Cron: cronDay, Type: RankAttributeAll})
		}
	}
	return
}

// OperateTime ...
type OperateTime struct {
	State      int                         `json:"state"`
	Stime      go_common_library_time.Time `json:"stime"`
	UpdateTime int                         `json:"update_time"`
	Type       int                         `json:"type"`
}

// IsAttributeALL 是否总榜
func (r *Rank) IsAttributeALL() bool {
	return ((r.RankAttribute >> RankAttributeAll) & int64(1)) == 1
}

// IsAttributeDay 是否日榜
func (r *Rank) IsAttributeDay() bool {
	return ((r.RankAttribute >> RankAttributeDay) & int64(1)) == 1
}

// IsAttributeWeek 是否周榜
func (r *Rank) IsAttributeWeek() bool {
	return ((r.RankAttribute >> RankAttributeWeek) & int64(1)) == 1
}

// IsAttributeMonth 是否月榜
func (r *Rank) IsAttributeMonth() bool {
	return ((r.RankAttribute >> RankAttributeMonth) & int64(1)) == 1
}

// Ratio ...
type Ratio struct {
	View    int64 `json:"view"`
	Danmaku int64 `json:"danmaku"`
	Reply   int64 `json:"reply"`
	Fav     int64 `json:"fav"`
	Coin    int64 `json:"coin"`
	Share   int64 `json:"share"`
	Like    int64 `json:"like"`
	Videos  int64 `json:"videos"`
	Revise  int64 `json:"revise"`
}

// Intervention 黑白名单
type Intervention struct {
	ID               int64                       ` json:"id"`
	OID              int64                       `json:"oid"`
	Score            int64                       `json:"score"`
	State            int                         `json:"state"`
	InterventionType int                         `json:"intervention_type"`
	ObjectType       int                         `json:"object_type"`
	Ctime            go_common_library_time.Time `json:"ctime"`
	Mtime            go_common_library_time.Time `json:"mtime"`
}

// OidResult 排行榜结果
type OidResult struct {
	ID            int64                       `json:"id"`
	OID           int64                       `json:"oid"`
	Rank          int64                       `json:"rank"`
	Score         int64                       `json:"score"`
	RankAttribute int                         `json:"rank_attribute"`
	State         int                         `json:"state"`
	Batch         int                         `json:"batch"`
	Remark        string                      `json:"remark"`
	Ctime         go_common_library_time.Time `json:"ctime"`
	Mtime         go_common_library_time.Time `json:"mtime"`
}

// Snapshot 快照
type Snapshot struct {
	ID            int64                       `json:"id"`
	MID           int64                       `json:"mid"`
	AID           int64                       `json:"aid"`
	TID           int64                       `json:"tid"`
	View          int64                       `json:"view"`
	Danmaku       int64                       `json:"danmaku"`
	Reply         int64                       `json:"reply"`
	Fav           int64                       `json:"fav"`
	Coin          int64                       `json:"coin"`
	Share         int64                       `json:"share"`
	Like          int64                       `json:"like"`
	Videos        int64                       `json:"videos"`
	Rank          int64                       `json:"rank"`
	RankAttribute int                         `json:"rank_attribute"`
	Score         int64                       `json:"score"`
	State         int                         `json:"state"`
	Batch         int                         `json:"batch"`
	Remark        string                      `json:"remark"`
	PubTime       int64                       `json:"pub_time"`
	ArcCtime      int64                       `json:"arc_ctime"`
	Ctime         go_common_library_time.Time `json:"ctime"`
	Mtime         go_common_library_time.Time `json:"mtime"`
}

// MidRank mid rank
type MidRank struct {
	Mid   int64       `json:"mid"`
	Score int64       `json:"score"`
	Rank  int64       `json:"rank"`
	Aids  []*AidScore `json:"aids"`
}

// AidScore ...
type AidScore struct {
	Aid   int64 `json:"aid"`
	Score int64 `json:"score"`
}
