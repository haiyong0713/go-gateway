package abtest

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/xstr"
	"hash/crc32"
	"strconv"
	"strings"
	"time"

	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	expdao "go-gateway/app/app-svr/app-resource/interface/dao/abtest"
	abTestMdl "go-gateway/app/app-svr/app-resource/interface/model/abtest"
	"go-gateway/app/app-svr/app-resource/interface/model/experiment"
	"go-gateway/app/app-svr/resource/service/model"

	farm "go-farm"

	"github.com/robfig/cron"
)

var (
	_emptyExperiment   = []*experiment.Experiment{}
	_defaultExperiment = map[int8][]*experiment.Experiment{

		model.PlatAndroid: {
			{
				ID:           10,
				Name:         "默认值",
				Strategy:     "default_value",
				Desc:         "默认值为不匹配处理",
				TrafficGroup: "0",
			},
		},
	}
)

type Service struct {
	dao *expdao.Dao
	// tick
	tick time.Duration
	epm  map[int8][]*experiment.Experiment
	c    *conf.Config
	// cron
	cron *cron.Cron
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: expdao.New(c),
		// tick
		tick: time.Duration(c.Tick),
		epm:  map[int8][]*experiment.Experiment{},
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.loadAbTest()
	if err := s.cron.AddFunc(s.c.Cron.LoadAbTest, s.loadAbTest); err != nil {
		panic(err)
	}
}

// TemporaryABTests 临时的各种abtest垃圾需求
func (s *Service) TemporaryABTests(c context.Context, buvid string, mid int64) (tests *experiment.ABTestV2) {
	id := farm.Hash32([]byte(buvid))
	n := int(id % 100)
	autoPlay := 1
	if n > s.c.ABTest.Range {
		autoPlay = 2
	}
	tests = &experiment.ABTestV2{
		AutoPlay:      autoPlay,
		UnloginAbTest: 1,
		LoginWindow:   1,
	}
	return
}

func (s *Service) Experiment(c context.Context, plat int8, build int) (eps []*experiment.Experiment) {
	if es, ok := s.epm[plat]; ok {
	LOOP:
		for _, ep := range es {
			for _, l := range ep.Limit {
				if model.InvalidBuild(build, l.Build, l.Condition) {
					continue LOOP
				}
			}
			eps = append(eps, ep)
		}
	}
	if eps == nil {
		if es, ok := _defaultExperiment[plat]; ok {
			eps = es
		} else {
			eps = _emptyExperiment
		}
	}
	return
}

func (s *Service) loadAbTest() {
	log.Info("cronLog start loadAbTest")
	c := context.TODO()
	lm, err := s.dao.ExperimentLimit(c)
	if err != nil {
		log.Error("s.dao.ExperimentLimit error(%v)", err)
		return
	}
	ids := make([]int64, 0, len(lm))
	for id := range lm {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return
	}
	eps, err := s.dao.ExperimentByIDs(c, ids)
	if err != nil {
		log.Error("s.dao.ExperimentByIDs(%v) error(%v)", ids, err)
		return
	}
	epm := make(map[int8][]*experiment.Experiment, len(eps))
	for _, ep := range eps {
		if l, ok := lm[ep.ID]; ok {
			ep.Limit = l
		}
		epm[ep.Plat] = append(epm[ep.Plat], ep)
	}
	s.epm = epm
}

