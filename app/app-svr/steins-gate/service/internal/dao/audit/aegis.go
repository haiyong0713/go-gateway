package audit

import (
	"context"
	"strconv"
	"time"

	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/steins"
	aegis "go-main/app/archive/aegis/admin/server/databus"

	"go-common/library/log"
)

const _aegisBusinessID = 12

// SendAegisMsg .
func (d *Dao) AddAegisMsg(c context.Context, graphID, mid, aid, state int64, title, diffMsg string) (err error) {
	m := &aegis.AddInfo{
		BusinessID: _aegisBusinessID,
		NetID:      _aegisBusinessID,
		OID:        strconv.FormatInt(graphID, 10),
		MID:        mid,
		Content:    title,
		Extra1:     state,
		Extra2:     aid,
		MetaData:   diffMsg,
		OCtime:     time.Now(),
	}
	if err = aegis.Add(m); err != nil {
		log.Error("SaveGraph aegis.Add(%v) error(%v)", m, err)
		err = ecode.GraphSendAuditErr
		return
	}
	log.Info("SaveGraph aegis.Add(%v) ok", m)
	return
}

// CancelAegisMsg .
func (d *Dao) CancelAegisMsg(c context.Context, graphID int64, reason string) (err error) {
	m := &aegis.CancelInfo{
		BusinessID: _aegisBusinessID,
		Oids:       []string{strconv.FormatInt(graphID, 10)},
		Reason:     reason,
	}
	err = d.cache.Do(c, func(ctx context.Context) {
		if e := steins.Retry(func() (err error) { // retry 10 times
			if err = aegis.Cancel(m); err != nil {
				log.Warn("SaveGraph aegis.Cancel(%v) error(%v)", m, err)
			} else {
				log.Info("SaveGraph aegis.Cancel(%v) ok", m)
			}
			return
		}, 10, steins.RetrySpan); e != nil {
			log.Error("SaveGraph aegis.Cancel(%v) error(%v)", m, e)
		}
	})
	return

}
