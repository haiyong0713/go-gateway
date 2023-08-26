package recommend

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"

	"github.com/pkg/errors"
)

// PubDislike is.
func (d *Dao) PubDislike(c context.Context, buvid, gt string, id, mid, reasonID, cmreasonID, feedbackID, upperID, rid,
	tagID int64, adcb, fromspmid, frommodule string, now time.Time, disableRcmd, fromAvid, fromType, materialId int64,
	reportData string) (err error) {
	return d.pub(c, buvid, gt, id, mid, reasonID, cmreasonID, feedbackID, upperID, rid, tagID, adcb, fromspmid,
		frommodule, 1, now, disableRcmd, fromAvid, fromType, materialId, reportData)
}

// PubDislikeCancel is.
func (d *Dao) PubDislikeCancel(c context.Context, buvid, gt string, id, mid, reasonID, cmreasonID, feedbackID, upperID,
	rid, tagID int64, adcb, fromspmid, frommodule string, now time.Time, disableRcmd, fromAvid, fromType,
	materialId int64, reportData string) (err error) {
	return d.pub(c, buvid, gt, id, mid, reasonID, cmreasonID, feedbackID, upperID, rid, tagID, adcb, fromspmid,
		frommodule, 2, now, disableRcmd, fromAvid, fromType, materialId, reportData)
}

// nolint: bilirailguncheck
func (d *Dao) pub(c context.Context, buvid, gt string, id, mid, reasonID, cmreasonID, feedbackID, upperID, rid, tagID int64,
	adcb, fromspmid, frommodule string, state int8, now time.Time, closeRcmd, fromAvid, fromType, materialId int64,
	reportData string) (err error) {
	key := strconv.FormatInt(mid, 10)
	msg := struct {
		Buvid       string `json:"buvid"`
		Goto        string `json:"goto"`
		ID          int64  `json:"id"`
		Mid         int64  `json:"mid"`
		ReasonID    int64  `json:"reason_id"`
		CMReasonID  int64  `json:"cm_reason_id"`
		FeedbackID  int64  `json:"feedback_id"`
		UpperID     int64  `json:"upper_id"`
		Rid         int64  `json:"rid"`
		TagID       int64  `json:"tag_id"`
		ADCB        string `json:"ad_cb"`
		State       int8   `json:"state"`
		Time        int64  `json:"time"`
		FromSpmid   string `json:"from_spmid"`
		FromModule  string `json:"from_module"`
		DisableRcmd int64  `json:"disable_rcmd"`
		FromAvid    int64  `json:"from_avid"`
		FromType    int64  `json:"from_type"`
		MaterialId  int64  `json:"material_id"`
		ReportData  string `json:"report_data"`
	}{Buvid: buvid, Goto: gt, ID: id, Mid: mid, ReasonID: reasonID, CMReasonID: cmreasonID, FeedbackID: feedbackID,
		UpperID: upperID, Rid: rid, TagID: tagID, ADCB: adcb, State: state, Time: now.Unix(), FromSpmid: fromspmid,
		FromModule: frommodule, DisableRcmd: closeRcmd, FromAvid: fromAvid, FromType: fromType, MaterialId: materialId,
		ReportData: reportData}
	if err = d.databus.Send(c, key, msg); err != nil {
		err = errors.Wrapf(err, "%s %v", key, msg)
		return
	}
	log.Info("d.dataBus.Pub(%s,%v)", key, msg)
	return
}
