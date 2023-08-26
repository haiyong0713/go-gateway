package http

import (
	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/web-svr/activity/interface/model/system"
	"go-gateway/app/web-svr/activity/interface/service"
)

func WXAuth(ctx *bm.Context) {
	v := new(model.WXAuthArgs)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SystemSvc.WXAuth(ctx, v))
}

func getConfig(ctx *bm.Context) {
	v := new(model.GetConfigArgs)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SystemSvc.GetConfig(ctx, v))
}

func internalGetUserInfoByUID(ctx *bm.Context) {
	v := new(struct {
		UID string `form:"uid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SystemSvc.GetUserInfoByUID(ctx, v.UID))
}

func internalGetUserInfoByCookie(ctx *bm.Context) {
	ctx.JSON(service.SystemSvc.GetUserInfoByCookie(ctx, GetSessionTokenFromCookie(ctx)))
}

func sign(ctx *bm.Context) {
	v := new(struct {
		AID      int64  `form:"aid" validate:"required"`
		Location string `form:"location"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.SystemSvc.Sign(ctx, v.AID, GetSessionTokenFromCookie(ctx), v.Location))
}

func activityInfo(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SystemSvc.ActivityInfo(ctx, v.AID, GetSessionTokenFromCookie(ctx)))
}

func GetSessionTokenFromCookie(ctx *bm.Context) string {
	cookie, err := ctx.Request.Cookie("bili_corp_token")
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func party2021(ctx *bm.Context) {
	ctx.JSON(service.SystemSvc.Party2021(ctx, GetSessionTokenFromCookie(ctx)))
}

func internalGetUsersInfoByUIDs(ctx *bm.Context) {
	v := new(struct {
		UIDs []string `form:"uids" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SystemSvc.GetUsersInfoByUIDs(ctx, v.UIDs))
}

func internalAddV(ctx *bm.Context) {
	v := new(struct {
		UserID         string `form:"user_id" validate:"required"`
		DepartmentName string `form:"department_name" validate:"required"`
		Name           string `form:"name" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.SystemSvc.AddV(ctx, v.UserID, v.DepartmentName, v.Name))
}

func systemVote(ctx *bm.Context) {
	v := new(struct {
		AID     int64  `form:"aid" validate:"required"`
		Content string `form:"content" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.SystemSvc.Vote(ctx, v.AID, GetSessionTokenFromCookie(ctx), v.Content))
}

func internalPrizeNotify(ctx *bm.Context) {
	v := new(struct {
		UIDs    string `form:"uids" validate:"required"`
		Message string `form:"message" validate:"required"`
		From    string `form:"from" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.SystemSvc.Notify(ctx, v.UIDs, v.Message, v.From))
}

func systemQuestion(ctx *bm.Context) {
	v := new(struct {
		AID     int64  `form:"aid" validate:"required"`
		Content string `form:"content" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(nil, service.SystemSvc.Question(ctx, v.AID, GetSessionTokenFromCookie(ctx), v.Content))
}

func systemQuestionList(ctx *bm.Context) {
	v := new(struct {
		AID int64 `form:"aid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.SystemSvc.QuestionList(ctx, v.AID, GetSessionTokenFromCookie(ctx)))
}
