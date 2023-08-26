package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

const (
	ruleURL              = "http://hawkeye.bilibili.co/api/v3/product/caster%E4%B8%9A%E5%8A%A1/team/"
	alertRule            = "/alert_rule/"
	findRuleIdURL        = "http://hawkeye.bilibili.co/api/v3/rules/alert_rule?pageSize=10&page=1&product=caster%E4%B8%9A%E5%8A%A1&"
	tag                  = "【网关业务机器人】"
	authURL              = "http://easyst.bilibili.co/v1/auth"
	treeURL              = "http://easyst.bilibili.co/v1/node/role/app"
	ActiveReceiverURL    = "http://hawkeye.bilibili.co/buzzer/receiver_groups?type=tree&role=rd&status=active&page=1&size=500"
	putReceiverURL       = "http://hawkeye.bilibili.co/buzzer/receiver_group"
	ExcludedRecevicerURL = "http://hawkeye.bilibili.co/buzzer/receiver_groups?type=tree&role=rd&status=excluded&page=1&size=100&team_match=%s"
)

func (s *Service) GetRuleId(ctx context.Context, rule *model.CodeRule) (int64, error) {
	ruleId, err := s.dao.SelectRuleId(ctx, rule)
	if err != nil {
		return -1, err
	}
	return ruleId, nil
}

func (s *Service) InsertCodeRule(ctx context.Context, rule *model.CodeRule) error {
	if err := s.dao.InsertCodeRule(ctx, rule); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteCodeRule(ctx context.Context, ruleId int64) error {
	if err := s.dao.DeleteCodeRule(ctx, ruleId); err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateHawkeyeRule(ctx context.Context, ruleId int64, updateRuleReq *model.UpdateRuleReq, cookie string) error {
	ruleDetailURL := ruleURL + updateRuleReq.Team + alertRule + strconv.FormatInt(ruleId, 10)
	headers := s.RuleCookieHeader(cookie)
	if _, err := httpPut(ruleDetailURL, updateRuleReq, headers); err != nil {
		return err
	}
	return nil
}

func (s *Service) InsertHawkeyeRule(ctx context.Context, insertRuleReq *model.InsertRuleReq, cookie string) error {
	var insertRuleReply *model.RuleReply
	headers := s.RuleCookieHeader(cookie)
	reqURL := ruleURL + insertRuleReq.Team + "/alert_rule"
	data, err := httpPost(reqURL, insertRuleReq, headers)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &insertRuleReply); err != nil {
		return err
	}
	if insertRuleReply.Code != 0 {
		return errors.New(insertRuleReply.Message)
	}
	return nil
}

func (s *Service) DeleteHawkeyeRule(ctx context.Context, team string, ruleId int64, cookie string) error {
	var deleteRuleReply *model.RuleReply
	ruleDetailURL := ruleURL + team + alertRule + strconv.FormatInt(ruleId, 10)
	headers := s.RuleCookieHeader(cookie)
	data, err := httpDelete(ruleDetailURL, headers)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &deleteRuleReply); err != nil {
		return err
	}
	if deleteRuleReply.Code != 0 {
		return errors.New(deleteRuleReply.Message)
	}
	return nil
}

func (s *Service) HawkeyeRuleDetail(ctx context.Context, team string, ruleId int64, cookie string) (*model.FindRuleDetail, error) {
	var ruleDetail *model.FindRuleDetailReply
	ruleDetailURL := ruleURL + team + alertRule + strconv.FormatInt(ruleId, 10)
	headers := s.RuleCookieHeader(cookie)
	data, err := httpGet(ruleDetailURL, headers)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &ruleDetail); err != nil {
		return nil, err
	}
	if ruleDetail.Code != 0 {
		return nil, errors.New(ruleDetail.Message)
	}
	return ruleDetail.Data, nil
}

func (s *Service) HawkeyeRuleIdByName(ctx context.Context, team string, keyword string, cookie string) (int64, error) {
	items, err := s.HawkeyeRuleByName(ctx, team, keyword, cookie)
	if err != nil {
		return -1, err
	}
	if len(items) == 0 {
		return -1, errors.New("cannot find rule id")
	}
	return items[0].ID, nil
}

