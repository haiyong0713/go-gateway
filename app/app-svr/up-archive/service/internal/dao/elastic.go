package dao

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/database/elastic"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_businessArcSearch = "gateway_archive_search"
	_businessArcList   = "gateway_archive_list"
	_index             = "gateway_archive_search"
)

func businessKey(keyword string) string {
	if keyword != "" {
		return _businessArcSearch
	}
	return _businessArcList
}

func NewElastic() (ela *elastic.Elastic, cf func(), err error) {
	var cfg struct {
		Search struct {
			Client *bm.ClientConfig
			Host   string
		}
	}
	if err = paladin.Get("elastic.toml").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	ela = elastic.NewElastic(&elastic.Config{
		Host:       cfg.Search.Host,
		HTTPClient: cfg.Search.Client,
	})
	cf = func() {}
	return
}

// nolint:gocognit
func (d *dao) arcPassedSearch(ctx context.Context, mid, tid int64, keyword string, kwFields []string, highlight bool, pn, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcPassedSearchReply, error) {
	r := d.elastic.NewRequest(businessKey(keyword)).Index(_index)
	eq := []map[string]interface{}{
		{"mid": mid},
		{"staff_mid": mid},
	}
	//!(medl_id==59 && state=1)
	nots := []map[string]interface{}{
		{
			"archive_flow.meal_id": d.noSpace,
			"archive_flow.state":   1,
		},
	}
	for _, val := range without {
		switch val {
		case api.Without_staff:
			eq = []map[string]interface{}{
				{"mid": mid},
			}
		case api.Without_live_playback:
			r.WhereIn("up_from", d.livePlaybackUpFrom)
			r.WhereNot(elastic.NotTypeIn, "up_from")
		case api.Without_no_space:
			//!(medl_id==59 && state=1) && !(medl_id==60 && state=1)
			nots = []map[string]interface{}{
				{
					"archive_flow.meal_id": d.noSpace,
					"archive_flow.state":   1,
				},
				{
					"archive_flow.meal_id": d.upNoSpace,
					"archive_flow.state":   1,
				},
			}
		default:
		}
	}
	comboEq := &elastic.Combo{}
	comboEq.ComboEQ(eq).MinEQ(1).MinAll(1)
	comboNots := &elastic.Combo{}
	comboNots.ComboNestedNots("archive_flow", nots).MinNested(len(nots)).MinAll(1)
	r.WhereCombo(comboEq, comboNots)
	r.WhereRange("state", 0, nil, elastic.RangeScopeLcRc)
	if tid != 0 {
		r.WhereEq("pid", tid)
	}
	if keyword != "" {
		r.WhereLike(kwFields, []string{keyword}, true, elastic.LikeLevelLow).Highlight(highlight)
	}
	if len(d.notAttrs) != 0 {
		r.WhereIn("attribute", d.notAttrs)
		r.WhereNot(elastic.NotTypeIn, "attribute")
	}
	if len(d.notAttrV2s) != 0 {
		r.WhereIn("attribute_v2", d.notAttrV2s)
		r.WhereNot(elastic.NotTypeIn, "attribute_v2")
	}
	r.Order(order.String(), sort.String())
	r.Order("aid", sort.String())
	r.OrderScoreFirst(false)
	r.Pn(pn).Ps(ps)
	log.Info("ArcPassedSearch elastic query:%v", r.Params())
	var result *model.ArcPassedSearchReply
	if err := r.Scan(ctx, &result); err != nil {
		return nil, errors.Wrap(err, r.Params())
	}
	if result == nil {
		return nil, nil
	}
	reply := &model.ArcPassedSearchReply{Page: result.Page}
	for _, val := range result.Result {
		if val == nil || val.Aid <= 0 {
			continue
		}
		reply.Result = append(reply.Result, val)
	}
	if keyword != "" && highlight {
		hm := map[int64]*model.ArcPassedHighlight{}
		for i := 0; i < len(result.Result)-1; i += 2 {
			if result.Result[i] == nil || result.Result[i+1] == nil {
				continue
			}
			var titles, contents []string
			tData := result.Result[i+1].Title
			if tData != nil {
				if err := json.Unmarshal(tData, &titles); err != nil {
					log.Error("ArcPassedSearch query:%v,error:%+v", r.Params(), err)
				}
			}
			cData := result.Result[i+1].Content
			if cData != nil {
				if err := json.Unmarshal(cData, &contents); err != nil {
					log.Error("ArcPassedSearch query:%v,error:%+v", r.Params(), err)
				}
			}
			var title, content string
			if len(titles) != 0 {
				title = titles[0]
			}
			if title == "" && len(result.Result[i+1].TitleItem) != 0 {
				title = result.Result[i+1].TitleItem[0]
			}
			if len(contents) != 0 {
				content = contents[0]
			}
			if title == "" && content == "" {
				continue
			}
			hm[result.Result[i].Aid] = &model.ArcPassedHighlight{Title: title, Content: content}
		}
		for _, val := range reply.Result {
			val.Highlight = hm[val.Aid]
		}
	}
	return reply, nil
}

