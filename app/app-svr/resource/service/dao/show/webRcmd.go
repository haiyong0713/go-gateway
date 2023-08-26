package show

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/xstr"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_webRcmd        = "SELECT id,card_type,card_value,`partition`,tag,avid,priority,`order` FROM web_rcmd WHERE `check`=2 AND deleted=0 AND stime<? AND etime>? ORDER BY id DESC"
	_webRcmdCard    = "SELECT id,type,title,`desc`,cover,re_type,re_value FROM web_rcmd_card WHERE deleted=0 ORDER BY id DESC"
	_webSpecialCard = "SELECT id,type,title,`desc`,cover,re_type,re_value,person,ctime,mtime FROM web_rcmd_card WHERE deleted=0 ORDER BY id DESC"
)

// WebRcmd get web rcmd.
func (d *Dao) WebRcmd(c context.Context) (rcmds []*pb.WebRcmd, err error) {
	now := time.Now()
	rows, err := d.db.Query(c, _webRcmd, now, now)
	if err != nil {
		log.Error("d.WebRcmd query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tag, avid, partition string
		rcmd := &pb.WebRcmd{}
		if err = rows.Scan(&rcmd.ID, &rcmd.CardType, &rcmd.CardValue, &partition, &tag, &avid, &rcmd.Priority, &rcmd.Order); err != nil {
			log.Error("WebRcmd rows.Scan err (%v)", err)
			return
		}
		rcmd.Partition, _ = xstr.SplitInts(partition)
		rcmd.Tag, _ = xstr.SplitInts(tag)
		rcmd.AvID, _ = xstr.SplitInts(avid)
		rcmds = append(rcmds, rcmd)
	}
	err = rows.Err()
	return
}

// WebRcmdCard get web rcmd card.
func (d *Dao) WebRcmdCard(c context.Context) (rcs []*pb.WebRcmdCard, err error) {
	rows, err := d.db.Query(c, _webRcmdCard)
	if err != nil {
		log.Error("d.WebRcmdCard query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rc := &pb.WebRcmdCard{}
		if err = rows.Scan(&rc.ID, &rc.Type, &rc.Title, &rc.Desc, &rc.Cover, &rc.ReType, &rc.ReValue); err != nil {
			log.Error("WebRcmdCard rows.Scan err (%v)", err)
			return
		}
		rcs = append(rcs, rc)
	}
	err = rows.Err()
	return
}

// WebSpecialCard get web rcmd card.
func (d *Dao) WebSpecialCard(c context.Context) (rcs []*pb2.WebSpecialCard, err error) {
	rows, err := d.db.Query(c, _webSpecialCard)
	if err != nil {
		log.Error("d.WebSpecialCard query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rc := &model.WebRcmdCard{}
		if err = rows.Scan(&rc.ID, &rc.Type, &rc.Title, &rc.Desc, &rc.Cover, &rc.ReType, &rc.ReValue, &rc.Person, &rc.Ctime, &rc.Mtime); err != nil {
			log.Error("WebSpecialCard rows.Scan err (%v)", err)
			return
		}
		rcs = append(rcs, &pb2.WebSpecialCard{
			Id:      rc.ID,
			Type:    rc.Type,
			Title:   rc.Title,
			Desc:    rc.Desc,
			Cover:   rc.Cover,
			ReType:  rc.ReType,
			ReValue: rc.ReValue,
			Person:  rc.Person,
			Ctime:   int64(rc.Ctime),
			Mtime:   int64(rc.Mtime),
		})
	}
	err = rows.Err()
	return
}
