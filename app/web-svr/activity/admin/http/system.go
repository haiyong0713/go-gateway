package http

import (
	"encoding/csv"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/web-svr/activity/admin/model/system"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
	"io/ioutil"
	"strconv"
	"strings"
)

func importSignVipList(ctx *bm.Context) {
	var (
		err  error
		data []byte
	)
	v := new(struct {
		AID int64 `form:"aid" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Errorc(ctx, "importSignVipList upload err(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Errorc(ctx, "importSignVipList ioutil.ReadAll err(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("importSignVipList r.ReadAll() err(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	var uids []string
	var seats []*model.UIDSeat
	for _, record := range records {
		// uid
		uid := strings.Trim(strings.TrimSpace(record[0]), "\ufeff")
		if uid == "" {
			continue
		}
		uids = append(uids, uid)
		// 桌号
		if len(record) > 1 {
			seat := strings.TrimSpace(record[1])
			if seat != "" {
				// 收集用户uid和桌号
				seats = append(seats, &model.UIDSeat{UID: uid, AID: v.AID, Content: seat})
			}
		}
	}

	ctx.JSON(nil, systemSrv.ImportSignVipList(ctx, v.AID, uids, seats))
}

func signList(ctx *bm.Context) {
	v := new(struct {
		AID  int64 `form:"aid" validate:"required,min=1"`
		Page int64 `form:"page" validate:"required,min=1"`
		Size int64 `form:"size" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(systemSrv.GetSignList(ctx, v.AID, v.Page, v.Size))
}

func signVipList(ctx *bm.Context) {
	v := new(struct {
		AID  int64 `form:"aid" validate:"required,min=1"`
		Page int64 `form:"page" validate:"required,min=1"`
		Size int64 `form:"size" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(systemSrv.GetSignVipList(ctx, v.AID, v.Page, v.Size))
}

func signUser(ctx *bm.Context) {
	v := new(struct {
		AID int64  `form:"aid" validate:"required,min=1"`
		UID string `form:"uid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(nil, systemSrv.SignUser(ctx, v.AID, v.UID))
}

func exportSignList(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	res, err := systemSrv.GetSignList(ctx, v.AID, 1, 0)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "活动" + strconv.FormatInt(v.AID, 10) + "签到人员明细.csv",
		Header:         []string{"uid", "昵称", "姓名", "签到时间", "定位信息"},
		Result:         systemSrv.ExportSignList(ctx, res),
	}

	exportCsv(ctx, exportCsvParam)
}

func exportSignVipList(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	res, err := systemSrv.GetSignVipList(ctx, v.AID, 1, 0)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "活动" + strconv.FormatInt(v.AID, 10) + "白名单人员签到明细.csv",
		Header:         []string{"uid", "昵称", "姓名", "是否签到", "签到时间", "定位信息"},
		Result:         systemSrv.ExportSignVipList(ctx, res),
	}

	exportCsv(ctx, exportCsvParam)
}

func actAdd(ctx *bm.Context) {
	v := new(model.SystemActAddArgs)
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(systemSrv.SystemActAdd(ctx, v))
}

func actEdit(ctx *bm.Context) {
	v := new(model.SystemActEditArgs)
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(nil, systemSrv.SystemActEdit(ctx, v))
}

func actState(ctx *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"required"`
		State int64 `form:"state"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(nil, systemSrv.SystemActState(ctx, v.ID, v.State))
}

func actInfo(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.SystemActInfo(ctx, v.ID))
}

func actList(ctx *bm.Context) {
	v := new(struct {
		Query string `form:"query"`
		Page  int64  `form:"page" validate:"required,min=1"`
		Size  int64  `form:"size" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.SystemActList(ctx, v.Query, v.Page, v.Size))
}

func seatList(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.SystemSeatList(ctx, v.AID))
}

func voteSum(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.VoteSumList(ctx, v.AID))
}

func voteOption(ctx *bm.Context) {
	v := new(struct {
		AID      int64 `form:"aid" validate:"required"`
		ItemID   int64 `form:"item_id"`
		OptionID int64 `form:"option_id"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.VoteOptionDetail(ctx, v.AID, v.ItemID, v.OptionID))
}

func voteDetailExport(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	body, header, err := systemSrv.ExportVoteDetail(ctx, v.AID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "投票活动id为" + strconv.FormatInt(v.AID, 10) + "投票明细.csv",
		Header:         header,
		Result:         body,
	}
	exportCsv(ctx, exportCsvParam)
}

func questionList(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.QuestionList(ctx, v.AID))
}

func questionState(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(systemSrv.QuestionState(ctx, v.ID))
}

func exportQuestionList(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required,min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	res, err := systemSrv.GetQuestionList(ctx, v.AID)
	if err != nil {
		log.Errorc(ctx, "systemSrv.GetQuestionList err")
		ctx.JSON(nil, err)
		return
	}

	exportCsv(ctx, &voguemdl.ExportCsvParam{
		FileNameFormat: "活动" + strconv.FormatInt(v.AID, 10) + "问答明细.csv",
		Header:         []string{"问题", "昵称", "姓名", "部门名称", "是否展示", "提问时间"},
		Result:         systemSrv.ExportQuestionList(ctx, res),
	})
}
