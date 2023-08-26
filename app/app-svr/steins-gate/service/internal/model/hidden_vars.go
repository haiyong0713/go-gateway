package model

import (
	"encoding/json"
	"math/rand"
	"sort"

	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

// HiddenVar def.
type HiddenVar struct {
	Value float64 `json:"value"`
	HiddenVarCore
}

type HiddenVarCore struct {
	ID            string `json:"id"`
	IDV2          string `json:"id_v2"`
	Type          int    `json:"type"`
	IsShow        int    `json:"is_show"`
	Name          string `json:"name"`
	SkipOverwrite int    `json:"skip_overwrite"`
}

// HiddenVarInt value为int型，兼容老版本输出
type HiddenVarInt struct {
	Value int64 `json:"value"`
	*HiddenVarCore
}

func (v *HiddenVarInt) FromHVar(hvar *HiddenVar) {
	v.HiddenVarCore = &hvar.HiddenVarCore
	v.Value = int64(hvar.Value)
}

func (v *HiddenVarCore) GetIDV2() {
	v.IDV2 = hvarIDToExpr(v.ID)
}

// RandomVar def.
type RandomVar struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	IDV2   string `json:"id_v2"`
	Value  int    `json:"value"`
	IsShow int    `json:"is_show"`
}

func (out *HiddenVar) IsDisplay() bool {
	return out.IsShow == 1
}

func (out *RandomVar) IsDisplay() bool {
	return out.IsShow == 1
}

func (out *HiddenVar) FromRandomVar(in *RandomVar) {
	out.Value = float64(in.Value)
	out.Type = RegionalVarTypeRandom
	out.ID = in.ID
	out.IDV2 = hvarIDToExpr(out.ID)
	out.Name = in.Name
	out.IsShow = in.IsShow
}

func (out *HiddenVar) FromRegionalVar(in *RegionalVal) {
	out.Value = float64(in.InitMin)
	out.Type = in.Type
	out.ID = in.ID
	out.IDV2 = hvarIDToExpr(out.ID)
	out.Name = in.Name
	out.IsShow = in.IsShow
}

// HiddenVarsRecord
type HiddenVarsRecord struct {
	Vars map[string]*HiddenVar `json:"vars"`
}

type HiddenVarRec struct {
	ID        int64  `json:"id"`
	MID       int64  `json:"mid"`
	GraphID   int64  `json:"graph_id"`
	CurrentID int64  `json:"current_id"`
	CursorID  int64  `json:"cursor"`
	Value     string `json:"value"`
}

// Show transforms the hidden variables' record into a string
func (out *HiddenVarsRecord) Show() string {
	if len(out.Vars) > 0 {
		varStr, _ := json.Marshal(out.Vars)
		return string(varStr)
	}
	return ""
}

func BuildChoicesNode(in []*api.GraphEdge, skinInfo *api.Skin) (out []*Choice) {
	for k, v := range in {
		ch := new(Choice)
		ch.ForNode(k, v, skinInfo)
		out = append(out, ch)
	}
	return
}

func BuildChoicesEdge(in []*api.GraphEdge, skinInfo *api.Skin) (out []*Choice) {
	for k, v := range in {
		ch := new(Choice)
		ch.ForEdge(k, v, skinInfo)
		out = append(out, ch)
	}
	return
}

