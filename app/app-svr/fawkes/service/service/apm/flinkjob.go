package monitor

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/app-svr/fawkes/service/model"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const _intMax = int(^uint(0) >> 1)

// ApmFlinkJobList Flink任务列表
func (s *Service) ApmFlinkJobList(c context.Context, req *apm.FlinkJobReq) (res *apm.FlinkJobRes, err error) {
	var (
		total        int
		flinkJobList []*apm.FlinkJobDB
	)
	if total, err = s.fkDao.ApmFlinkJobCount(c, req.LogID, req.Name, req.Description, req.Owner, req.Operator,
		req.State, req.StartTime, req.EndTime); err != nil {
		log.Error("ApmFlinkJobList s.fkDao.ApmFlinkJobCount error(%v)", err)
		return
	}
	if total == 0 {
		return
	}
	if flinkJobList, err = s.fkDao.ApmFlinkJobList(c, req.LogID, req.Name, req.Description, req.Owner, req.Operator,
		req.State, req.StartTime, req.EndTime, req.Pn, req.Ps); err != nil {
		log.Error("ApmFlinkJobList s.fkDao.ApmFlinkJobList error(%v)", err)
		return
	}

	for _, data := range flinkJobList {
		data.ModifyCount, err = s.fkDao.ApmFlinkJobPublishModifyCount(c, data.ID)
		if err != nil {
			log.Error("ApmFlinkJobList s.fkDao.ApmFlinkJobPublishModifyCount error(%v)", err)
			return
		}
	}
	res = &apm.FlinkJobRes{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    req.Pn,
			Ps:    req.Ps,
		},
		Items: flinkJobList,
	}
	return
}

// ApmFlinkJobAdd Flink任务添加
func (s *Service) ApmFlinkJobAdd(c context.Context, req *apm.FlinkJobReq) (resp interface{}, err error) {
	r, err := s.fkDao.ApmFlinkJobAdd(c, req.LogID, req.Name, req.Description, req.Owner, req.Operator, req.State)
	if err != nil {
		log.Error("ApmFlinkJobAdd s.fkDao.ApmFlinkJobAdd error(%v)", err)
		return
	}
	return struct {
		ID int64 `json:"id"`
	}{r}, err
}

// ApmFlinkJobUpdate Flink任务更新
func (s *Service) ApmFlinkJobUpdate(c context.Context, req *apm.FlinkJobReq) (resp interface{}, err error) {
	eff, err := s.fkDao.ApmFlinkJobUpdate(c, req.LogID, req.Name, req.Description, req.Owner, req.Operator, req.State, req.ID)
	if err != nil {
		log.Error("ApmFlinkJobUpdate s.fkDao.ApmFlinkJobUpdate error(%v)", err)
		return
	}
	return struct {
		EffectRows int64 `json:"effect_rows"`
	}{eff}, err
}

// ApmFlinkJobDel Flink任务通过id删除
func (s *Service) ApmFlinkJobDel(c context.Context, req *apm.FlinkJobReq) (resp interface{}, err error) {
	eff, err := s.fkDao.ApmFlinkJobDel(c, req.ID)
	if err != nil {
		log.Error("ApmFlinkJobDel s.fkDao.ApmFlinkJobDel error(%v)", err)
		return
	}
	return struct {
		EffectRows int64 `json:"effect_rows"`
	}{eff}, err
}

// ApmFlinkJobRelationList Flink任务和Events关联列表
func (s *Service) ApmFlinkJobRelationList(c context.Context, req *apm.EventFlinkRelReq) (resp []*apm.Event, err error) {
	resp, err = s.fkDao.ApmFlinkJobRelationList(c, req.JobID)
	if err != nil {
		log.Error("ApmEventFlinkRelList s.fkDao.s.fkDao.ApmEventFlinkRelList error(%v)", err)
		return
	}
	return
}

// ApmFlinkJobRelationAdd Flink任务和Events关联关系添加
func (s *Service) ApmFlinkJobRelationAdd(c context.Context, req *apm.EventFlinkRelReq) (err error) {
	if _, err = s.fkDao.ApmFlinkJobRelationAdd(c, req.EventID, req.JobID, req.Operator, apm.FlinkJobAdd); err != nil {
		log.Error("ApmEventJobRelAdd s.fkDao.TxApmFlinkJobRelationAdd error(%v)", err)
	}
	return
}

