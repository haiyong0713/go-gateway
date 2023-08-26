package show

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	xtime "time"

	"go-common/library/log"
	"go-common/library/time"
)

// const
const (
	_rtypeAV                 = "av"
	_seriePassed             = 2
	_serieNotOperated        = 0
	_archiveHonorGoodHisDesc = "入站必刷%d大视频"
	_archiveHonorWeeklyDesc  = "第%d期每周必看"
	_archiveHonorWeeklyType  = 2
	_archiveHonorGoodHisType = 1
)

// Merak is the config for wechat alert
type Merak struct {
	Names    []string
	Template string
	Title    string
	Cron     string
}

// Serie def.
type Serie struct {
	ID            int64
	Type          string
	Number        int64
	Status        int
	Stime         time.Time
	Etime         time.Time
	Pubtime       time.Time
	Subject       string
	ShareSubtitle string
	PushTitle     string
	PushSubtitle  string
}

// PushBody returns the push title of the serie
func (v *Serie) PushBody() string {
	return fmt.Sprintf("%d年第%d期 | ", v.Stime.Time().Year(), v.Number)
}

// MedialistTitle returns
func (v *Serie) MedialistTitle() string {
	return fmt.Sprintf("「%s %d年第%d期」", v.Subject, v.Stime.Time().Year(), v.Number)
}

// UUID returns the unique string of the serie to avoid push repeatedly
func (v *Serie) UUID(mid int64) string {
	var b bytes.Buffer
	b.WriteString(strconv.FormatInt(mid, 10))
	b.WriteString(strconv.FormatInt(v.ID, 10))
	b.WriteString(v.Type)
	b.WriteString(strconv.FormatInt(v.Number, 10))
	mh := md5.Sum(b.Bytes())
	return hex.EncodeToString(mh[:])
}

// CanRecovery checks whether the serie can be recovered
func (v *Serie) CanRecovery() bool {
	return v.Status == _serieNotOperated
}

// Passed def.
func (v *Serie) Passed() bool {
	return v.Status == _seriePassed
}

// Passed def.
func (v *BinlogSerie) Passed() bool {
	return v.Status == _seriePassed
}

// SerieRes def.
type SerieRes struct {
	RID   int64
	Rtype string
}

// IsArc def.
func (v *SerieRes) IsArc() bool {
	return v.Rtype == _rtypeAV
}

// FromAv def.
func (v *SerieRes) FromAv(aid int64) {
	v.RID = aid
	v.Rtype = _rtypeAV
}

// AIPopular databus .
type AIPopular struct {
	Route string `json:"route"`
}

// AIWeeklySel def.
type AIWeeklySel struct {
	AID int64 `json:"aid"`
}

// AIWeeklyItems
type AIWeeklyItems struct {
	List []*AIWeeklySel `json:"list"`
}

// DatabusRes is the result of databus binlog message
type DatabusRes struct {
	Action string `json:"action"`
	Table  string `json:"table"`
}

// SerieDatabus .
type SerieDatabus struct {
	Old *BinlogSerie `json:"old"`
	New *BinlogSerie `json:"new"`
}

// BinlogSerie def.
type BinlogSerie struct {
	Type    string `json:"type"`
	Number  int64  `json:"number"`
	Deleted int    `json:"deleted"`
	Status  int    `json:"status"`
	ID      int64  `json:"id"`
	MediaID int64  `json:"media_id"`
	Pubtime string `json:"pubtime"`
}

// PubtimeValue gets the int64 format of the pubtime
func (v *BinlogSerie) PubtimeValue() (pubtime int64, err error) {
	local, _ := xtime.LoadLocation("Local")
	var timeValue xtime.Time
	timeValue, err = xtime.ParseInLocation("2006-01-02 15:04:05", v.Pubtime, local)
	if err != nil {
		log.Warn("TimeTrans %s, Error %v", v.Pubtime, err)
		return
	}
	if pubtime = timeValue.Unix(); pubtime < 1 {
		err = fmt.Errorf("time %s transform error", v.Pubtime)
		log.Warn("TimeTrans Err %v", err)
	}
	return
}

// IsPassed def.
func (v *BinlogSerie) IsPassed() bool {
	return v.Deleted == 0 && v.Status == 2
}

type HonorMsg struct {
	Action string `json:"action"`
	Aid    int64  `json:"aid"`
	Type   int    `json:"type"`
	Url    string `json:"url"`
	Desc   string `json:"desc"`
	NaUrl  string `json:"na_url"`
}

type OTTSeriesMsg struct {
	Action string `json:"action"`
	Number int64  `json:"number"`
	Aid    int64  `json:"aid"`
}

func (v *HonorMsg) FromWeeklySelected(aid, number int64, url string, action string, naUrl string) {
	v.Aid = aid
	v.Type = _archiveHonorWeeklyType
	v.Desc = fmt.Sprintf(_archiveHonorWeeklyDesc, number)
	v.Url = url
	v.Action = action
	v.NaUrl = naUrl
}

func (v *HonorMsg) FromGoodHistory(aid int64, action string, count int64, url string) {
	v.Aid = aid
	v.Action = action
	v.Url = url
	if count > 0 {
		v.Desc = fmt.Sprintf(_archiveHonorGoodHisDesc, count)
	}
	v.Type = _archiveHonorGoodHisType
}

// SerieFull is the full structure of one serie in MC
type SerieFull struct {
	Config *SerieConfig   `json:"config"`
	List   []*SelectedRes `json:"list"`
}

// SerieConfig is the structure in the selected series API
type SerieConfig struct {
	SerieCore
	Label         string `json:"label"`
	Hint          string `json:"hint"`
	Color         int    `json:"color"`
	Cover         string `json:"cover"`
	ShareTitle    string `json:"share_title"`
	ShareSubtitle string `json:"share_subtitle"`
	MediaID       int64  `json:"media_id"` // 播单ID
}

// SerieCore is the core fields of selected serie
type SerieCore struct {
	ID      int64     `json:"-"`
	Type    string    `json:"-"`
	Number  int64     `json:"number"`
	Subject string    `json:"subject"`
	Stime   time.Time `json:"-"`
	Etime   time.Time `json:"-"`
	Status  int       `json:"status"`
}

// SelectedRes represents selected resources
type SelectedRes struct {
	RID        int64  `json:"rid"`
	Rtype      string `json:"rtype"`
	SerieID    int64  `json:"serie_id"`
	Position   int    `json:"position"`
	RcmdReason string `json:"rcmd_reason"`
}

// IsArc def.
func (v *SelectedRes) IsArc() bool {
	return v.Rtype == _rtypeAV
}
