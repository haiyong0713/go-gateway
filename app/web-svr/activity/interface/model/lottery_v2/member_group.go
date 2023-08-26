package lottery

import (
	"fmt"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"strconv"
	"strings"
	"sync"
)

const (
	_ = iota
	// VipTypeNew 新大会员
	VipTypeNew
	// VipTypeOld 老大会员
	VipTypeOld
	// VipTypeAll 大会员（年度+月度）
	VipTypeAll
	// VipTypeYear 年度大会员
	VipTypeYear
	// VipTypeMonth 月度大会员
	VipTypeMonth
	// VipTypeNone 非大会员
	VipTypeNone
)

const (
	_ = iota
	// GroupTypeNewMember 新用户
	GroupTypeNewMember
	// GroupTypeVipMember 大会员
	GroupTypeVipMember
	// GroupTypeAction 行为
	GroupTypeAction
	// GroupTypeCartoon 漫画
	GroupTypeCartoon
	// GroupTypeMemberLevel 用户等级
	GroupTypeMemberLevel
)
const (
	// GroupTypeActionReserve 预约行为
	GroupTypeActionReserve = 1
	// MemberLevelSymbolMoreThan 用户组大于
	MemberLevelSymbolMoreThan = 1
	// MemberLevelSymbolLessThan 用户组小于
	MemberLevelSymbolLessThan = 2
)

const (
	// isNew 是新用户
	isNew = 1
)

// MemberNewInfo ...
type MemberNewInfo struct {
	Info map[int64]bool
	lock sync.RWMutex
}

// MemberIdsInfo ...
type MemberIdsInfo struct {
	Info map[int64]bool
	lock sync.RWMutex
}

// Set ...
func (m *MemberIdsInfo) Set(key int64, value bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Info[key] = value
}

// Set ...
func (m *MemberNewInfo) Set(key int64, value bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Info[key] = value
}

// MemberGroup 用户组
type MemberGroup struct {
	ID    int64    `json:"id"`
	Name  string   `json:"name"`
	Group []*Group `json:"group"`
}

// MemberGroupDB 用户组 DB
type MemberGroupDB struct {
	ID    int64      `json:"id"`
	SID   string     `json:"sid"`
	Name  string     `json:"name"`
	Group string     `json:"group"`
	State int        `json:"state"`
	Ctime xtime.Time `json:"ctime"`
	Mtime xtime.Time `json:"mtime"`
}

// GroupInterface ...
type GroupInterface interface {
	Init(group *Group) error
	Check(member *MemberInfo, memberNewInfo *MemberNewInfo, memberIdsInfo *MemberIdsInfo) error
}

// Group ...
type Group struct {
	GroupType int                    `json:"group_type"`
	Params    map[string]interface{} `json:"params"`
}

// GroupNewMember 新用户组
type GroupNewMember struct {
	ID     int64 `json:"id"`
	IsNew  int   `json:"is_new"`
	Period int64 `json:"period"`
}

// GroupVip 大会员类型
type GroupVip struct {
	VipType int `json:"vip_type"`
}

// GroupAction 行为类型
type GroupAction struct {
	Action int     `json:"action"`
	IdsDB  string  `json:"ids"`
	Ids    []int64 `json:"_"`
}

// GroupMemberLevel 用户等级
type GroupMemberLevel struct {
	Level  int `json:"level"`
	Symbol int `json:"symbol"`
}

// GroupCartoon 漫画类型
type GroupCartoon struct {
}

// GetMemberGroup ...
func GetMemberGroup(group *Group) (GroupInterface, error) {
	switch group.GroupType {
	case GroupTypeNewMember:
		g := &GroupNewMember{}
		g.Init(group)
		return g, nil
	case GroupTypeVipMember:
		g := &GroupVip{}
		g.Init(group)
		return g, nil
	case GroupTypeAction:
		g := &GroupAction{}
		g.Init(group)
		return g, nil
	case GroupTypeCartoon:
		g := &GroupCartoon{}
		g.Init(group)
		return g, nil
	case GroupTypeMemberLevel:
		g := &GroupMemberLevel{}
		g.Init(group)
		return g, nil

	default:
		return nil, ecode.ActivityLotteryTimesTypeError
	}
}

