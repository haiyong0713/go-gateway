package guess

import (
	"fmt"

	xtime "go-common/library/time"
)

// MainID .
type MainID struct {
	ID        int64 `json:"id"`
	OID       int64 `json:"oid"`
	IsDeleted int64 `json:"is_deleted"`
	Business  int64 `json:"business"`
}

// MainRes main details.
type MainRes struct {
	*MainGuess
	Details []DetailGuess
}

// MainDetail .
type MainDetail struct {
	ID         int64   `json:"id"`
	Business   int64   `json:"business"`
	Oid        int64   `json:"oid"`
	Title      string  `json:"title"`
	StakeType  int64   `json:"stake_type"`
	MaxStake   int64   `json:"max_stake"`
	ResultID   int64   `json:"result_id"`
	GuessCount int64   `json:"guess_count"`
	Stime      int64   `json:"stime"`
	Etime      int64   `json:"etime"`
	DetailID   int64   `json:"detail_id"`
	Option     string  `json:"option"`
	Odds       float32 `json:"odds"`
	TotalStake int64   `json:"total_stake"`
	IsDeleted  int64   `json:"is_deleted"`

	TemplateType int64 `json:"template_type"`
}

// MainGuess main guess.
type MainGuess struct {
	ID         int64      `json:"id"`
	Business   int64      `json:"business"`
	Oid        int64      `json:"oid"`
	Title      string     `json:"title"`
	StakeType  int64      `json:"stake_type"`
	MaxStake   int64      `json:"max_stake"`
	ResultID   int64      `json:"result_id"`
	GuessCount int64      `json:"guess_count"`
	IsDeleted  int64      `json:"is_deleted"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
	Stime      int64      `json:"stime"`
	Etime      int64      `json:"etime"`

	TemplateType int64  `json:"template_type"`
	RightOption  string `json:"right_option"`
}

// DetailGuess .
type DetailGuess struct {
	ID         int64      `json:"id"`
	MainID     int64      `json:"main_id"`
	Option     string     `json:"option"`
	Odds       float32    `json:"odds"`
	TotalStake int64      `json:"total_stake"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
}

// UserGuessLog user guess stat.
type UserGuessLog struct {
	ID        int64      `json:"id"`
	Mid       int64      `json:"mid"`
	MainID    int64      `json:"main_id"`
	DetailID  int64      `json:"detail_id"`
	StakeType int64      `json:"stake_type"`
	Stake     int64      `json:"stake"`
	Income    float32    `json:"income"`
	Status    int64      `json:"status"`
	Ctime     xtime.Time `json:"ctime"`
}

func NewMainRes() *MainRes {
	newOne := new(MainRes)
	{
		newOne.MainGuess = new(MainGuess)
		newOne.Details = make([]DetailGuess, 0)
	}

	return newOne
}

func (info *MainRes) DeepCopy() *MainRes {
	newOne := NewMainRes()
	if info.MainGuess != nil {
		*newOne.MainGuess = *info.MainGuess
	}

	if info.Details != nil {
		newOne.Details = info.Details[:]
	}

	return newOne
}

func (info *MainRes) GenHotMapKey() string {
	return GenHotMapKeyByOIDAndBusiness(info.Oid, info.Business)
}

func GenMainResByDetail(info *MainRes, v *MainDetail) *MainRes {
	tmpGuessDetail := DetailGuess{
		ID:         v.DetailID,
		Option:     v.Option,
		Odds:       v.Odds,
		TotalStake: v.TotalStake,
	}

	if info == nil || info.ID == 0 {
		info = &MainRes{
			MainGuess: &MainGuess{
				ID:           v.ID,
				Business:     v.Business,
				Oid:          v.Oid,
				Title:        v.Title,
				StakeType:    v.StakeType,
				MaxStake:     v.MaxStake,
				ResultID:     v.ResultID,
				GuessCount:   v.GuessCount,
				Stime:        v.Stime,
				Etime:        v.Etime,
				IsDeleted:    v.IsDeleted,
				TemplateType: v.TemplateType,
			},
			Details: []DetailGuess{},
		}
	}

	if v.ResultID == v.DetailID {
		info.RightOption = v.Option
	}

	info.Details = append(info.Details, tmpGuessDetail)

	return info
}

func (info *MainID) DeepCopy() *MainID {
	newOne := new(MainID)
	{
		newOne.ID = info.ID
		newOne.Business = info.Business
		newOne.OID = info.OID
		newOne.IsDeleted = info.IsDeleted
	}

	return newOne
}

func GenHotMapKeyByOIDAndBusiness(oID, business int64) string {
	return fmt.Sprintf("%v_%v", oID, business)
}
