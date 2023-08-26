package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// 手动重新执行一次今天的task
func manuallyRunJob(c *bm.Context) {
	var (
		err error
		req = struct {
			LogDate string `json:"log_date" validate:"required"`
		}{}
	)

	if err = c.BindWith(&req, binding.JSON); err != nil {
		return
	}

	runnerKey := "job_rank_dataplat_" + req.LogDate

	//nolint:biligowordcheck
	go rankSvc.AtomGenDataplatAvRank(runnerKey)

	c.JSON(nil, err)
}

// 给网关用，返回所有的可见榜单详情
func openRankList(c *bm.Context) {
	var (
		err   error
		pager *rankModel.OpenListPager
		req   = &rankModel.RankCommonQuery{}
		res   = map[string]interface{}{}
	)

	if err = c.Bind(req); err != nil {
		return
	}

	if pager, err = rankSvc.OpenRankList(c, req.Size, req.Page); err != nil {
		res["message"] = "获取榜单失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	c.JSON(pager, err)
}

// 获取榜单内全部视频信息
func rankDetail(c *bm.Context) {
	var (
		err   error
		pager *rankModel.RankDetailPager
	)
	res := map[string]interface{}{}
	req := &rankModel.RankCommonQuery{}
	if err = c.Bind(req); err != nil {
		return
	}

	if pager, err = rankSvc.GetRankAVList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	c.JSON(pager, err)
}

// 获取全部榜单
func rankList(c *bm.Context) {
	var (
		err   error
		pager *rankModel.RankListPager
	)
	res := map[string]interface{}{}
	req := &rankModel.RankCommonQuery{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, username := util.UserInfo(c)
	if pager, err = rankSvc.GetRankList(c, req, username, uid); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

// 获取某榜单配置信息
func rankConfig(c *bm.Context) {
	var (
		err    error
		config *rankModel.RankConfigRes
	)
	res := map[string]interface{}{}
	req := &rankModel.RankCommonQuery{}
	if err = c.Bind(req); err != nil {
		return
	}
	if config, err = rankSvc.GetRankConfig(c, req); err != nil {
		res["message"] = "榜单配置获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(config, nil)
}

// 添加新的排行榜
func rankAdd(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
		req = &rankModel.RankConfigReq{}
	)

	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}

	uid, username := util.UserInfo(c)

	if err = rankSvc.AddNewRank(req, username, uid); err != nil {

		res["message"] = "添加失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	c.JSON(res, nil)
}

// 编辑榜单
func rankEdit(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &rankModel.EditRankConfigReq{}
	if err = c.BindWith(req, binding.JSON); err != nil {
		return
	}
	uid, username := util.UserInfo(c)
	if err = rankSvc.EditRankConfig(req, username, uid); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	c.JSON(res, nil)
}

// 添加稿件干预,包括指定排名,前端隐藏,赋予额外分数
func rankArchiveEdit(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	var req rankModel.RankArchiveIntervention
	if err = c.BindWith(&req, binding.JSON); err != nil {
		return
	}
	if req.RankId < 1 || req.Avid < 1 {
		err = ecode.Error(ecode.RequestErr, "必须提供正确的榜单id和稿件avid!")
		c.JSONMap(res, err)
		return
	}
	if req.Rank < 0 {
		req.Rank = 0
	}
	if err = rankSvc.RankArchiveEdit(c, &req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["message"] = "编辑成功"
	c.JSON(res, nil)
}

// 手动为榜单添加稿件
func rankArchiveAdd(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	var req rankModel.RankCommonQuery
	if err = c.BindWith(&req, binding.JSON); err != nil {
		return
	}
	if err = rankSvc.RankArchiveAdd(int64(req.Id), req.Avid); err != nil {

		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["message"] = "编辑成功"
	c.JSON(res, nil)
}

// 发布榜单
func rankPublish(c *bm.Context) {
	var (
		err error
		res map[string]string
	)

	var req rankModel.RankCommonQuery
	if err = c.BindWith(&req, binding.JSON); err != nil {
		return
	}

	_, username := util.UserInfo(c)
	if err = rankSvc.RankPublish(c, req.Id, username); err != nil {
		c.JSON(res, err)
		return
	}
	c.JSON(res, nil)
}

// 结榜. 将结榜配置加入最新发布的榜单中,更改榜单的状态为'已结榜'
func rankTerminate(c *bm.Context) {
	var (
		err error
		res map[string]string
	)

	var req rankModel.TernimateContent
	if err = c.BindWith(&req, binding.JSON); err != nil {
		return
	}
	_, username := util.UserInfo(c)
	if err = rankSvc.RankTerminate(c, &req, username); err != nil {
		c.JSON(res, err)
		return
	}
	c.JSON(res, nil)
}

// 榜单操作,尚未生效!
func rankOption(c *bm.Context) {
	var (
		err error
		res map[string]string
	)

	var req rankModel.RankCommonQuery
	if err = c.BindWith(&req, binding.JSON); err != nil {
		return
	}
	if err = rankSvc.RankOption(c, &req); err != nil {
		//	res["message"] = "列表获取失败 " + err.Error()
		c.JSON(res, err)
		return
	}
	c.JSON(res, nil)
}
