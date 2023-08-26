package article

import (
	"context"
	"fmt"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/article"
	"go-gateway/app/app-svr/app-feed/admin/util"

	articleModel "git.bilibili.co/bapis/bapis-go/article/model"
	articleRpc "git.bilibili.co/bapis/bapis-go/article/service"
)

const (
	articleURL  = "/x/internal/article/meta"
	articlesURL = "/x/internal/article/metas"
)

// Article .
func (d *Dao) Article(c context.Context, articleids []int64) (res *article.Article, err error) {
	params := url.Values{}
	params.Set("id", xstr.JoinInts(articleids))
	res = new(article.Article)
	if err = d.articleHTTPClient.Get(c, d.c.Host.API+articleURL, "", params, &res); err != nil {
		log.Error("Article Req(%v) error(%v) res(%+v)", articleids, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.c.Host.API+articleURL+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("Article Req(%v) error(%v) res(%+v)", articleids, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Article, d.c.Host.API+articleURL+"?"+params.Encode())
	}
	return
}

// ArticlesInfo .
func (d *Dao) ArticlesInfo(c context.Context, articleids []int64) (res map[int64]*article.ArticleInfo, err error) {
	params := url.Values{}
	params.Set("ids", xstr.JoinInts(articleids))
	resp := new(article.Articles)
	if err = d.articleHTTPClient.Get(c, d.c.Host.API+articlesURL, "", params, &resp); err != nil {
		return
	}
	if resp.Code != ecode.OK.Code() {
		log.Error("ArticlesInfo Req(%v) error(%v) res(%+v)", articleids, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.c.Host.API+articlesURL+"?"+params.Encode(), err.Error())
	}
	res = make(map[int64]*article.ArticleInfo)
	for _, r := range resp.Data {
		if r == nil {
			log.Error("ArticlesInfo url(%s) res is nil", d.c.Host.API+articleURL+"?"+params.Encode())
			continue
		}
		res[r.ID] = r
	}
	return
}

func (d *Dao) ArticlesInfoRpc(c context.Context, idList []int64) (resp *articleRpc.ArticleMetasSimpleReply, err error) {
	req := &articleRpc.ArticleMetasSimpleReq{
		Ids: idList,
	}

	resp, err = d.articleRpcClient.ArticleMetasSimple(c, req)
	if err != nil {
		log.Errorc(c, "dao.ListArticle rpc ArticleMetasSimple idList(%+v) error(%+v)", idList, err)
		return
	}
	return
}

func (d *Dao) ArticleRpc(c context.Context, id int64) (ret *articleModel.Meta, err error) {
	resp, err := d.ArticlesInfoRpc(c, []int64{id})
	if err != nil {
		log.Errorc(c, "dao.DetailArticle ListArticle id(%+v) error(%+v)", id, err)
		return
	}

	ret, ok := resp.Res[id]
	if !ok {
		log.Errorc(c, "dao.DetailArticle get info for id(%v) error(%+v)", id, err)
		return
	}

	return ret, nil
}