// ApmFlinkJobRelationDel Flink任务和Events关联关系删除
func (s *Service) ApmFlinkJobRelationDel(c context.Context, req *apm.EventFlinkRelReq) (err error) {
	var origin *apm.EventFlinkRelDB
	if origin, err = s.fkDao.ApmFlinkJobRelation(c, req.JobID, req.EventID); err != nil {
		log.Error("ApmFlinkJobRelationDel s.fkDao.ApmFlinkJobRelation error(%v)", err)
		return
	}
	if origin.State == apm.FlinkJobAdd {
		if _, err = s.fkDao.ApmFlinkJobRelationDel(c, req.JobID, req.EventID); err != nil {
			log.Error("TxApmFlinkJobRelationDel s.fkDao.TxApmFlinkJobRelationDel error(%v)", err)
			return
		}
	} else {
		if _, err = s.fkDao.ApmFlinkJobRelationDelByUpdate(c, req.JobID, req.EventID, apm.FlinkJobDel); err != nil {
			log.Error("TxApmFlinkJobRelationDelByUpdate s.fkDao.TxApmFlinkJobRelationDelByUpdate error(%v)", err)
			return
		}
	}
	return
}

// ApmFlinkJobPublish Flink任务和Events关联关系发布
func (s *Service) ApmFlinkJobPublish(c context.Context, req *apm.EventFlinkRelReq) (add int64, err error) {
	if _, err = s.fkDao.ApmFlinkJobPublishDel(c, req.JobID); err != nil {
		log.Error("ApmFlinkJobPublish s.fkDao.ApmFlinkJobPublishDel error(%v)", err)
		return
	}
	jsonByte, err := s.apmFlinkJsonGenerator(c, req.JobID)
	if err != nil {
		log.Error("ApmFlinkJobPublish s.apmFlinkJsonGenerator error(%v)", err)
		return
	}
	flinkJob, err := s.fkDao.ApmFlinkJobById(c, req.JobID)
	if err != nil {
		log.Error("ApmEventFlinkRelPublish s.fkDao.ApmFlinkJobById error(%v)", err)
		return
	}
	fileName := fmt.Sprintf("flink_job_%s-%s.json", flinkJob.Name, time.Now().Format("20060102150405"))
	localPath := fmt.Sprintf("%s/%v/%v", s.c.FlinkJob.LocalPath.LocalDir, "flink_job_"+flinkJob.Name, fileName)
	filePath := s.c.LocalPath.LocalDir + localPath
	selectorFilePath := strings.Replace(filePath, fileName, fmt.Sprintf("flink_job_%s.json", flinkJob.Name), 1)
	//	检测文件夹是否存在
	detectSubDir := fmt.Sprintf("%s/%v", s.c.FlinkJob.LocalPath.LocalDir, "flink_job_"+flinkJob.Name)
	detectDir := s.c.LocalPath.LocalDir + detectSubDir
	exist, err := pathExists(detectDir)
	if err != nil {
		log.Error("ApmEventFlinkRelPublish pathExists error(%v)", err)
		return
	}
	if !exist {
		err = os.Mkdir(detectDir, os.ModePerm)
		if err != nil {
			log.Error("ApmEventFlinkRelPublish os.Mkdir error(%v)", err)
			return
		}
	}
	//nolint:gosec
	if err = ioutil.WriteFile(filePath, jsonByte, 0644); err != nil {
		log.Error("ApmEventFlinkRelPublish ioutil.WriteFile error(%v)", err)
		return
	}
	//nolint:gosec
	if err = ioutil.WriteFile(selectorFilePath, jsonByte, 0644); err != nil {
		log.Error("ApmEventFlinkRelPublish ioutil.WriteFile error(%v)", err)
		return
	}
	if add, err = s.fkDao.ApmFlinkJobPublish(c, req.JobID, hex.EncodeToString(jsonByte[:]), localPath, req.Description, req.Operator); err != nil {
		log.Error("ApmEventFlinkRelPublish s.fkDao.ApmEventFlinkRelPublish error(%v)", err)
		return
	}
	if _, err = s.fkDao.ApmFlinkJobPublishStateUpdate(c, apm.FlinkJobPublish, req.JobID); err != nil {
		log.Error("ApmFlinkJobPublish s.fkDao.ApmFlinkJobPublishStateUpdate error(%v)", err)
		return
	}
	return
}