func (d *dao) arcPassedSearchTag(ctx context.Context, mid int64, keyword string, kwFields []string, without []api.Without) (map[int64]int64, error) {
	r := d.elastic.NewRequest(businessKey(keyword)).Index(_index)
	eq := []map[string]interface{}{
		{"mid": mid},
		{"staff_mid": mid},
	}
	//!(medl_id==59 && state=1)
	nots := []map[string]interface{}{
		{
			"archive_flow.meal_id": d.noSpace,
			"archive_flow.state":   1,
		},
	}
	for _, val := range without {
		switch val {
		case api.Without_staff:
			eq = []map[string]interface{}{
				{"mid": mid},
			}
		case api.Without_live_playback:
			r.WhereIn("up_from", d.livePlaybackUpFrom)
			r.WhereNot(elastic.NotTypeIn, "up_from")
		case api.Without_no_space:
			//!(medl_id==59 && state=1) && !(medl_id==60 && state=1)
			nots = []map[string]interface{}{
				{
					"archive_flow.meal_id": d.noSpace,
					"archive_flow.state":   1,
				},
				{
					"archive_flow.meal_id": d.upNoSpace,
					"archive_flow.state":   1,
				},
			}
		default:
		}
	}
	comboEq := &elastic.Combo{}
	comboEq.ComboEQ(eq).MinEQ(1).MinAll(1)
	comboNots := &elastic.Combo{}
	comboNots.ComboNestedNots("archive_flow", nots).MinNested(len(nots)).MinAll(1)
	r.WhereCombo(comboEq, comboNots)
	r.WhereRange("state", 0, nil, elastic.RangeScopeLcRc)
	if keyword != "" {
		r.WhereLike(kwFields, []string{keyword}, true, elastic.LikeLevelLow)
	}
	if len(d.notAttrs) != 0 {
		r.WhereIn("attribute", d.notAttrs)
		r.WhereNot(elastic.NotTypeIn, "attribute")
	}
	if len(d.notAttrV2s) != 0 {
		r.WhereIn("attribute_v2", d.notAttrV2s)
		r.WhereNot(elastic.NotTypeIn, "attribute_v2")
	}
	r.GroupBy(elastic.EnhancedModeGroupBy, "pid", nil)
	log.Info("ArcPassedSearchTag elastic query:%v", r.Params())
	var result *struct {
		Result map[string][]struct {
			Key   string `json:"key"`
			Count int64  `json:"doc_count"`
		} `json:"result"`
	}
	if err := r.Scan(ctx, &result); err != nil {
		return nil, errors.Wrap(err, r.Params())
	}
	if result == nil {
		return nil, nil
	}
	reply := map[int64]int64{}
	if vals, ok := result.Result["group_by_pid"]; ok {
		for _, val := range vals {
			tid, _ := strconv.ParseInt(val.Key, 10, 64)
			if tid <= 0 {
				continue
			}
			reply[tid] = val.Count
		}
	}
	return reply, nil
}

