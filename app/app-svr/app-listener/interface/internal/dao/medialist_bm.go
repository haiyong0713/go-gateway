package dao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

type MediaListReqContext struct {
	Ctx      context.Context
	PageSize int
	FetchAll bool // 是否获取所有资源信息
	MaxWant  int  // 需要的资源上限（对FetchAll做限制）
	FnDo     func(ctx context.Context, req *http.Request, res interface{}, v ...string) error
	FnUri    func(path string) string
	Anchor   *v1.PlayItem
	total    int
}

func (c *MediaListReqContext) defaultPager() {
	const defaultPageSize = 99
	if c.PageSize <= 0 {
		c.PageSize = defaultPageSize
	}
	const defaultMaxWant = 300
	if c.MaxWant <= 0 {
		c.MaxWant = defaultMaxWant
	}
}

func (c *MediaListReqContext) genericParams(p url.Values) {
	for k, v := range appParams(c.Ctx) {
		p.Set(k, v[0])
	}
}

func (c *MediaListReqContext) setRequestParams(p url.Values, kvs ...string) {
	if len(kvs)%2 != 0 {
		panic("programmer error: unexpected odd kvs count while setting request params")
	}
	for i := 0; i < len(kvs); i += 2 {
		p.Set(kvs[i], kvs[i+1])
	}
}

type MediaListDoListOpt struct {
	Typ   int
	BizId int64
}

//nolint:gocognit
func (c *MediaListReqContext) DoList(opt MediaListDoListOpt) (ret []model.MediaListItem, err error) {
	if c.FnDo == nil {
		panic("programmer error: unexpected nil FnDo for MediaListReqContext")
	}
	req, _ := http.NewRequest(http.MethodGet, c.FnUri(_mediaListResources), nil)
	param := url.Values{}
	c.genericParams(param)
	c.defaultPager()

	// request params
	kvs := []string{
		"desc", "true", // 必须保持在第一组
		"type", strconv.Itoa(opt.Typ),
		"biz_id", strconv.FormatInt(opt.BizId, 10),
		"direction", "false",
		"with_current", "true",
		"sort_field", "1", // 1：创建时间；2: 播放量；3：收藏量
		"ps", strconv.Itoa(c.PageSize),
	}
	// flag
	useOidOffset := false
	useDescReverse := false    // 使用 逆序进行反向而不是向上翻页
	var upLast, downLast int64 // 向上和向下翻页的锚点元素
	if c.Anchor != nil {
		if c.Anchor.Oid > 0 && c.Anchor.ItemType == model.PlayItemUGC {
			kvs = append(kvs, "oid", strconv.FormatInt(c.Anchor.Oid, 10))
			upLast = c.Anchor.Oid
			useOidOffset = true
		}
	}

	switch opt.Typ {
	case model.MediaListTypeLater:
		// 清除稍候再看插入的锚点
		if useOidOffset {
			kvs = kvs[0 : len(kvs)-2]
		}
		useOidOffset = false // 对于稍候再看类型 一次性获取所有内容
		kvs[1] = "false"     // 稍后再看默认保持正序
	case model.MediaListTypeSeries:
		useOidOffset = true // 对于系列视频 不支持页码翻页
		useDescReverse = true
	}

	c.setRequestParams(param, kvs...)
	data, err := c.doListSingle(req, param)
	if err != nil {
		return
	}
	c.total = data.Total
	if len(data.MediaList) > 0 {
		downLast = data.MediaList[len(data.MediaList)-1].Avid
	}
	ret = data.MediaList

	if !c.FetchAll {
		return
	}
	if c.total <= c.PageSize && c.Anchor == nil {
		// fetchAll的情况下 总数少于 pageSize 且 没有指定锚点 直接返回即可
		return
	}
	// 获取其他页面的时候需要重写一部分请求参数
	kvs = append(kvs[0:0], "with_current", "false")
	eg := errgroup.WithCancel(c.Ctx)
	if useOidOffset {
		c.setRequestParams(param, kvs...)
		// 使用oid翻页
		// 需要分为两个方向，向上和向下
		mu := sync.Mutex{}
		eg.Go(func(ctx context.Context) error {
			// 向下获取
			reqDown := req.Clone(ctx)
			paramDown := cloneUrlParam(param)
			for {
				c.setRequestParams(paramDown, "oid", strconv.FormatInt(downLast, 10))
				data, err := c.doListSingle(reqDown, paramDown)
				if err != nil {
					return err
				}
				if len(data.MediaList) > 0 {
					downLast = data.MediaList[len(data.MediaList)-1].Avid
				}
				mu.Lock()
				ret = append(ret, data.MediaList...)
				if len(ret) >= c.MaxWant {
					mu.Unlock()
					break
				}
				mu.Unlock()
				if !data.HasMore {
					break
				}
			}
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			// 没锚点，不用考虑向上了
			if upLast <= 0 {
				return nil
			}
			// 向上获取
			reqUp := req.Clone(ctx)
			paramUp := cloneUrlParam(param)
			if useDescReverse {
				// 默认为 desc true  设置为反向
				c.setRequestParams(paramUp, "desc", "false")
			} else {
				c.setRequestParams(paramUp, "direction", "true") // 设置方向为向上
			}
			for {
				c.setRequestParams(paramUp, "oid", strconv.FormatInt(upLast, 10))
				data, err := c.doListSingle(reqUp, paramUp)
				if err != nil {
					return err
				}
				if len(data.MediaList) > 0 {
					if !useOidOffset {
						upLast = data.MediaList[0].Avid
					} else {
						upLast = data.MediaList[len(data.MediaList)-1].Avid
					}
				}
				mu.Lock()
				if !useDescReverse {
					// 正常情况下直接写入到列表头部
					ret = append(data.MediaList, ret...)
				} else {
					// 逆序下翻的情况下，需要逆序列表然后写入
					oriRet := ret
					ret = make([]model.MediaListItem, 0, len(ret)+len(data.MediaList))
					for i := len(data.MediaList) - 1; i >= 0; i-- {
						ret = append(ret, data.MediaList[i])
					}
					ret = append(ret, oriRet...)
				}
				if len(ret) >= c.MaxWant {
					mu.Unlock()
					break
				}
				mu.Unlock()
				if !data.HasMore {
					break
				}
			}
			return nil
		})
	} else {
		// 使用pn ps翻页
		kvs = append(kvs, "use_pn", "true")
		totalPages := c.total / c.PageSize
		if c.total%c.PageSize != 0 {
			totalPages += 1
		}
		if c.total > c.MaxWant {
			totalPages = c.MaxWant / c.PageSize
		}
		c.setRequestParams(param, kvs...)
		// 分页请求
		initCh := make(chan struct{}, 1)
		var tmp chan struct{}
		for i := 2; i <= totalPages; i++ {
			reqClone := req.Clone(c.Ctx)
			paramClone := cloneUrlParam(param)
			c.setRequestParams(paramClone, "pn", strconv.Itoa(i))
			var prev, next chan struct{}
			//nolint:gomnd
			if i <= 2 {
				prev = initCh
			} else {
				prev = tmp
			}
			next = make(chan struct{}, 1)
			tmp = next
			eg.Go(func(ctx context.Context) error {
				data, err := c.doListSingle(reqClone, paramClone)
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-prev:
				}
				ret = append(ret, data.MediaList...)
				next <- struct{}{}
				return nil
			})
		}
		initCh <- struct{}{}
	}
	err = eg.Wait()

	return
}