// FilterEdges filters the edges according to the user's record
//
//nolint:gocognit
func FilterEdges(record *HiddenVarsRecord, in []*api.GraphEdge, randomV *RandomVar) (out []*api.GraphEdge, err error) {
	if len(in) == 0 {
		return
	}
	for _, v := range in {
		var canAppend = true
		if len(v.Condition) != 0 {
			var conds []*EdgeCondition
			if err = json.Unmarshal([]byte(v.Condition), &conds); err != nil {
				return
			}
			for _, cond := range conds {
				var currentValue float64
				if randomV != nil && cond.VarID == randomV.ID { // 如果为随机变量，则使用随机值
					currentValue = float64(randomV.Value)
				} else if record == nil { // 未登录变量值按0处理
					currentValue = 0
				} else { // 条件中出现未知变量不处理
					rec, ok := record.Vars[cond.VarID]
					if !ok {
						continue
					}
					currentValue = float64(rec.Value)
				}
				switch cond.Condition {
				case EdgeConditionTypeGe:
					if currentValue < float64(cond.Value) {
						canAppend = false
						break
					}
				case EdgeConditionTypeGt:
					if currentValue <= float64(cond.Value) {
						canAppend = false
						break
					}
				case EdgeConditionTypeEq:
					if currentValue != float64(cond.Value) {
						canAppend = false
						break
					}
				case EdgeConditionTypeLt:
					if currentValue >= float64(cond.Value) {
						canAppend = false
						break
					}
				case EdgeConditionTypeLe:
					if currentValue > float64(cond.Value) {
						canAppend = false
						break
					}
				}
			}
		}
		if canAppend {
			out = append(out, v)
		}
	}
	return
}

// 判断所有条件是不是都不能出
func IsQuestionProm(record *HiddenVarsRecord, randomV *RandomVar, in []*Question) (out bool, err error) {
	if len(in) != 1 { // 非中插树应该只有一个question
		return
	}
	out = true // 默认都不能出
	rec := new(HiddenVarsRecord)
	rec.DeepCopy(record)
	rec.Vars[randomV.ID] = &HiddenVar{ // 需要验证隐藏变量是不是也满足
		Value: float64(randomV.Value),
		HiddenVarCore: HiddenVarCore{
			ID:   randomV.ID,
			IDV2: randomV.IDV2,
			Name: randomV.Name,
		},
	}
	for _, v := range in {
		if len(v.Choices) == 0 { // 如果没有选项也不上报
			return false, nil
		}
		for _, choice := range v.Choices {
			if choice.Condition == "" { // 如果有一个选项的Condition为空，也不用上报
				return false, nil
			}
			if result, err := rec.ExpressionCondition(choice.Condition); err == nil && result { // 只要有一个选项条件能过，就认为能跳转，不用上报
				return false, nil
			}
		}
	}
	return
}

// ApplyAttr applies the edge's attribute into the hidden variables record
func (out *HiddenVarsRecord) ApplyAttr(attrStr string) (err error) {
	if len(attrStr) != 0 {
		var attributes []*EdgeAttribute
		if err = json.Unmarshal([]byte(attrStr), &attributes); err != nil {
			return
		}
		for _, attr := range attributes {
			if _, ok := out.Vars[attr.VarID]; !ok { // 如果attribute中含有graph中未声明的变量，不处理
				continue
			}
			out.Vars[attr.VarID].MergeAttribute(attr) // 合入当前节点的action
		}
	}
	return
}

func (out *HiddenVarsRecord) DeepCopy(in *HiddenVarsRecord) {
	out.Vars = make(map[string]*HiddenVar, len(in.Vars))
	for k, v := range in.Vars {
		rec := new(HiddenVar)
		*rec = *v
		out.Vars[k] = rec
	}
}

// MergeAttribute merges an attribute into the hidden var value
func (out *HiddenVar) MergeAttribute(a *EdgeAttribute) {
	if a.VarID != out.ID {
		return
	}
	switch a.Action {
	case EdgeAttrActionSub:
		out.Value -= float64(a.Value)
	case EdgeAttrActionAdd:
		out.Value += float64(a.Value)
	case EdgeAttrActionAssign:
		out.Value = float64(a.Value)
	}
}

func GetVarsMap(graph *api.GraphInfo) (res map[string]*RegionalVal, err error) {
	if graph == nil {
		return
	}
	if len(graph.RegionalVars) == 0 { // 当前图无隐藏变量
		err = ecode.GraphHiddenVarRecordNilErr
		return
	}
	var hvars []*RegionalVal
	if err = json.Unmarshal([]byte(graph.RegionalVars), &hvars); err != nil {
		return
	}
	res = make(map[string]*RegionalVal, len(hvars))
	for _, v := range hvars {
		res[v.ID] = v
	}
	return
}

func RandInt(r *rand.Rand, min, max int) int {
	if min >= max || max == 0 {
		return max
	}
	return r.Intn(max-min+1) + min // 左闭右闭
}