// AbServer is
func (s *Service) AbServer(c context.Context, buvid, device, mobiAPP, filteredStr, model, brand, osver string, build int, mid int64) (interface{}, error) {
	res, err := s.dao.AbServer(c, buvid, device, mobiAPP, filteredStr, model, brand, osver, build, mid)
	if err != nil {
		return nil, err
	}
	if s.c.TF == nil || s.c.TF.Rule == "" {
		return res, nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(res, &result); err != nil {
		log.Error("%+v", err)
		return res, nil
	}
	vars, ok := result["vars"].([]interface{})
	if !ok {
		return result, nil
	}
	for _, v := range vars {
		value, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := value["name"]; ok {
			if name == "tf_rules" {
				value["value"] = s.c.TF.Rule
			}
		}
	}
	return result, nil
}

func (s *Service) AbTestList(c context.Context, param *abTestMdl.AbTestListParam, mid int64, buvid string) (*abTestMdl.AbTestListReply, error) {
	var res = make(map[string]*abTestMdl.AbTestItem)
	for _, expName := range strings.Split(param.Keys, ",") {
		expConfig, ok := s.c.Experiment.Config[expName]
		if !ok {
			continue
		}
		if expConfig.ExpName == "" {
			continue
		}
		if !expConfig.Switch {
			res[expConfig.ExpName] = &abTestMdl.AbTestItem{Result: abTestMdl.GrpupDefault}
			continue
		}
		if expConfig.ExpType == "" {
			res[expConfig.ExpName] = &abTestMdl.AbTestItem{Result: abTestMdl.GrpupDefault}
			continue
		}
		var bucket int
		switch expConfig.ExpType {
		case "mid":
			bucket = int(crc32.ChecksumIEEE([]byte(strconv.FormatInt(mid, 10)))) % expConfig.Bucket
		case "buvid":
			bucket = int(crc32.ChecksumIEEE([]byte(buvid))) % expConfig.Bucket
		default:
			res[expConfig.ExpName] = &abTestMdl.AbTestItem{Result: abTestMdl.GrpupDefault}
			continue
		}
		// 分组判断逻辑
		var hitGroup string
		for _, groupConfig := range expConfig.Groups {
			// 白名单逻辑
			switch expConfig.ExpType {
			case "mid":
				var (
					wlMids []int64
					errTmp error
				)
				if wlMids, errTmp = xstr.SplitInts(groupConfig.WhiteList); errTmp != nil {
					log.Error("AbTestList expName(%v) groupName(%v) whiteList error(%v)", expConfig.ExpName, groupConfig.GroupName, errTmp)
					continue
				}
				for _, wlMid := range wlMids {
					if wlMid == mid {
						hitGroup = groupConfig.GroupName
						break
					}
				}
			case "buvid":
				for _, wlBuvid := range strings.Split(groupConfig.WhiteList, ",") {
					if wlBuvid == buvid {
						hitGroup = groupConfig.GroupName
						break
					}
				}
			}
			// 常规分组
			if bucket >= groupConfig.Start && bucket <= groupConfig.End {
				hitGroup = groupConfig.GroupName
				break
			}
		}
		res[expConfig.ExpName] = &abTestMdl.AbTestItem{Result: abTestMdl.GrpupDefault}
		if hitGroup != "" {
			res[expConfig.ExpName].Result = hitGroup
		}
	}
	return &abTestMdl.AbTestListReply{List: res}, nil
}

var TinyABFlag map[string]*ab.StringFlag

func (s *Service) TinyAbtest(ctx context.Context) *abTestMdl.TinyAbReply {
	abRes := tinyABTestRun(ctx)
	reply := &abTestMdl.TinyAbReply{
		ABResult: abRes,
	}
	//目前两个实验还是网关下发业务配置，后面接入走在线参数，该接口只提供极小包实验命中情况
	if abtest, ok := abRes["tiny_upgrade_pop_v2"]; ok {
		reply.PopupStyle = abtest
		reply.UpgradeInform = &abTestMdl.UpgradeInform{
			Text:   s.c.Custom.TinyUpgradeInform,
			Title:  s.c.Custom.TinyUpgradeInformTitle,
			Timing: s.c.Custom.TinyUpgradeInformTiming,
			ABTest: abtest,
		}
	}
	return reply
}

func tinyABTestRun(ctx context.Context) map[string]*abTestMdl.ABTest {
	out := make(map[string]*abTestMdl.ABTest)
	for flagValue, flag := range TinyABFlag {
		t, ok := ab.FromContext(ctx)
		if !ok {
			log.Error("failed to get abtest context")
			continue
		}
		if !strings.HasPrefix(flag.Value(t), "1") { //没有命中实验
			continue
		}
		abtest := &abTestMdl.ABTest{
			Exp: int64(ab.ExpHit),
		}
		for _, v := range t.Snapshot() {
			if v.Type == ab.ExpHit {
				abtest.GroupID = v.Value
				break
			}
		}
		out[flagValue] = abtest
	}
	return out
}

func Init(cfg *conf.Config) {
	TinyABFlag = make(map[string]*ab.StringFlag)
	for _, v := range cfg.ABTestFlags {
		TinyABFlag[v] = ab.String(v, fmt.Sprintf("极小包实验-%s", v), "miss")
	}
}