func (c *MediaListReqContext) doListSingle(req *http.Request, param url.Values) (ret *model.MediaListData, err error) {
	req.URL.RawQuery = param.Encode()
	resp := model.BmGenericResp{}
	if err = c.FnDo(c.Ctx, req, &resp); err != nil {
		return nil, wrapHttpError(err, req.URL.String(), req)
	}
	if err = resp.IsNormal(); err != nil {
		return nil, errors.WithMessagef(err, "doListSingle resp is not normal req(%s) resp(%+v)", req.URL.String(), resp)
	}
	ret = new(model.MediaListData)

	err = json.Unmarshal(resp.Data.Bytes(), ret)
	if err != nil {
		return
	}
	return
}

const (
	_mediaListResources = "/x/v2/medialist/resource/list"
)

type MediaListDoPageOpt struct {
	Typ    int
	BizId  int64
	Offset string
}

type DoPageResp struct {
	Items   []model.MediaListItem
	Offset  string
	HasMore bool
}

func (c *MediaListReqContext) DoPage(opt MediaListDoPageOpt) (ret DoPageResp, err error) {
	if c.FnDo == nil {
		panic("programmer error: unexpected nil FnDo for MediaListReqContext")
	}
	req, _ := http.NewRequest(http.MethodGet, c.FnUri(_mediaListResources), nil)
	param := url.Values{}
	c.genericParams(param)
	c.defaultPager()

	// request params
	kvs := []string{
		"desc", "true", // 必须保持在第一组
		"type", strconv.Itoa(opt.Typ),
		"biz_id", strconv.FormatInt(opt.BizId, 10),
		"direction", "false",
		"sort_field", "1", // 1：创建时间；2: 播放量；3：收藏量
		"ps", strconv.Itoa(c.PageSize),
		"with_current", "true",
	}
	if len(opt.Offset) > 0 {
		// 翻页时不带翻页元素
		kvs = append(kvs[:len(kvs)-2], "with_current", "false")
		kvs = append(kvs, "oid", opt.Offset)
	} else if c.Anchor != nil {
		if c.Anchor.Oid > 0 && c.Anchor.ItemType == model.PlayItemUGC {
			kvs = append(kvs, "oid", strconv.FormatInt(c.Anchor.Oid, 10))
		}
	}

	switch opt.Typ {
	case model.MediaListTypeLater:
		kvs[1] = "false" // 稍后再看默认保持正序
	}

	c.setRequestParams(param, kvs...)
	data, err := c.doListSingle(req, param)
	if err != nil {
		return
	}
	c.total = data.Total
	if len(data.MediaList) > 0 {
		ret.Offset = strconv.FormatInt(data.MediaList[len(data.MediaList)-1].Avid, 10)
	}
	ret.Items = data.MediaList
	if data.HasMore || len(data.MediaList) >= c.PageSize {
		ret.HasMore = true
	}
	return
}
