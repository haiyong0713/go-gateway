package service

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	cheesemdl "go-gateway/app/app-svr/app-dynamic/interface/model/cheese"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	grpcShortURL "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"

	"github.com/pkg/errors"
)

func (s *Service) getEmoji(ctx context.Context, emojis []string) (map[string]*mdlv2.EmojiItem, error) {
	params := url.Values{}
	params.Set("texts", strings.Join(emojis, ","))
	params.Set("business", "dynamic")
	var ret struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data *mdlv2.Emoji `json:"data"`
	}
	if err := s.httpMgr.Get(ctx, s.emojiURL, "", params, &ret); err != nil {
		return nil, errors.WithStack(err)
	}
	if ret.Code != 0 || ret.Data == nil {
		return nil, errors.Wrapf(ecode.Int(ret.Code), "getEmoji url: %v, code: %v msg: %v or data nil", s.emojiURL, ret.Code, ret.Msg)
	}
	return ret.Data.Emote, nil
}

func (s *Service) shortUrls(ctx context.Context, urls []string) (map[string]string, error) {
	var max50 = 50
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	res := make(map[string]string, len(urls))
	for i := 0; i < len(urls); i += max50 {
		var partUrls []string
		if i+max50 > len(urls) {
			partUrls = urls[i:]
		} else {
			partUrls = urls[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			slices, err := s.shortUrlSlice(ctx, partUrls)
			if err != nil {
				return err
			}
			mu.Lock()
			for k, v := range slices {
				res[k] = v
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("shortUrls eg.wait(%+v)", err)
		return nil, err
	}
	return res, nil
}

func (s *Service) shortUrlSlice(c context.Context, urls []string) (map[string]string, error) {
	var args = &grpcShortURL.ShortUrlsReq{ShortUrls: urls}
	resTmp, err := s.shortURLGRPC.ShortUrls(c, args)
	if err != nil {
		return nil, errors.Wrapf(err, "s.shortURLGRPC.shortUrls args=%+v", args)
	}
	var res = make(map[string]string)
	for _, reTmp := range resTmp.GetDetails() {
		if resTmp == nil {
			continue
		}
		res[reTmp.ShortUrl] = reTmp.OriginUrl
	}
	return res, nil
}

// 课程
func (s *Service) AdditionalCheese(ctx context.Context, ssids []int64) (map[int64]*cheesemdl.Cheese, error) {
	params := url.Values{}
	params.Set("season_ids", xstr.JoinInts(ssids))
	req, err := s.httpMgr.NewRequest(http.MethodGet, s.attachCheeseCard, metadata.String(ctx, metadata.RemoteIP), params)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var resTmp struct {
		Code int                 `json:"code"`
		Msg  string              `json:"message"`
		Data []*cheesemdl.Cheese `json:"data"`
	}
	if err = s.httpMgr.Do(ctx, req, &resTmp); err != nil {
		return nil, err
	}
	if resTmp.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(resTmp.Code), s.attachCheeseCard+"?"+params.Encode())
	}
	var res = make(map[int64]*cheesemdl.Cheese)
	for _, cheese := range resTmp.Data {
		if cheese == nil || cheese.ID == 0 {
			continue
		}
		res[cheese.ID] = cheese
	}
	return res, nil
}