// Init ...
func (g *GroupNewMember) Init(group *Group) error {
	params := group.Params
	isNew, ok := params["is_new"]
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	isNewStr := fmt.Sprintf("%v", isNew)
	isNewInt, err := strconv.Atoi(isNewStr)
	if err != nil {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	var periodInt int64
	g.IsNew = isNewInt
	period, ok := params["period"]
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	switch period.(type) {
	case float64:
		periodInt = int64(period.(float64))
	case int:
		periodInt = int64(period.(int))
	case int64:
		periodInt = period.(int64)
	default:
		return ecode.ActivityLotteryMemberGroupStructError

	}
	g.Period = periodInt
	return nil
}

// Check ...
func (g *GroupNewMember) Check(member *MemberInfo, memberNewInfo *MemberNewInfo, memberIdsInfo *MemberIdsInfo) error {
	if res, ok := memberNewInfo.Info[g.Period]; ok {
		if g.IsNew == isNew {
			if res {
				return nil
			}
			return ecode.ActivityLotteryMemberNotNewError
		}
		if !res {
			return nil
		}
		return ecode.ActivityLotteryMemberNotOldError
	}
	return ecode.ActivityLotteryMemberInfoError
}

// Init ...
func (g *GroupVip) Init(group *Group) error {
	params := group.Params
	vipType, ok := params["vip_type"]
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	vipTypeStr := fmt.Sprintf("%v", vipType)
	vipTypeInt, err := strconv.Atoi(vipTypeStr)
	if err != nil {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	g.VipType = vipTypeInt
	return nil
}

// Check ...
func (g *GroupVip) Check(member *MemberInfo, memberNewInfo *MemberNewInfo, memberIdsInfo *MemberIdsInfo) error {
	switch g.VipType {
	case VipTypeNew:
		return member.IsNewVip()
	case VipTypeOld:
		return member.IsOldVip()
	case VipTypeAll:
		return member.IsVip()
	case VipTypeYear:
		return member.IsAnnualVip()
	case VipTypeMonth:
		return member.IsMonthVip()
	case VipTypeNone:
		return member.IsNotVip()
	default:
		return ecode.ActivityLotteryMemberGroupVipTypeError
	}
}

// Init ...
func (g *GroupAction) Init(group *Group) error {
	params := group.Params
	var actionInt int
	action, ok := params["action"]
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	switch action.(type) {
	case float64:
		actionInt = int(action.(float64))
	case int:
		actionInt = int(action.(int))
	case int64:
		actionInt = action.(int)
	default:
		return ecode.ActivityLotteryMemberGroupStructError

	}
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	g.Action = actionInt
	idsDB, ok := params["ids"].(string)
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	idsStr := strings.Split(idsDB, ",")
	for _, v := range idsStr {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		g.Ids = append(g.Ids, id)
	}
	return nil
}

// Check ...
func (g *GroupAction) Check(member *MemberInfo, memberNewInfo *MemberNewInfo, memberIdsInfo *MemberIdsInfo) error {
	if g.Action == GroupTypeActionReserve {
		if g.Ids != nil {
			if memberIdsInfo == nil {
				return ecode.ActivityLotteryMemberGroupNotReserveError
			}
			for _, v := range g.Ids {
				if idsInfo, ok := memberIdsInfo.Info[v]; !ok || !idsInfo {
					return ecode.ActivityLotteryMemberGroupNotReserveError
				}
			}
		}
	}
	return nil
}

// Init ...
func (g *GroupCartoon) Init(group *Group) error {
	return nil
}

// Check ...
func (g *GroupCartoon) Check(member *MemberInfo, memberNewInfo *MemberNewInfo, memberIdsInfo *MemberIdsInfo) error {
	if member.IsCartoonNew {
		return nil
	}
	return ecode.ActivityLotteryMemberGroupNotCartoonNewError
}

// Init ...
func (g *GroupMemberLevel) Init(group *Group) error {
	params := group.Params
	level, ok := params["level"]
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	switch level.(type) {
	case float64:
		g.Level = int(level.(float64))
	case int:
		g.Level = int(level.(int))
	case int64:
		g.Level = level.(int)
	default:
		return ecode.ActivityLotteryMemberGroupStructError

	}
	symbol, ok := params["symbol"]
	if !ok {
		return ecode.ActivityLotteryMemberGroupStructError
	}
	switch symbol.(type) {
	case float64:
		g.Symbol = int(symbol.(float64))
	case int:
		g.Symbol = int(symbol.(int))
	case int64:
		g.Symbol = symbol.(int)
	default:
		return ecode.ActivityLotteryMemberGroupStructError
	}
	return nil
}

// Check ...
func (g *GroupMemberLevel) Check(member *MemberInfo, memberNewInfo *MemberNewInfo, memberIdsInfo *MemberIdsInfo) error {
	if g.Symbol == MemberLevelSymbolMoreThan {
		if member.Level > int32(g.Level) {
			return nil
		}
		return ecode.ActivityLotteryMemberGroupMemberLevelError
	}
	if g.Symbol == MemberLevelSymbolLessThan {
		if member.Level < int32(g.Level) {
			return nil
		}
		return ecode.ActivityLotteryMemberGroupMemberLevelError
	}
	return nil
}
