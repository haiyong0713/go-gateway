package open

import (
	"context"
	"encoding/json"

	openmdl "go-gateway/app/app-svr/fawkes/service/model/open"
	"go-gateway/app/app-svr/fawkes/service/service/casbin"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (

	// AuthAddEvent 权限新增事件
	AuthAddEvent = "inner.auth.add"
	// AuthUpdateEvent 权限更新事件
	AuthUpdateEvent = "inner.auth.update"
	// AuthDeleteEvent 权限删除事件
	AuthDeleteEvent = "inner.auth.delete"
)

type AuthAddArg struct {
	PT []*PathToken
}

type AuthDeleteArg struct {
	PT []*PathToken
}

type AuthUpdateArg struct {
	OldPT []*PathToken
	NewPT []*PathToken
}

type PathToken struct {
	Token  string
	Path   string
	AppKey []string
}

// PathAuthAddAction 新增权限
func (s *Service) PathAuthAddAction(ctx context.Context, args AuthAddArg) (err error) {
	e := casbin.GetInstance()
	rules := genPolicyRules(args.PT)
	ok, err := e.AddPolicies(rules)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if !ok {
		marshal, _ := json.Marshal(args)
		log.Errorc(ctx, "add policies failed %v", marshal)
		return
	}
	return
}

// PathAuthUpdateAction 更新权限
func (s *Service) PathAuthUpdateAction(ctx context.Context, args AuthUpdateArg) (err error) {
	e := casbin.GetInstance()
	oldRules := genPolicyRules(args.OldPT)
	newRules := genPolicyRules(args.NewPT)
	removeOk, removeErr := e.RemovePolicies(oldRules)
	if removeErr != nil {
		log.Errorc(ctx, "remove err %v", removeErr)
		return
	}
	addOk, addErr := e.AddPolicies(newRules)
	if addErr != nil {
		log.Errorc(ctx, "add err %v", addErr)
		return
	}
	if !removeOk || !addOk {
		oldMarshal, _ := json.Marshal(oldRules)
		newMarshal, _ := json.Marshal(newRules)
		log.Errorc(ctx, "update policies failed oldRules: %v\nnewRules: %v", oldMarshal, newMarshal)
		return
	}
	return
}

// PathAuthDeleteAction 删除权限
func (s *Service) PathAuthDeleteAction(ctx context.Context, args AuthDeleteArg) (err error) {
	e := casbin.GetInstance()
	rules := genPolicyRules(args.PT)
	ok, err := e.RemovePolicies(rules)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if !ok {
		marshal, _ := json.Marshal(args)
		log.Errorc(ctx, "update policies failed %v", marshal)
		return
	}
	return
}

func genPolicyRules(pt []*PathToken) (oldP [][]string) {
	var rules [][]string
	for _, v := range pt {
		var r []string
		r = append(r, v.Token, v.Path)
		if len(v.AppKey) == 0 {
			r = append(r, openmdl.AnyAppKey)
			rules = append(rules, r)
		} else {
			for _, ak := range v.AppKey {
				tmp := r
				tmp = append(tmp, ak)
				rules = append(rules, tmp)
			}
		}
	}
	return rules
}
