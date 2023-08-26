package http

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/component/boss"
	model "go-gateway/app/web-svr/activity/admin/model/lottery"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"
)

func lotteryMidAddTimes(ctx *bm.Context) {

	v := new(struct {
		Mid int64  `form:"mid" validate:"min=1"`
		Sid string `form:"sid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(lotterySrv.MidTimes(ctx, v.Mid, v.Sid))
}

func list(c *bm.Context) {
	var request = &model.ListParam{}
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.List(c, request))
}

func listDraft(c *bm.Context) {
	var request = &model.ListParam{}
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.ListDraft(c, request))
}

func add(c *bm.Context) {
	var request = &model.AddParam{}
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.Add(c, request, userName))
}

func addDraft(c *bm.Context) {
	var request = &model.AddParam{}
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.AddDraft(c, request, userName))
}

func deleteLottery(c *bm.Context) {
	var (
		params = c.Request.Form
		id     int64
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "id不能为空"))
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.Delete(c, id, userName))
}

func unsedLottery(c *bm.Context) {
	var request = &model.UsedParam{}
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.LotteryRecord(c, request.SID, request.MID))
}

func detail(c *bm.Context) {
	var (
		params = c.Request.Form
		sid    string
	)
	if sid = params.Get("sid"); sid == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "sid不能为空"))
		return
	}
	c.JSON(lotterySrv.Detail(c, sid))
}

func detailDraft(c *bm.Context) {
	var (
		params = c.Request.Form
		sid    string
	)
	if sid = params.Get("sid"); sid == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "sid不能为空"))
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(lotterySrv.DetailDraft(c, sid, userName))
}

func edit(c *bm.Context) {
	var (
		request = &model.EditParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.Edit(c, request, cookie, userName))
}

func editDraft(c *bm.Context) {
	var (
		request = &model.EditParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.EditDraft(c, request, cookie, userName))
}

func giftAdd(c *bm.Context) {
	var (
		request = &model.GiftAddParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.GiftAdd(c, request, cookie, userName))
}

func giftAddDraft(c *bm.Context) {
	var (
		request = &model.GiftAddParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.GiftAddDraft(c, request, cookie, userName))
}

func giftEdit(c *bm.Context) {
	var (
		request = &model.GiftEditParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.GiftEdit(c, request, cookie, userName))
}

func giftEditDraft(c *bm.Context) {
	var (
		request = &model.GiftEditParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.GiftEditDraft(c, request, cookie, userName))
}

func lotteryDraftAudit(c *bm.Context) {
	var (
		request = &model.AuditParam{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.Audit(c, request.SID, request.State, userName, request.RejectReason))
}

func memberGroupEdit(c *bm.Context) {
	var (
		request = &model.MemberGroupEditParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.MemberGroupEdit(c, request, cookie, userName))
}

func memberGroupDraftEdit(c *bm.Context) {
	var (
		request = &model.MemberGroupEditParam{}
		cookie  = c.Request.Header.Get("cookie")
	)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.MemberGroupDraftEdit(c, request, cookie, userName))
}
func giftList(c *bm.Context) {
	var (
		request = &model.GiftListParam{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.GiftList(c, request))
}

func giftListDraft(c *bm.Context) {
	var (
		request = &model.GiftListParam{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.GiftListDraft(c, request))
}

func memberGroupList(c *bm.Context) {
	var (
		request = &model.MemberGroupListParam{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.MemberGroupList(c, request))
}

func memberGroupListDraft(c *bm.Context) {
	var (
		request = &model.MemberGroupListParam{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.MemberGroupListDraft(c, request))
}

func giftWinList(c *bm.Context) {
	var (
		request = &model.GiftWinListParam{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.GiftWinList(c, request))
}

func giftUpload(c *bm.Context) {
	var (
		aid int64
		sid string
		err error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if aid, err = strconv.ParseInt(c.Request.Form.Get("aid"), 10, 64); err != nil {
		log.Errorc(c, "strconv.ParseInt() failed. error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "参数错误"))
		return
	}
	if sid = c.Request.Form.Get("sid"); sid == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "sid不可为空"))
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "csv文件解析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件解析失败"))
		return
	}
	//reader := csv.NewReader(bom.NewReader(file))
	reader := csv.NewReader(file)
	keys := []string{}
	records, err := reader.ReadAll()
	if err != nil {
		log.Errorc(c, "csv文件读取析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件读取失败"))
		return
	}
	for _, line := range records {
		if len(line) <= 0 {
			continue
		}
		keys = append(keys, line[0])
	}
	key := model.GetUploadKey(sid, aid)
	if err = lotterySrv.UpdUploadStatus(c, model.UploadStart, aid); err != nil {
		log.Errorc(c, "数据库更新失败，error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.ServerErr, "数据库更新失败"))
		return
	}
	lotterySrv.UploadInfo[key] = &model.UploadInfo{Status: model.UploadStart}
	go lotterySrv.GiftUpload(context.Background(), keys, aid, sid, userName)
	c.JSON(nil, err)
}

func addTimesBatchRetry(c *blademaster.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, lotterySrv.AddtimesRetry(c, v.ID, userName))

}

func addTimesBatch(c *blademaster.Context) {

	v := new(struct {
		Sid string `form:"sid" validate:"required"`
		Cid int64  `form:"cid" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "importDetailCSV upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Errorc(c, "importDetailCSV ioutil.ReadAll err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Errorc(c, "importDetailCSV r.ReadAll() err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var args []int64
Loop:
	for i, row := range records {
		if i == 0 {
			continue
		}
		// import csv state online
		var arg int64
		for field, value := range row {
			value = strings.TrimSpace(value)
			switch field {
			case 0:
				if value == "" {
					log.Warn("importDetailCSV name provinceID(%s)", value)
					continue Loop
				}
				mid, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					// continue Loop
				}
				arg = mid

			}

		}
		args = append(args, arg)
	}
	if len(args) == 0 {
		log.Errorc(c, "importDetailCSV args no after filter")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	go func() {
		ctx := context.Background()
		lotterySrv.DoBatchMidAddTimes(ctx, args, userName, v.Cid, v.Sid)
	}()
	c.JSON(nil, nil)
}

func addTimesMidList(c *bm.Context) {
	v := new(struct {
		Mid int64  `form:"mid" validate:"min=1"`
		Sid string `form:"sid" validate:"min=1"`
		Pn  int    `form:"pn" default:"1" validate:"min=1"`
		Ps  int    `form:"ps" default:"10" validate:"min=1,max=20"`
		Cid int64  `form:"cid"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(lotterySrv.MidAddTimesLog(c, v.Mid, v.Sid, v.Cid, v.Pn, v.Ps))
}

func addTimesLogList(c *bm.Context) {
	var (
		request = &model.BatchAddTimesParams{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(lotterySrv.AddTimesBatchList(c, request))
}

func giftExport(c *bm.Context) {
	var (
		param   = c.Request.Form
		aid     int64
		sid     string
		err     error
		infoStr [][]string
	)
	if aid, err = strconv.ParseInt(param.Get("aid"), 10, 64); err != nil {
		log.Errorc(c, "strconv.ParseInt(%v) failed. error(%v)", param.Get("aid"), err)
		c.JSON(nil, err)
		return
	}
	if sid = param.Get("sid"); sid == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "sid不能为空"))
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	go func() {
		c := context.Background()
		if infoStr, err = lotterySrv.GiftExport(c, aid, sid); err != nil {
			log.Errorc(c, "lotterySrv.GiftExport(aid:%v, sid:%v) failed. error(%v)", aid, sid, err)
			exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("抽奖%s导出数据失败，请重试或者在救火大队反馈跟进。", sid))
			return
		}
		categoryHeader := []string{"Mid", "姓名", "奖品名称", "奖品类型", "收货地址", "联系方式", "获奖时间"}
		b := &bytes.Buffer{}
		b.WriteString("\xEF\xBB\xBF")
		wr := csv.NewWriter(b)
		_ = wr.Write(categoryHeader)
		for i := 0; i < len(infoStr); i++ {
			wr.Write(infoStr[i])
		}
		wr.Flush()
		url, err := boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("lotterydata/%s/中奖名单.%d.%s.csv", time.Now().Format("20060102150405"), aid, sid), b)
		if err != nil {
			log.Error("lotterySrv.GiftExportAll(sid:%v) failed. error(%v)", sid, err)
			exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("抽奖%s导出数据失败，请重试或者在救火大队反馈跟进。", sid))
			return
		}
		exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("抽奖%s导出数据成功，下载链接:%s", sid, url))
	}()
	c.JSON(nil, nil)
}

func fixLotteryGiftTask(c *bm.Context) {
	c.JSON(nil, lotterySrv.FixLotteryGiftTask(c))
}

func vipCheck(c *bm.Context) {
	var (
		params = c.Request.Form
		vip    string
		cookie = c.Request.Header.Get("cookie")
	)
	if vip = params.Get("vipID"); vip == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "vipID必传，请检查"))
		return
	}
	c.JSON(lotterySrv.VipCheck(c, vip, cookie))
}

func batchAddTimes(c *bm.Context) {
	var (
		err  error
		data []byte
	)
	v := new(struct {
		ID    string `form:"id"`
		Times int64  `form:"times" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("importDetailCSV upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("batchAddTimes ioutil.ReadAll err(%v)", err)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("batchAddTimes r.ReadAll() err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if l := len(records); l > 1000000 || l <= 1 {
		log.Error("batchAddTimes len(%d) err", l)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	mids := make([]int64, 0)
	for _, row := range records {
		for field, value := range row {
			switch field {
			case 0:
				mid, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					log.Warn("batchAddTimes strconv.ParseInt(%s) error(%v)", value, err)
					continue
				}
				mids = append(mids, mid)
			}
		}
	}
	c.JSON(nil, lotterySrv.BatchAddTimes(c, v.ID, v.Times, mids))
}

func giftExportAll(c *bm.Context) {
	var (
		param   = c.Request.Form
		sid     string
		err     error
		infoStr [][]string
	)
	if sid = param.Get("sid"); sid == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "sid不能为空"))
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	go func() {
		c := context.Background()
		if infoStr, err = lotterySrv.GiftExportAll(c, sid); err != nil {
			log.Error("lotterySrv.GiftExportAll(sid:%v) failed. error(%v)", sid, err)
			exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("抽奖%s导出数据失败，请重试或者在救火大队反馈跟进。", sid))
			return
		}
		categoryHeader := []string{"Mid", "姓名", "奖品ID", "奖品名称", "奖品类型", "收货地址", "联系方式", "获奖时间"}
		b := &bytes.Buffer{}
		b.WriteString("\xEF\xBB\xBF")
		wr := csv.NewWriter(b)
		_ = wr.Write(categoryHeader)
		for i := 0; i < len(infoStr); i++ {
			wr.Write(infoStr[i])
		}
		wr.Flush()
		url, err := boss.Client.UploadObject(c, boss.Bucket, fmt.Sprintf("lotterydata/%s/全部中奖名单.%s.csv", time.Now().Format("20060102150405"), sid), b)
		if err != nil {
			log.Error("lotterySrv.GiftExportAll(sid:%v) failed. error(%v)", sid, err)
			exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("抽奖%s导出数据失败，请重试或者在救火大队反馈跟进。", sid))
			return
		}
		exporttask.SendWeChatTextMessage(c, []string{userName}, fmt.Sprintf("抽奖%s导出数据成功，下载链接:%s", sid, url))
	}()
	c.JSON(nil, nil)
}

func wxLotteryLog(c *bm.Context) {
	v := new(struct {
		Mid      int64 `form:"mid"`
		GiftType int64 `form:"gift_type"`
		Pn       int64 `form:"pn" default:"1" validate:"min=1"`
		Ps       int64 `form:"ps" default:"10" validate:"min=1,max=20"`
	})
	if err := c.Bind(v); err != nil {
		panic(err)
	}
	list, count, err := lotterySrv.WxLotteryLog(c, v.Mid, v.GiftType, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}