func (d *dao) arcPassedSearchCursor(ctx context.Context, mid, score int64, containScore bool, ps int, without []api.Without, sort api.Sort) (*model.ArcPassedSearchReply, error) {
	r := d.elastic.NewRequest(businessKey("")).Index(_index)
	eq := []map[string]interface{}{
		{"mid": mid},
		{"staff_mid": mid},
	}
	//!(medl_id==59 && state=1)
	nots := []map[string]interface{}{
		{
			"archive_flow.meal_id": d.noSpace,
			"archive_flow.state":   1,
		},
	}
	for _, val := range without {
		switch val {
		case api.Without_staff:
			eq = []map[string]interface{}{
				{"mid": mid},
			}
		case api.Without_live_playback:
			r.WhereIn("up_from", d.livePlaybackUpFrom)
			r.WhereNot(elastic.NotTypeIn, "up_from")
		case api.Without_no_space:
			//!(medl_id==59 && state=1) && !(medl_id==60 && state=1)
			nots = []map[string]interface{}{
				{
					"archive_flow.meal_id": d.noSpace,
					"archive_flow.state":   1,
				},
				{
					"archive_flow.meal_id": d.upNoSpace,
					"archive_flow.state":   1,
				},
			}
		default:
		}
	}
	comboEq := &elastic.Combo{}
	comboEq.ComboEQ(eq).MinEQ(1).MinAll(1)
	comboNots := &elastic.Combo{}
	comboNots.ComboNestedNots("archive_flow", nots).MinNested(len(nots)).MinAll(1)
	r.WhereCombo(comboEq, comboNots)
	if score > 0 {
		pubtime := time.Unix(score, 0).Format("2006-01-02 15:04:05")
		switch sort {
		case api.Sort_desc:
			if containScore {
				r.WhereRange("pubtime", nil, pubtime, elastic.RangeScopeLcRc)
			} else {
				r.WhereRange("pubtime", nil, pubtime, elastic.RangeScopeLoRo)
			}
		case api.Sort_asc:
			if containScore {
				r.WhereRange("pubtime", pubtime, nil, elastic.RangeScopeLcRc)
			} else {
				r.WhereRange("pubtime", pubtime, nil, elastic.RangeScopeLoRo)
			}
		}
	}
	r.WhereRange("state", 0, nil, elastic.RangeScopeLcRc)
	if len(d.notAttrs) != 0 {
		r.WhereIn("attribute", d.notAttrs)
		r.WhereNot(elastic.NotTypeIn, "attribute")
	}
	if len(d.notAttrV2s) != 0 {
		r.WhereIn("attribute_v2", d.notAttrV2s)
		r.WhereNot(elastic.NotTypeIn, "attribute_v2")
	}
	r.Order("pubtime", sort.String())
	r.OrderScoreFirst(false)
	r.Pn(1).Ps(ps)
	log.Info("ArcPassedSearchCursor elastic query:%v", r.Params())
	var result *model.ArcPassedSearchReply
	if err := r.Scan(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *dao) arcPassedSearchCursorAid(ctx context.Context, mid, score int64, equalScore bool, aid, tid int64, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcCursorAidSearchReply, error) {
	r := d.elastic.NewRequest(businessKey("")).Index(_index)
	eq := []map[string]interface{}{
		{"mid": mid},
		{"staff_mid": mid},
	}
	//!(medl_id==59 && state=1)
	nots := []map[string]interface{}{
		{
			"archive_flow.meal_id": d.noSpace,
			"archive_flow.state":   1,
		},
	}
	for _, val := range without {
		switch val {
		case api.Without_staff:
			eq = []map[string]interface{}{
				{"mid": mid},
			}
		case api.Without_live_playback:
			r.WhereIn("up_from", d.livePlaybackUpFrom)
			r.WhereNot(elastic.NotTypeIn, "up_from")
		case api.Without_no_space:
			//!(medl_id==59 && state=1) && !(medl_id==60 && state=1)
			nots = []map[string]interface{}{
				{
					"archive_flow.meal_id": d.noSpace,
					"archive_flow.state":   1,
				},
				{
					"archive_flow.meal_id": d.upNoSpace,
					"archive_flow.state":   1,
				},
			}
		default:
		}
	}
	comboEq := &elastic.Combo{}
	comboEq.ComboEQ(eq).MinEQ(1).MinAll(1)
	comboNots := &elastic.Combo{}
	comboNots.ComboNestedNots("archive_flow", nots).MinNested(len(nots)).MinAll(1)
	r.WhereCombo(comboEq, comboNots)
	if tid != 0 {
		r.WhereEq("pid", tid)
	}
	rangeFunc := func(field string, data interface{}, scope model.Scope, sort api.Sort) {
		switch scope {
		case model.ScopeRange:
			switch sort {
			case api.Sort_desc:
				r.WhereRange(field, nil, data, elastic.RangeScopeLoRo)
			case api.Sort_asc:
				r.WhereRange(field, data, nil, elastic.RangeScopeLoRo)
			default:
			}
		case model.ScopeRangeContain:
			switch sort {
			case api.Sort_desc:
				r.WhereRange(field, nil, data, elastic.RangeScopeLcRc)
			case api.Sort_asc:
				r.WhereRange(field, data, nil, elastic.RangeScopeLcRc)
			default:
			}
		case model.ScopeEqual:
			r.WhereRange(field, data, data, elastic.RangeScopeLcRc)
		default:
		}
	}
	var scoreVal interface{}
	switch order {
	case api.SearchOrder_pubtime:
		scoreVal = time.Unix(score, 0).Format("2006-01-02 15:04:05")
	default:
		scoreVal = score
	}
	// desc
	// aid=0;不过滤
	// aid=n,n>0,score=m;(score==m&&aid=<n)||score<m
	// asc
	// aid=0;不过滤
	// aid=n,n>0,score=m;(score==m&&aid>=n)||score>m
	if aid > 0 {
		if equalScore {
			rangeFunc(order.String(), scoreVal, model.ScopeEqual, sort)
			rangeFunc("aid", aid, model.ScopeRangeContain, sort)
		} else {
			rangeFunc(order.String(), scoreVal, model.ScopeRange, sort)
		}
	}
	r.WhereRange("state", 0, nil, elastic.RangeScopeLcRc)
	if len(d.notAttrs) != 0 {
		r.WhereIn("attribute", d.notAttrs)
		r.WhereNot(elastic.NotTypeIn, "attribute")
	}
	if len(d.notAttrV2s) != 0 {
		r.WhereIn("attribute_v2", d.notAttrV2s)
		r.WhereNot(elastic.NotTypeIn, "attribute_v2")
	}
	r.Order(order.String(), sort.String())
	r.Order("aid", sort.String())
	r.OrderScoreFirst(false)
	r.Pn(1).Ps(ps)
	log.Info("ArcPassedSearchCursorAid elastic query:%v", r.Params())
	var result *model.ArcPassedSearchReply
	if err := r.Scan(ctx, &result); err != nil {
		return nil, errors.Wrap(err, r.Params())
	}
	if result == nil {
		return nil, nil
	}
	reply := &model.ArcCursorAidSearchReply{}
	for _, val := range result.Result {
		if val == nil {
			continue
		}
		var score int64
		switch order {
		case api.SearchOrder_pubtime:
			pubtime, err := time.ParseInLocation("2006-01-02 15:04:05", val.Pubtime, time.Local)
			if err != nil {
				log.Error("日志告警 转换投稿时间错误 data:%+v,error:%+v", val, err)
				continue
			}
			score = pubtime.Unix()
		case api.SearchOrder_click:
			score = val.Click
		case api.SearchOrder_fav:
			score = val.Fav
		case api.SearchOrder_share:
			score = val.Share
		case api.SearchOrder_reply:
			score = val.Reply
		case api.SearchOrder_coin:
			score = val.Coin
		case api.SearchOrder_dm:
			score = val.Dm
		case api.SearchOrder_likes:
			score = val.Likes
		}
		reply.Result = append(reply.Result, &model.ArcsCursorAidResult{
			Aid:   val.Aid,
			Score: score,
		})
	}
	reply.Page = result.Page
	return reply, nil
}

func (d *dao) arcPassedSearchScore(ctx context.Context, mid, aid, tid int64, order api.SearchOrder, without []api.Without) (*model.ArcScoreResult, error) {
	r := d.elastic.NewRequest(businessKey("")).Index(_index)
	r.WhereEq("aid", aid)
	eq := []map[string]interface{}{
		{"mid": mid},
		{"staff_mid": mid},
	}
	//!(medl_id==59 && state=1)
	nots := []map[string]interface{}{
		{
			"archive_flow.meal_id": d.noSpace,
			"archive_flow.state":   1,
		},
	}
	for _, val := range without {
		switch val {
		case api.Without_staff:
			eq = []map[string]interface{}{
				{"mid": mid},
			}
		case api.Without_live_playback:
			r.WhereIn("up_from", d.livePlaybackUpFrom)
			r.WhereNot(elastic.NotTypeIn, "up_from")
		case api.Without_no_space:
			//!(medl_id==59 && state=1) && !(medl_id==60 && state=1)
			nots = []map[string]interface{}{
				{
					"archive_flow.meal_id": d.noSpace,
					"archive_flow.state":   1,
				},
				{
					"archive_flow.meal_id": d.upNoSpace,
					"archive_flow.state":   1,
				},
			}
		default:
		}
	}
	comboEq := &elastic.Combo{}
	comboEq.ComboEQ(eq).MinEQ(1).MinAll(1)
	comboNots := &elastic.Combo{}
	comboNots.ComboNestedNots("archive_flow", nots).MinNested(len(nots)).MinAll(1)
	r.WhereCombo(comboEq, comboNots)
	if tid != 0 {
		r.WhereEq("pid", tid)
	}
	r.WhereRange("state", 0, nil, elastic.RangeScopeLcRc)
	if len(d.notAttrs) != 0 {
		r.WhereIn("attribute", d.notAttrs)
		r.WhereNot(elastic.NotTypeIn, "attribute")
	}
	if len(d.notAttrV2s) != 0 {
		r.WhereIn("attribute_v2", d.notAttrV2s)
		r.WhereNot(elastic.NotTypeIn, "attribute_v2")
	}
	log.Info("ArcPassedSearchScore elastic query:%v", r.Params())
	var result *model.ArcPassedSearchReply
	if err := r.Scan(ctx, &result); err != nil {
		return nil, errors.Wrap(err, r.Params())
	}
	if result == nil {
		return nil, nil
	}
	if len(result.Result) > 1 {
		log.Error("日志告警 搜索获取到多个结果,elastic query:%v", r.Params())
	}
	for _, val := range result.Result {
		var score int64
		switch order {
		case api.SearchOrder_pubtime:
			pubtime, err := time.ParseInLocation("2006-01-02 15:04:05", val.Pubtime, time.Local)
			if err != nil {
				log.Error("日志告警 转换投稿时间错误 data:%+v,error:%+v", val, err)
				continue
			}
			score = pubtime.Unix()
		case api.SearchOrder_click:
			score = val.Click
		case api.SearchOrder_fav:
			score = val.Fav
		case api.SearchOrder_share:
			score = val.Share
		case api.SearchOrder_reply:
			score = val.Reply
		case api.SearchOrder_coin:
			score = val.Coin
		case api.SearchOrder_dm:
			score = val.Dm
		case api.SearchOrder_likes:
			score = val.Likes
		}
		return &model.ArcScoreResult{
			Score: score,
		}, nil
	}
	return nil, nil
}

func (d *dao) arcsPassedSearchSort(ctx context.Context, mids []int64, tid int64, ps int, order api.SearchOrder, sort api.Sort) (map[int64][]int64, error) {
	r := d.elastic.NewRequest(businessKey("")).Index(_index)
	r.WhereIn("mid", mids)
	r.WhereRange("state", 0, nil, elastic.RangeScopeLcRc)
	if tid != 0 {
		r.WhereEq("pid", tid)
	}
	if len(d.notAttrs) != 0 {
		r.WhereIn("attribute", d.notAttrs)
		r.WhereNot(elastic.NotTypeIn, "attribute")
	}
	if len(d.notAttrV2s) != 0 {
		r.WhereIn("attribute_v2", d.notAttrV2s)
		r.WhereNot(elastic.NotTypeIn, "attribute_v2")
	}
	//!(medl_id==59 && state=1)
	nots := []map[string]interface{}{
		{
			"archive_flow.meal_id": d.noSpace,
			"archive_flow.state":   1,
		},
	}
	comboNots := &elastic.Combo{}
	comboNots.ComboNestedNots("archive_flow", nots).MinNested(len(nots)).MinAll(1)
	r.WhereCombo(comboNots)
	subAid := elastic.GroupBy{}
	subAid.Field = "aid"
	subAid.Size = ps
	subAid.Mode = elastic.EnhancedModeGroupBy
	subAid.Order = []map[string]string{{"max_" + order.String(): sort.String()}}
	// 按照pubtime 排序
	subPubtime := elastic.GroupBy{}
	subPubtime.Field = order.String()
	subPubtime.Mode = elastic.EnhancedModeMax
	subAid.Subs = append(subAid.Subs, subPubtime)
	r.GroupBySub(elastic.EnhancedModeGroupBy, "mid", nil, subAid)
	log.Info("ArcsPassedSort elastic query:%v", r.Params())
	var result *model.ArcSortReply
	if err := r.Scan(ctx, &result); err != nil {
		return nil, errors.Wrap(err, r.Params())
	}
	if result == nil || result.Result == nil {
		return nil, nil
	}
	reply := map[int64][]int64{}
	for _, val := range result.Result.GroupByMid {
		if val == nil || val.GroupByAid == nil {
			continue
		}
		mid, err := strconv.ParseInt(val.Key, 10, 64)
		if err != nil {
			log.Error("日志告警 ArcsPassedSort mid,error:%+v", err)
			continue
		}
		var aids []int64
		for _, v := range val.GroupByAid.Buckets {
			aid, err := strconv.ParseInt(v.Key, 10, 64)
			if err != nil {
				log.Error("日志告警 ArcsPassedSort aid错误,error:%+v", err)
				continue
			}
			aids = append(aids, aid)
		}
		if len(aids) == 0 {
			continue
		}
		reply[mid] = aids
	}
	return reply, nil
}