// GenerateNodeRecs 对于node图生成隐藏变量存档集合
func (attrsCache *EdgeAttrsCache) GenerateEdgeRecs(edgeIDs []int64, initRec *HiddenVarsRecord) (recs map[int64]*HiddenVarsRecord) {
	recs = make(map[int64]*HiddenVarsRecord)
	if !attrsCache.HasAttrs {
		return
	}
	attrs := make(map[int64]string)
	for _, v := range attrsCache.EdgeAttrs {
		attrs[v.ID] = v.Attribute
	}
	for _, edgeID := range edgeIDs {
		if edgeID == RootEdge { // 根节点处理
			continue
		}
		if atr, ok := attrs[edgeID]; ok {
			//nolint:errcheck
			initRec.ApplyAttr(atr)
		}
		newRec := new(HiddenVarsRecord)
		newRec.DeepCopy(initRec)
		recs[edgeID] = newRec
	}
	return
}

// GenerateNodeRecs 对于node图生成隐藏变量存档集合
func (attrsCache *EdgeAttrsCache) GenerateNodeRecs(nodeIDs []int64, initRec *HiddenVarsRecord) (recs map[int64]*HiddenVarsRecord) {
	recs = make(map[int64]*HiddenVarsRecord)       // 防止返回nil
	if !attrsCache.HasAttrs || len(nodeIDs) <= 1 { // 一个nodeIDs无法构成边
		return
	}
	attrs := make(map[int64]map[int64]string)
	for _, v := range attrsCache.EdgeAttrs { // 按照fromNID和toNID整理attributes
		if _, ok := attrs[v.FromNID]; !ok {
			attrs[v.FromNID] = make(map[int64]string)
		}
		attrs[v.FromNID][v.ToNID] = v.Attribute
	}
	for i := 0; i < len(nodeIDs)-1; i++ {
		fromN := nodeIDs[i]
		toN := nodeIDs[i+1]
		if ats, okFrom := attrs[fromN]; okFrom {
			if atr, okTo := ats[toN]; okTo {
				//nolint:errcheck
				initRec.ApplyAttr(atr)
			}
			newRec := new(HiddenVarsRecord)
			newRec.DeepCopy(initRec)
			recs[toN] = newRec
		}
	}
	return
}

// DisplayHvars 展示隐藏变量
func DisplayHvars(hvarRec *HiddenVarsRecord, randomV *RandomVar) (results []*HiddenVar) {
	if hvarRec == nil {
		return
	}
	for _, v := range hvarRec.Vars { // map to slice
		v.GetIDV2()
		results = append(results, v)
	}
	if randomV != nil { // 增加展示随机变量
		rv := new(HiddenVar)
		rv.FromRandomVar(randomV)
		results = append(results, rv)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID }) // 对返回列表排序
	return
}

// FromRecNode 处理node的存档为隐藏变量request
func FromRecNode(lastRec *api.GameRecords, portal int32, firstID, pickID, currentCursor int64, inChoices, inCursorChoices string) (v *HvarReq) {
	v = new(HvarReq)
	v.fromRecordCore(lastRec, portal, firstID, pickID, currentCursor, inChoices, inCursorChoices)
	if portal == 0 && lastRec != nil {
		v.CurrentID = lastRec.CurrentNode
	}
	return
}

// FromRecEdge 处理node的存档为隐藏变量request
func FromRecEdge(lastRec *api.GameRecords, portal int32, firstID, pickID, currentCursor int64, inChoices, inCursorChoices string) (v *HvarReq) {
	v = new(HvarReq)
	v.fromRecordCore(lastRec, portal, firstID, pickID, currentCursor, inChoices, inCursorChoices)
	if portal == 0 && lastRec != nil {
		v.CurrentID = lastRec.CurrentEdge
	}
	return
}

// FromRecHandler专职处理存档为隐藏变量request，保持隐藏变量对edge和node的不敏感

type FromRecHandler func(lastRec *api.GameRecords, portal int32, firstID, pickID, currentCursor int64, inChoices, inCursorChoices string) (v *HvarReq)
