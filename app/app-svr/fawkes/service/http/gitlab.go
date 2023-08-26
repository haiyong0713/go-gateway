package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"strconv"
)

func branchTagList(c *bm.Context) {
	var (
		params     = c.Request.Form
		appKey, kw string
		branchType int
		res        = map[string]interface{}{}
		err        error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if branchType, err = strconv.Atoi(params.Get("type")); err != nil {
		res["message"] = "type 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if branchType < 0 || branchType > 1 {
		res["message"] = "未知 type"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	kw = params.Get("keyword")
	c.JSON(s.GitSvr.BranchTagList(c, appKey, branchType, kw))
}

func relatedMRCreate(c *bm.Context) {
	var (
		params                                                 = c.Request.Form
		projID, assignee, subRepos, sourceBranch, targetBranch string
		res                                                    = map[string]interface{}{}
	)
	if projID = params.Get("proj_id"); projID == "" {
		res["message"] = "proj_id 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if assignee = params.Get("assignee"); assignee == "" {
		res["message"] = "assignee 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if sourceBranch = params.Get("src_branch"); sourceBranch == "" {
		res["message"] = "src_branch 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if targetBranch = params.Get("tgt_branch"); targetBranch == "" {
		res["message"] = "tgt_branch 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	subRepos = params.Get("subrepos")
	c.JSON(s.GitSvr.CreateRelatedMR(c, projID, assignee, subRepos, sourceBranch, targetBranch))
}

func branchCommit(c *bm.Context) {
	var (
		params         = c.Request.Form
		projID, branch string
		res            = map[string]interface{}{}
	)
	if projID = params.Get("proj_id"); projID == "" {
		res["message"] = "proj_id 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if branch = params.Get("branch"); branch == "" {
		res["message"] = "branch 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.GitSvr.RemoteCommitFormBranch(c, projID, branch))
}