func (s *Service) HawkeyeRuleByName(ctx context.Context, team string, keyword string, cookie string) ([]*model.FindRuleDetail, error) {
	headers := s.RuleCookieHeader(cookie)
	reqURL := findRuleIdURL + fmt.Sprintf("team=%s&keyword=%s", team, url.QueryEscape(keyword))
	rst, err := httpGet(reqURL, headers)
	if err != nil {
		return nil, err
	}
	findRuleIdReply := &model.FindRuleIdReply{}
	if err = json.Unmarshal(rst, &findRuleIdReply); err != nil {
		return nil, err
	}
	if findRuleIdReply.Code != 0 {
		return nil, errors.New(findRuleIdReply.Message)
	}
	items := findRuleIdReply.Data.Items
	return items, nil
}

func (s *Service) HawkeyeGetReceiverGroup(ctx context.Context, url string, cookie string) ([]int64, error) {
	headers := s.RuleCookieHeader(cookie)
	rst, err := httpGet(url, headers)
	if err != nil {
		return nil, err
	}
	receiverGroupReply := &model.ReceiverGroupReply{}
	if err = json.Unmarshal(rst, &receiverGroupReply); err != nil {
		return nil, err
	}
	if receiverGroupReply.Code != 0 {
		return nil, errors.New(receiverGroupReply.Message)
	}
	items := receiverGroupReply.Data.Items
	var ids []int64
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (s *Service) HawkeyeSetReceiverGroup(ctx context.Context, req *model.PutReceiverGroupReq, cookie string) error {
	headers := s.RuleCookieHeader(cookie)
	if _, err := httpPut(putReceiverURL, req, headers); err != nil {
		return err
	}
	return nil
}

// UpdateRuleQuery update second query with given rule
func (s *Service) UpdateRuleQuery(ctx context.Context, ruleDetail *model.FindRuleDetail, query *model.RuleQuery, cookie string) error {
	updateRuleReq := new(model.UpdateRuleReq)
	err := copier.Copy(updateRuleReq, ruleDetail)
	if err != nil {
		return err
	}
	updateRuleReq.Querys[1] = query
	if err = s.UpdateHawkeyeRule(ctx, ruleDetail.ID, updateRuleReq, cookie); err != nil {
		return err
	}
	return nil
}

// UpdateRuleThreshold update the existed rule threshold
func (s *Service) UpdateRuleThreshold(ctx context.Context, team string, ruleId int64, threshold int64, cookie string) error {
	ruleDetail, err := s.HawkeyeRuleDetail(ctx, team, ruleId, cookie)
	if err != nil {
		_ = s.dao.DeleteCodeRule(ctx, ruleId)
		return errors.New("database error, please retry")
	}
	//nolint:gomnd
	if len(ruleDetail.Querys) < 2 {
		return errors.New("cannot match query num")
	}
	query := ruleDetail.Querys[1]
	query.Threshold = threshold
	if err = s.UpdateRuleQuery(ctx, ruleDetail, query, cookie); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

// NewRule insert new rule in Hawkeye
func (s *Service) NewRule(ctx context.Context, baseRuleId int64, req *model.ConfigRuleReq, cookie string) (string, error) {
	insertRuleReq := &model.InsertRuleReq{}
	ruleDetail, err := s.HawkeyeRuleDetail(ctx, req.Team, baseRuleId, cookie)
	if err != nil {
		return "", err
	}
	if err = copier.Copy(insertRuleReq, ruleDetail); err != nil {
		return "", err
	}
	insertRuleReq.Name = tag + insertRuleReq.Name + "｜" + req.Interface + "｜" + req.Code
	//nolint:gomnd
	if len(insertRuleReq.Querys) < 2 {
		return "", errors.New("cannot match query num")
	}
	query := insertRuleReq.Querys[1]
	for i, scope := range query.Scopes {
		if scope.LabelName == "code" {
			scope.LabelMatchers[0].MatcherType = "list"
			scope.LabelMatchers[0].Values = []string{req.Code}
		} else if scope.LabelName == "method" || scope.LabelName == "path" {
			scope.LabelMatchers[0].MatcherType = "list"
			scope.LabelMatchers[0].Values = []string{req.Interface}
		} else {
			continue
		}
		query.Scopes[i] = scope
	}
	query.Threshold = req.Threshold
	insertRuleReq.Querys[1] = query
	if err = s.InsertHawkeyeRule(ctx, insertRuleReq, cookie); err != nil {
		return "", err
	}
	return insertRuleReq.Name, nil
}

func (s *Service) ThresholdConfig(ctx context.Context, req *model.ConfigRuleReq, cookie string) error {
	codeRule := &model.CodeRule{
		Team:   req.Team,
		Type:   req.Type,
		Method: req.Interface,
		Code:   req.Code,
	}
	ruleId, err := s.dao.SelectRuleId(ctx, codeRule)
	if err != nil {
		log.Error("%+v", err)
		return xecode.Errorf(xecode.RequestErr, err.Error())
	}
	if ruleId == 0 {
		if err = s.NewRuleThreshold(ctx, req, cookie); err != nil {
			log.Error("%+v", err)
			return xecode.Errorf(xecode.RequestErr, err.Error())
		}
		return nil
	}
	if err = s.UpdateRuleThreshold(ctx, req.Team, ruleId, req.Threshold, cookie); err != nil {
		log.Error("%+v", err)
		return xecode.Errorf(xecode.RequestErr, err.Error())
	}
	return nil
}

// NewRuleThreshold add new rule and update base rule
func (s *Service) NewRuleThreshold(ctx context.Context, req *model.ConfigRuleReq, cookie string) error {
	baseRule := &model.CodeRule{
		Team:   req.Team,
		Type:   req.Type,
		Method: "base",
		Code:   "base",
	}
	baseRuleId, err := s.dao.SelectRuleId(ctx, baseRule)
	if err != nil {
		return err
	}
	if baseRuleId == 0 {
		baseRuleId, err = s.NewBaseRule(ctx, req, cookie)
		if err != nil {
			return err
		}
	}
	ruleName, err := s.NewRule(ctx, baseRuleId, req, cookie)
	if err != nil {
		return err
	}
	ruleId, err := s.HawkeyeRuleIdByName(ctx, req.Team, ruleName, cookie)
	if err != nil {
		return err
	}
	codeRule := &model.CodeRule{
		Team:   req.Team,
		Type:   req.Type,
		Method: req.Interface,
		Code:   req.Code,
		RuleId: ruleId,
	}
	if err = s.dao.InsertCodeRule(ctx, codeRule); err != nil {
		return nil
	}
	return nil
}

func (s *Service) NewBaseRule(ctx context.Context, req *model.ConfigRuleReq, cookie string) (int64, error) {
	baseRuleId, err := s.HawkeyeRuleIdByName(ctx, req.Team, req.Type, cookie)
	if err != nil {
		return -1, err
	}
	baseRule := &model.CodeRule{
		Team:   req.Team,
		Type:   req.Type,
		Method: "base",
		Code:   "base",
		RuleId: baseRuleId,
	}
	if err = s.dao.InsertCodeRule(ctx, baseRule); err != nil {
		return -1, err
	}
	return baseRuleId, nil
}

func (s *Service) CustomizedTeamRule(ctx context.Context, team string, cookie string) ([]*model.CustomizedTeamRuleReply, error) {
	rsts := make([]*model.CustomizedTeamRuleReply, 0)
	rules, err := s.HawkeyeRuleByName(ctx, team, tag, cookie)
	if err != nil {
		return nil, xecode.Errorf(xecode.RequestErr, err.Error())
	}
	for _, rule := range rules {
		name := strings.Split(rule.Name, "｜")
		code, err := strconv.ParseInt(name[2], 10, 64)
		//nolint:gomnd
		if len(rule.Querys) < 2 {
			return nil, errors.New("blank rule detail")
		}
		if err != nil {
			return nil, err
		}
		rst := &model.CustomizedTeamRuleReply{
			ID:        rule.ID,
			Team:      rule.Team,
			Type:      name[0],
			Method:    name[1],
			Code:      code,
			Threshold: rule.Querys[1].Threshold,
		}
		rsts = append(rsts, rst)
	}
	return rsts, nil
}

func (s *Service) EditRule(ctx context.Context, team string, id int64, threshold int64, cookie string) error {
	err := s.UpdateRuleThreshold(ctx, team, id, threshold, cookie)
	if err != nil {
		return xecode.Errorf(xecode.RequestErr, err.Error())
	}
	return nil
}

func (s *Service) DeleteRule(ctx context.Context, team string, id int64, cookie string) error {
	err := s.DeleteHawkeyeRule(ctx, team, id, cookie)
	if err != nil {
		return xecode.Errorf(xecode.RequestErr, err.Error())
	}
	_ = s.dao.DeleteCodeRule(ctx, id)
	return nil
}

func (s *Service) FetchRoleTree(ctx context.Context, cookie string) (*model.TreeReply, error) {
	treeAuthURL, err := url.Parse(authURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json;charset=UTF-8"
	headers["Cookie"] = cookie
	data, err := httpGet(treeAuthURL.String(), headers)
	if err != nil {
		return nil, err
	}
	result := &model.TokenResult{}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	//nolint:gomnd
	if result.Status != 200 {
		return nil, errors.Errorf("Failed to request tree token: %+v", result)
	}
	token := &model.Token{}
	if err = json.Unmarshal(result.Data, token); err != nil {
		return nil, errors.WithStack(err)
	}
	roleTreeURL, err := url.Parse(treeURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	headers = make(map[string]string)
	headers["Content-Type"] = "application/json;charset=UTF-8"
	headers["X-Authorization-Token"] = token.Token
	data, err = httpGet(roleTreeURL.String(), headers)
	if err != nil {
		return nil, err
	}
	res := &model.Resp{}
	if err = json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	reply := &model.TreeReply{}
	for _, data := range res.Data {
		p := strings.TrimPrefix(data.Path, "bilibili.")
		option := &model.Options{
			Value: p,
			Label: p,
		}
		reply.Options = append(reply.Options, option)
	}
	return reply, nil
}

func (s *Service) ExcludeAllReceiverGroups(ctx context.Context, cookie string) error {
	ids, err := s.HawkeyeGetReceiverGroup(ctx, ActiveReceiverURL, cookie)
	if err != nil {
		return nil
	}
	req := &model.PutReceiverGroupReq{
		GroupIds: ids,
		Status:   "excluded",
	}
	if err = s.HawkeyeSetReceiverGroup(ctx, req, cookie); err != nil {
		return err
	}
	return nil
}

func (s *Service) ActiveOwnerReceiverGroups(ctx context.Context, cookie string) error {
	username, err := s.GetCookieUsername(ctx, cookie)
	if err != nil {
		return err
	}
	groups, err := s.dao.GetUserService(ctx, username)
	if err != nil {
		return err
	}
	var ids []int64
	for _, group := range groups {
		url := fmt.Sprintf(ExcludedRecevicerURL, group)
		id, err := s.HawkeyeGetReceiverGroup(ctx, url, cookie)
		if err != nil {
			return err
		}
		ids = append(ids, id...)
	}
	req := &model.PutReceiverGroupReq{
		GroupIds: ids,
		Status:   "active",
	}
	if err = s.HawkeyeSetReceiverGroup(ctx, req, cookie); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetCookieUsername(ctx context.Context, cookie string) (string, error) {
	str := strings.Replace(cookie, " ", "", -1)
	strCookie := strings.Split(str, ";")
	for _, field := range strCookie {
		tmp := strings.Split(field, "=")
		if tmp[0] == "username" {
			return tmp[1], nil
		}
	}
	return "", errors.New("cannot find username from cookie")
}

func (s *Service) RootReceiverGroups(ctx context.Context, cookie string) error {
	if err := s.ExcludeAllReceiverGroups(ctx, cookie); err != nil {
		return xecode.Errorf(xecode.RequestErr, err.Error())
	}
	if err := s.ActiveOwnerReceiverGroups(ctx, cookie); err != nil {
		return xecode.Errorf(xecode.RequestErr, err.Error())
	}
	return nil
}

func (s *Service) MyService(ctx context.Context, cookie string) (*model.MyServiceReply, error) {
	username, err := s.GetCookieUsername(ctx, cookie)
	if err != nil {
		return nil, err
	}
	primary, err := s.dao.GetPrimaryService(ctx, username)
	if err != nil {
		return nil, err
	}
	strPrimary := strings.Join(primary, ",")
	if strPrimary != "" {
		strPrimary = "主负责：" + strPrimary
	}
	secondary, err := s.dao.GetSecondaryService(ctx, username)
	if err != nil {
		return nil, err
	}
	strSecondary := strings.Join(secondary, ",")
	if strSecondary != "" {
		strSecondary = "次要负责：" + strSecondary
	}
	reply := &model.MyServiceReply{
		Primary:   strPrimary,
		Secondary: strSecondary,
	}
	return reply, nil
}
