package exporttask

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	httpClient  *bm.Client
	globalToken string
	lastUpdate  time.Time
	userID      = map[string]string{}
)

func (s *Service) GetUserIDMap() map[string]string {
	return userID
}

func (s *Service) UpdateMemberInfo() error {
	return GetMemberInfo()
}

func WeChatAccessToken(c context.Context) (token string, err error) {
	var (
		u      string
		params = url.Values{}
		res    struct {
			ErrCode     int    `json:"errcode"`
			ErrMsg      string `json:"errmsg"`
			AccessToken string `json:"access_token"`
			ExpiresIn   int32  `json:"expires_in"`
		}
	)
	if time.Now().Before(lastUpdate.Add(time.Hour)) {
		return globalToken, nil
	}
	u = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	params.Set("corpid", "wx0833ac9926284fa5")
	params.Set("corpsecret", "bsvsd21voBpVdEN1XWruDYCJH23FbQFy814BpPDFDS0")
	if err = httpClient.Get(c, u, "", params, &res); err != nil {
		return
	}
	if res.ErrCode != 0 {
		log.Errorc(c, "wechatAccessToken: errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
		return
	}
	token = res.AccessToken
	lastUpdate = time.Now()
	globalToken = token
	return
}

func SendWeChatTextMessage(c context.Context, username []string, message string) (err error) {
	idArr := make([]string, 0, len(username))
	for _, name := range username {
		if id, ok := userID[name]; ok {
			idArr = append(idArr, id)
		}
	}
	if len(idArr) == 0 {
		log.Errorc(c, "unknow userid list: %v", username)
		return errors.New("unknow userid list")
	}
	var (
		req   *http.Request
		token string
		res   struct {
			ErrCode      int64  `json:"errcode"`
			ErrMsg       string `json:"errmsg"`
			InvalidUser  string `json:"invaliduser"`
			InvalidParty string `json:"invalidparty"`
			InvalidTag   string `json:"invalidtag"`
		}
	)
	params := url.Values{}

	if token, err = WeChatAccessToken(c); err != nil {
		log.Errorc(c, "sendMessageToUser get token error(%v)", err)
		return
	}

	params.Set("access_token", token)
	_url := "https://qyapi.weixin.qq.com/cgi-bin/message/send?" + params.Encode()

	var buf []byte
	buf, _ = json.Marshal(struct {
		ToUser  string            `json:"touser"`
		MsgType string            `json:"msgtype"`
		AgentID int               `json:"agentid"`
		Text    map[string]string `json:"text"`
	}{
		ToUser:  strings.Join(idArr, "|"),
		MsgType: "text",
		AgentID: 1000051,
		Text:    map[string]string{"content": message},
	})

	body := bytes.NewBuffer(buf)

	if req, err = http.NewRequest("POST", _url, body); err != nil {
		log.Errorc(c, "sendMessageToUser url(%s) error(%v)", _url, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = httpClient.Do(c, req, &res); err != nil {
		log.Errorc(c, "sendMessageToUser Do failed url(%s) response(%+v) error(%v)", _url, res, err)
		err = ecode.ServerErr
		return
	}
	log.Infoc(c, "sendMessageToUser res %v", res)
	return
}

func GetMemberInfoProc() {
	_ = GetMemberInfo()
	for range time.Tick(time.Hour * 24) {
		_ = GetMemberInfo()
	}
}

func GetMinDeptToFetch(ctx context.Context, token string) (parentIds map[int]string, err error) {
	params := url.Values{}
	parentIds = make(map[int]string)
	params.Set("access_token", token)
	u := "https://qyapi.weixin.qq.com/cgi-bin/department/list?" + params.Encode()
	log.Infoc(ctx, "GetMinDeptToFetch %s", u)
	var req *http.Request
	if req, err = http.NewRequest("GET", u, nil); err != nil {
		log.Errorc(ctx, "GetMinDeptToFetch url(%s) error(%v)", u, err)
		return
	}
	type Dept struct {
		Id       int    `json:"id"`
		ParentId int    `json:"parentid"`
		Name     string `json:"name"`
	}
	res := struct {
		Errcode  int     `json:"errcode"`
		Errmsg   string  `json:"errmsg"`
		DeptList []*Dept `json:"department"`
	}{}
	if req != nil {
		if err = httpClient.Do(ctx, req, &res); err != nil {
			log.Errorc(ctx, "GetMinDeptToFetch do failed. url(%s) res(%v) error(%v)", u, res, err)
			return
		}
		if res.Errcode > 0 {
			log.Errorc(ctx, "GetMinDeptToFetch do failed, url(%s) res(%v) error(%v)", u, res, res.Errmsg)
		}
	}
	deptMap := make(map[int]*Dept, len(res.DeptList))
	for _, dept := range res.DeptList {
		t := dept
		deptMap[t.Id] = t
		if t.ParentId == 1 {
			parentIds[t.Id] = t.Name
		} else if _, ok := deptMap[t.ParentId]; ok {
			parentIds[t.ParentId] = t.Name
		} else {
			parentIds[t.Id] = t.Name
		}
	}
	//部门向上合并
	for {
		tmpParentIds := make(map[int]string)
		for deptId := range parentIds {
			t, ok := deptMap[deptId]
			if !ok { //找不到, 不再次进行合并
				tmpParentIds[deptId] = "UNKNOWN"
				continue
			}
			if t.ParentId == 1 { //parentId=1的, 不再次进行合并
				tmpParentIds[deptId] = t.Name
				continue
			}
			if _, ok := deptMap[t.ParentId]; ok {
				tmpParentIds[t.ParentId] = t.Name
			} else {
				tmpParentIds[t.Id] = t.Name
			}
		}
		if len(parentIds) == len(tmpParentIds) { //本次合并没有变化, 达到退出条件
			break
		}
		parentIds = tmpParentIds
	}
	return
}

func GetMemberInfo() (err error) {
	c, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*20))
	defer cancel()
	var token string
	if token, err = WeChatAccessToken(c); err != nil {
		log.Errorc(c, "GetMemberInfo sendMessageToUser get token error(%v)", err)
		return
	}
	tmpMap := make(map[string]string)
	depts, err := GetMinDeptToFetch(c, token)
	if err != nil {
		log.Errorc(c, "GetMemberInfo GetMinDeptToFetch error(%v)", err)
	}
	for deptId := range depts {
		if deptId == 1 {
			continue
		}
		params := url.Values{}
		params.Set("access_token", token)
		params.Set("department_id", strconv.Itoa(deptId))
		params.Set("fetch_child", "1")
		u := "https://qyapi.weixin.qq.com/cgi-bin/user/list?" + params.Encode()
		log.Infoc(c, "GetMemberInfo %s", u)
		var req *http.Request
		if req, err = http.NewRequest("GET", u, nil); err != nil {
			log.Errorc(c, "GetMemberInfo url(%s) error(%v)", u, err)
			return
		}
		res := struct {
			Errcode  int    `json:"errcode"`
			Errmsg   string `json:"errmsg"`
			Userlist []struct {
				Userid string `json:"userid"`
				Name   string `json:"english_name"`
			} `json:"userlist"`
		}{}
		if req != nil {
			if err = httpClient.Do(c, req, &res); err != nil {
				log.Errorc(c, "GetMemberInfo do failed. url(%s) res(%v) error(%v)", u, res, err)
				return
			}
			if res.Errcode > 0 {
				log.Errorc(c, "GetMemberInfo do failed, url(%s) res(%v) error(%v)", u, res, res.Errmsg)
			}
		}
		if len(res.Userlist) > 0 {
			for _, one := range res.Userlist {
				tmpMap[one.Name] = one.Userid
			}
		}
	}

	if len(tmpMap) > 0 {
		userID = tmpMap
	}
	return
}