// ApmFlinkJobPublishList Flink任务和Events关联关系发布列表
func (s *Service) ApmFlinkJobPublishList(c context.Context, req *apm.EventFlinkRelPublishListReq) (res *apm.EventFlinkRelPublishListRes, err error) {
	total, err := s.fkDao.ApmFlinkJobPublishCount(c, req.FlinkJobID)
	if err != nil {
		log.Error("ApmEventFlinkRelPublishList s.fkDao.ApmEventFlinkRelPublishCount error(%v)", err)
		return
	}
	if total == 0 {
		return
	}
	publishList, err := s.fkDao.ApmFlinkJobPublishList(c, req.FlinkJobID, req.Pn, req.Ps)
	if err != nil {
		log.Error("ApmEventFlinkRelPublish s.fkDao.ApmEventFlinkRelPublishList error(%v)", err)
		return
	}
	for _, data := range publishList {
		data.LocalUrl = s.c.LocalPath.LocalDomain + data.LocalPath
	}
	res = &apm.EventFlinkRelPublishListRes{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    req.Pn,
			Ps:    req.Ps,
		},
		Items: publishList,
	}
	return
}

func (s *Service) ApmFlinkJobPublishDiff(c context.Context, jobID int64) (resp *apm.EventFlinkPublishDiff, err error) {
	localPathHistory, err := s.fkDao.ApmFlinkJobLastPath(c, jobID)
	if err != nil {
		log.Error("ApmFlinkJobPublishDiff s.fkDao.ApmFlinkJobPublishDiff error(%v)", err)
		return
	}
	filePathHistory := s.c.LocalPath.LocalDir + localPathHistory
	fileCur, err := s.apmFlinkJsonGenerator(c, jobID)
	if err != nil {
		log.Error("ApmFlinkJobPublishDiff s.apmFlinkJsonGenerator error(%v)", err)
		return
	}
	if len(localPathHistory) == 0 {
		resp = &apm.EventFlinkPublishDiff{CurVersion: string(fileCur), HistoryVersion: ""}
		return
	}
	historyBuf, err := ioutil.ReadFile(filePathHistory)
	if err != nil {
		log.Error("ApmFlinkJobPublishDiff ioutil.ReadFile error(%v)", err)
		return
	}
	resp = &apm.EventFlinkPublishDiff{CurVersion: string(fileCur), HistoryVersion: string(historyBuf)}
	return
}

func (s *Service) apmFlinkJsonGenerator(c context.Context, jobID int64) (jsonRes []byte, err error) {
	var (
		resMap          = make(map[string][]string)
		relEventJsonRes = make(map[string]*apm.EventSubRes)
		res             *apm.JsonEvent
		eventGroup      = make(map[string][]*apm.Event)
		eventsDuplicate []*apm.EventDuplicateRes
		wideTables      = make([]string, 0)
	)
	events, err := s.fkDao.ApmFlinkJobRelationList(c, jobID)
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	for _, event := range events {
		if event.Databases == "" || event.TableName == "" {
			return
		}
		var groupName string
		if event.IsWideTable == 1 {
			wideTable := fmt.Sprintf("%v.%v", event.Databases, event.TableName)
			wideTables = append(wideTables, wideTable)
			groupName = event.Name
		} else {
			groupName = fmt.Sprintf("%v.%v", event.Databases, event.TableName)
		}
		eventGroup[groupName] = append(eventGroup[groupName], event)
	}
	for key, evs := range eventGroup {
		var (
			ids         []int64
			names       []string
			sampleRates []int
			mapName     string
		)
		for _, ev := range evs {
			ids = append(ids, ev.ID)
			names = append(names, ev.Name)
			sampleRates = append(sampleRates, ev.SampleRate)
		}
		eventDuplicate := &apm.EventDuplicateRes{
			EventIDS:    ids,
			DBName:      evs[0].Databases,
			TableName:   evs[0].TableName,
			Names:       names,
			SampleRates: sampleRates,
			EventCount:  len(ids),
		}
		// 得到映射后的event名字
		if len(evs) == 1 {
			mapName = evs[0].Name
		} else {
			mapName = key
		}
		eventDuplicate.MapName = mapName
		eventsDuplicate = append(eventsDuplicate, eventDuplicate)
	}
	for _, eventDuplicate := range eventsDuplicate {
		mapName := eventDuplicate.MapName
		if eventDuplicate.EventCount > 1 {
			// 在同一map中
			//	存在映射
			resMap[mapName] = append(resMap[mapName], eventDuplicate.Names...)
			sort.Slice(resMap[mapName], func(i, j int) bool {
				return resMap[mapName][i] < resMap[mapName][j]
			})
		}
		// 扩展字段
		var relEventSub *apm.EventSubRes
		if relEventSub, err = s.generateEventSubList(c, eventDuplicate); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		relEventJsonRes[mapName] = relEventSub
	}
	deWideTables := removeDuplication(wideTables)
	res = &apm.JsonEvent{
		Mapping:   resMap,
		WideTable: deWideTables,
		Event:     relEventJsonRes,
	}
	jsonRes, err = json.MarshalIndent(res, "", " ")
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	return
}

