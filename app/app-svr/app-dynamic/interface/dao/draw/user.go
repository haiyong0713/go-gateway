package dao

import (
	"context"
	"net/url"
	"strconv"

	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	"github.com/pkg/errors"
)

const (
	_urlGetFollowing = "/feed/v0/Feed/get_followings"
	_urlGetUserInfos = "/account/v1/User/infos"
	_urlMainSearch   = "/main/search"
	_urlGetRecentAt  = "/dynamic_mix/v0/dynamic_mix/ircmd_at"
)

// 获取用户最近关注的k人
func (d *Dao) GetUserLatestFollowTopK(ctx context.Context, uid uint64, k int, _ string) (users []*model.UserSearchItem, err error) {
	params := url.Values{}
	params.Set("uid", strconv.Itoa(int(uid)))
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			All []interface{} `json:"all"`
		} `json:"data"`
	}
	queryUrl := d.conf.Hosts.VcCo + _urlGetFollowing
	if err = d.client.Get(ctx, queryUrl, "", params, &ret); err != nil {
		log.Error("%s query failed, mid(%d), error(%v)", _urlGetFollowing, uid, err)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s return error(%d) msg(%s) mid(%d)", _urlGetFollowing, ret.Code, ret.Msg, uid)
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl+"?"+params.Encode())
		return nil, err
	}
	if len(ret.Data.All) == 0 {
		log.Error("%s return empty, mid(%d)", _urlMainSearch, uid)
		return
	}
	intAll, err := convertSliceInterfaceToInt(ret.Data.All)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	users, err = d.GetUserSearchItems(ctx, intAll[:model.MinInt(len(ret.Data.All), k)])
	if err != nil {
		log.Error("fetch user infos with mids failed, error(%+v)", err)
	}
	return
}

func convertSliceInterfaceToInt(arr []interface{}) (ret []int64, err error) {
	for _, e := range arr {
		switch v := e.(type) {
		case float64:
			ret = append(ret, int64(v))
		case string:
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, int64(i))
		default:
			return nil, errors.New("invalid json content, should not contain non-number")
		}
	}
	return
}

// 获取用户最近@人
func (d *Dao) GetUserLatestAtUsers(ctx context.Context, uid uint64, _ string) (users []*model.UserSearchItem, err error) {
	params := url.Values{}
	params.Set("uid", strconv.Itoa(int(uid)))
	params.Set("teenagers_mode", "0")
	queryUrl := d.conf.Hosts.VcCo + _urlGetRecentAt
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			RecentAtUsers []struct {
				Info struct {
					Face  string `json:"face"`
					UID   uint64 `json:"uid"`
					Uname string `json:"uname"`
				} `json:"info"`
			} `json:"recent_at_users"`
		} `json:"data"`
	}
	if err = d.client.Get(ctx, queryUrl, "", params, &ret); err != nil {
		log.Error("%s query failed, mid(%d), error(%v)", _urlGetRecentAt, uid, err)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s return err code(%d) msg(%s) mid(%d)", _urlGetRecentAt, ret.Code, ret.Msg, uid)
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl+"?"+params.Encode())
		return nil, err
	}
	if len(ret.Data.RecentAtUsers) == 0 {
		log.Error("%s return empty mid(%d)", _urlGetRecentAt, uid)
		return
	}
	for _, user := range ret.Data.RecentAtUsers {
		users = append(users, &model.UserSearchItem{
			Face: user.Info.Face,
			Name: user.Info.Uname,
			Mid:  user.Info.UID,
		})
	}
	return
}

// 批量获取用户详情
func (d *Dao) GetUserSearchItems(ctx context.Context, mids []int64) (users []*model.UserSearchItem, err error) {
	type User struct {
		Face     string `json:"face"`
		Uid      uint64 `json:"uid"`
		UserName string `json:"uname"`
	}
	var ret struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data []*User `json:"data"`
	}
	midsStr := xstr.JoinInts(mids)
	params := url.Values{}
	params.Set("uids", midsStr)
	queryUrl := d.conf.Hosts.VcCo + _urlGetUserInfos
	if err = d.client.Get(ctx, queryUrl, "", params, &ret); err != nil {
		log.Error("%s query failed, mids(%s) error(%v)", _urlGetUserInfos, midsStr, err)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s return error(%d),msg(%s) mids(%s)", _urlGetUserInfos, ret.Code, ret.Msg, midsStr)
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl+"?"+params.Encode())
		return nil, err
	}
	if len(ret.Data) == 0 {
		log.Error("%s return empty, mids(%s)", _urlGetUserInfos, midsStr)
		return make([]*model.UserSearchItem, 0), nil
	}
	// 按入参mids的顺序append
	userInfoMap := make(map[uint64]*User, len(ret.Data))
	for _, userInfo := range ret.Data {
		if userInfo != nil {
			userInfoMap[userInfo.Uid] = userInfo
		}
	}
	for _, uid := range mids {
		user := userInfoMap[uint64(uid)]
		users = append(users, &model.UserSearchItem{
			Face: user.Face,
			Name: user.UserName,
			Mid:  user.Uid,
		})
	}
	return
}

// 搜索用户
func (d *Dao) SearchUser(ctx context.Context, uid uint64, word string, page, pageSize int) (users []*model.UserSearchItem, hasMore bool, err error) {
	page = fixPage(page)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("search_type", "bili_user")
	params.Set("keyword", word)
	params.Set("page", strconv.Itoa(page))
	params.Set("page_size", strconv.Itoa(pageSize))
	var ret struct {
		Code     int                     `json:"code"`
		Result   []*model.UserSearchItem `json:"result"`
		NumPages int                     `json:"numPages"`
		Page     int                     `json:"page"`
	}
	queryUrl := d.conf.Hosts.SearchCo + _urlMainSearch
	if err = d.client.Get(ctx, queryUrl, "", params, &ret); err != nil {
		log.Error("%s query failed, mid(%d), word(%s), error(%v)", _urlMainSearch, uid, word, err)
		return nil, hasMore, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s return error(%d), mid(%d), word(%s)", _urlMainSearch, ret.Code, uid, word)
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl+"?"+params.Encode())
		return nil, hasMore, err
	}
	users = ret.Result
	if len(users) == 0 {
		log.Error("%s return empty, mid(%d), word(%s)", _urlMainSearch, uid, word)
		return
	}
	if ret.Page < ret.NumPages {
		hasMore = true
	}
	return
}
