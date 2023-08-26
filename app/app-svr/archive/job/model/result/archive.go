package result

import (
	"database/sql/driver"
	"sync"
	"time"

	"go-gateway/app/app-svr/archive/service/api"
)

type ArchiveUpInfo struct {
	Table  string
	Action string
	Nw     *api.Arc
	Old    *api.Arc
}

type ResultDelay struct {
	Lock sync.RWMutex
	AIDs map[int64]struct{}
}

// Result archive result
type Archive struct {
	ID          int64     `json:"id"`
	AID         int64     `json:"aid"`
	Mid         int64     `json:"mid"`
	TypeID      int32     `json:"typeid"`
	Videos      int       `json:"videos"`
	Title       string    `json:"title"`
	Cover       string    `json:"cover"`
	Content     string    `json:"content"`
	Duration    int       `json:"duration"`
	Attribute   int32     `json:"attribute"`
	Copyright   int8      `json:"copyright"`
	Access      int       `json:"access"`
	PubTime     wocaoTime `json:"pubtime"`
	CTime       wocaoTime `json:"ctime"`
	MTime       wocaoTime `json:"mtime"`
	State       int       `json:"state"`
	MissionID   int64     `json:"mission_id"`
	OrderID     int64     `json:"order_id"`
	RedirectURL string    `json:"redirect_url"`
	Forward     int64     `json:"forward"`
	Dynamic     string    `json:"dynamic"`
	SeasonID    int64     `json:"season_id"`
	AttributeV2 int64     `json:"attribute_v2"`
	UpFromV2    int32     `json:"up_from"`
	FirstFrame  string    `json:"first_frame"`
}

type wocaoTime string

// Scan scan time.
func (jt *wocaoTime) Scan(src interface{}) (err error) {
	switch sc := src.(type) {
	case time.Time:
		*jt = wocaoTime(sc.Format("2006-01-02 15:04:05"))
	case string:
		*jt = wocaoTime(sc)
	}
	return
}

// Value get time value.
func (jt wocaoTime) Value() (driver.Value, error) {
	return time.Parse("2006-01-02 15:04:05", string(jt))
}

func FromNotifyArc(arc *api.Arc) (notifyArc *Archive) {
	if arc == nil {
		return
	}
	notifyArc = &Archive{
		AID:         arc.Aid,
		Mid:         arc.Author.Mid,
		TypeID:      arc.TypeID,
		Videos:      int(arc.Videos),
		Title:       arc.Title,
		Cover:       arc.Pic,
		Content:     arc.Desc,
		Duration:    int(arc.Duration),
		Attribute:   arc.Attribute,
		Copyright:   int8(arc.Copyright),
		Access:      int(arc.Access),
		PubTime:     wocaoTime(arc.PubDate.Time().Format("2006-01-02 15:04:05")),
		CTime:       wocaoTime(arc.Ctime.Time().Format("2006-01-02 15:04:05")),
		State:       int(arc.State),
		MissionID:   arc.MissionID,
		OrderID:     arc.OrderID,
		RedirectURL: arc.RedirectURL,
		Forward:     arc.Forward,
		Dynamic:     arc.Dynamic,
		SeasonID:    arc.SeasonID,
		AttributeV2: arc.AttributeV2,
		UpFromV2:    arc.UpFromV2,
		FirstFrame:  arc.FirstFrame,
	}
	return notifyArc
}