// pathExists 判断路径是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// removeEventFieldDuplicate 扩展字段去重
func removeEventFieldDuplicate(fieldList []*apm.EventFieldSub) (res []*apm.EventFieldSub) {
	resultMap := map[string]bool{}
	for _, v := range fieldList {
		data, _ := json.Marshal(v)
		resultMap[string(data)] = true
	}
	for k := range resultMap {
		var t *apm.EventFieldSub
		err := json.Unmarshal([]byte(k), &t)
		if err != nil {
			log.Error("removeEventFieldDuplicate json.Unmarshal error(%v)", err)
			return
		}
		res = append(res, t)
	}
	return
}

// generateEventSubList 生成事件子集
func (s *Service) generateEventSubList(c context.Context, eventDuplicate *apm.EventDuplicateRes) (res *apm.EventSubRes, err error) {
	res = &apm.EventSubRes{}
	var sampleRateMax = _intMax
	for _, sampleRate := range eventDuplicate.SampleRates {
		if sampleRate < sampleRateMax {
			sampleRateMax = sampleRate
		}
	}
	if sampleRateMax > 0 && sampleRateMax < 10000 {
		res.SampleRate = sampleRateMax
	}
	var eventFieldSubList []*apm.EventFieldSub
	for _, id := range eventDuplicate.EventIDS {
		var relEventFieldSubList []*apm.EventFieldSub
		if relEventFieldSubList, err = s.generateEventFieldSubList(c, id); err != nil {
			log.Error("s.generateEventField error(%v)", err)
			return
		}
		eventFieldSubList = append(eventFieldSubList, relEventFieldSubList...)
	}
	fieldNotDuplicateList := removeEventFieldDuplicate(eventFieldSubList)
	sort.Slice(fieldNotDuplicateList, func(i, j int) bool {
		return fieldNotDuplicateList[i].PropertyName < fieldNotDuplicateList[j].PropertyName
	})
	res.DBName = eventDuplicate.DBName
	res.TableName = eventDuplicate.TableName
	if len(fieldNotDuplicateList) == 0 {
		res.Properties = make([]*apm.EventFieldSub, 0)
	} else {
		res.Properties = fieldNotDuplicateList
	}
	return
}

// generateEventFieldSubList 生成扩展字段子集
func (s *Service) generateEventFieldSubList(c context.Context, id int64) (res []*apm.EventFieldSub, err error) {
	var (
		fv    int64
		files []*apm.EventFieldFile
	)
	// 从file归档表中读最新的字段集合
	if fv, err = s.fkDao.ApmEventFieldFileLastFV(c, id); err != nil {
		log.Errorc(c, "ApmEventFieldLastFV error %v", err)
		return
	}
	if files, err = s.fkDao.ApmEventFieldFileList(c, id, fv); err != nil {
		log.Errorc(c, "ApmEventFieldFileList error %v", err)
		return
	}
	for _, file := range files {
		//	跳过不进入clickhouse的字段
		if file.IsClickhouse == 0 {
			continue
		}
		// 跳过删除的字段
		if file.FieldState == apm.EventFieldStateDelete {
			continue
		}
		fieldType := file.FieldType
		var relEventFieldSub *apm.EventFieldSub
		if fieldType == 0 {
			relEventFieldSub = &apm.EventFieldSub{
				DefaultValue: file.DefaultValue,
				PropertyName: file.FieldKey,
			}
		} else {
			var defaultValue int64
			if defaultValue, err = strconv.ParseInt(file.DefaultValue, 10, 64); err != nil {
				log.Error("generateEventFieldSubList strconv.ParseInt error(%v)", err)
				return
			}
			relEventFieldSub = &apm.EventFieldSub{
				DefaultValue: defaultValue,
				PropertyName: file.FieldKey,
			}
		}
		res = append(res, relEventFieldSub)
	}
	return
}

func removeDuplication(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}
