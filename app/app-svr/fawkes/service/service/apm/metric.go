package monitor

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strings"
	"time"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	yamlV2 "gopkg.in/yaml.v2"
)

const metricConfigFile = "mobile_apm_metrics.yaml"

// ApmMetricList metric表查询
func (s *Service) ApmMetricList(c context.Context, req *apm.PrometheusMetricListReq) (res *apm.PrometheusMetricRes, err error) {
	var (
		total      int
		metricList []*apm.PrometheusMetric
	)
	if total, err = s.fkDao.ApmMetricCount(c, req.Metric, req.ApmDatabaseName, req.ApmTableName, req.Operator, req.State, req.Status, req.BusID); err != nil {
		log.Errorc(c, "ApmMetricCount error %v", err)
		return
	}
	if metricList, err = s.fkDao.ApmMetricList(c, req.Metric, req.ApmDatabaseName, req.ApmTableName, req.Operator, req.Pn, req.Ps, req.State, req.Status, req.BusID, false); err != nil {
		log.Errorc(c, "ApmMetricList error %v ", err)
		return
	}
	var (
		IsModify bool
		diff     *apm.MetricDiff
	)
	if diff, err = s.ApmMetricPublishDiff(c); err != nil {
		log.Errorc(c, "ApmMetricPublishDiff error %v", err)
		return
	}
	if diff.HistoryVersion != diff.CurVersion {
		IsModify = true
	}
	res = &apm.PrometheusMetricRes{
		Items: metricList,
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    req.Pn,
			Ps:    req.Ps,
		},
		IsModify: IsModify,
	}
	return
}

// ApmMetricSet metric表修改
func (s *Service) ApmMetricSet(c context.Context, req *apm.PrometheusMetric) (err error) {
	var origin *apm.PrometheusMetric
	if origin, err = s.fkDao.ApmMetricByMetric(c, req.Metric); err != nil {
		log.Error("ApmMetricSet s.fkDao.ApmMetricByID error(%v)", err)
		return
	}
	if origin == nil {
		if _, err = s.fkDao.ApmMetricAdd(c, req.Metric, req.MetricType, req.ExecSQL, req.LabeledKeys, req.ValueKey, req.TimestampKey, req.Description, req.ApmDatabaseName, req.ApmTableName, req.Operator, req.URL, req.TimeFilter, req.TimeOffset, req.BusID, req.Status); err != nil {
			log.Error("ApmMetricSet s.fkDao.ApmMetricAdd error(%v)", err)
			return
		}
	} else {
		curState := apmMetricStateChange(origin, req)
		if curState == 0 {
			return
		}
		if _, err = s.fkDao.ApmMetricUpdate(c, req.Metric, req.MetricType, req.ExecSQL, req.LabeledKeys, req.ValueKey, req.TimestampKey,
			req.Description, req.ApmDatabaseName, req.ApmTableName, req.Operator, req.URL, req.TimeFilter, req.TimeOffset, req.BusID, curState, req.Status); err != nil {
			log.Error("ApmMetricSet s.fkDao.ApmMetricUpdate error(%v)", err)
		}
	}
	return
}

// apmMetricStateChange metrics状态判断
func apmMetricStateChange(origin, current *apm.PrometheusMetric) (curState int8) {
	if apmMetricPubFieldEqual(origin, current) && origin.ApmDatabaseName == current.ApmDatabaseName &&
		origin.ApmTableName == current.ApmTableName && origin.TimeFilter == current.TimeFilter && origin.TimeOffset == current.TimeOffset && origin.URL == current.URL && origin.BusID == current.BusID && origin.Status == current.Status {
		return
	}
	if apmMetricPubFieldEqual(origin, current) && (origin.ApmDatabaseName != current.ApmDatabaseName ||
		origin.ApmTableName != current.ApmTableName || origin.TimeFilter != current.TimeFilter || origin.TimeOffset != current.TimeOffset || origin.URL != current.URL || origin.BusID != current.BusID) {
		curState = origin.State
		return
	}
	if origin.State == apm.PrometheusAdd {
		curState = apm.PrometheusAdd
	} else if origin.State == apm.PrometheusDel {
		curState = apm.PrometheusDel
	} else {
		curState = apm.PrometheusModify
	}
	return
}

// apmMetricPubFieldEqual metrics发布字段是否相等
func apmMetricPubFieldEqual(origin, current *apm.PrometheusMetric) (flag bool) {
	originSql := strings.ReplaceAll(origin.ExecSQL, " ", "")
	curSql := strings.ReplaceAll(current.ExecSQL, " ", "")
	return origin.Metric == current.Metric && origin.MetricType == current.MetricType && originSql == curSql && origin.LabeledKeys == current.LabeledKeys &&
		origin.ValueKey == current.ValueKey && origin.TimestampKey == current.TimestampKey && origin.Description == current.Description
}

// ApmMetricDel metric表删除
func (s *Service) ApmMetricDel(c context.Context, req *apm.PrometheusMetric, isUndoDel int64) (err error) {
	var (
		origin *apm.PrometheusMetric
	)
	if origin, err = s.fkDao.ApmMetricByMetric(c, req.Metric); err != nil {
		log.Error("ApmMetricDel s.fkDao.ApmMetricByID error(%v)", err)
		return
	}
	if origin.State == apm.PrometheusAdd {
		if _, err = s.fkDao.ApmMetricDel(c, req.Metric); err != nil {
			log.Error("ApmMetricDel s.fkDao.ApmMetricDel2 error(%v)", err)
		}
	} else {
		var curState int8
		if isUndoDel == 0 {
			curState = apm.PrometheusDel
		} else {
			curState = apm.PrometheusPublish
		}
		if _, err = s.fkDao.ApmMetricDelByUpdate(c, curState, req.Metric); err != nil {
			log.Error("ApmMetricUpdate s.fkDao.ApmMetricDel error(%v)", err)
		}
	}
	return
}

// ApmMetricPublish metric发布
func (s *Service) ApmMetricPublish(c context.Context, req *apm.PrometheusMetricPublishReq) (add int64, err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmMetricPublishDel(tx); err != nil {
			log.Errorc(c, "ApmMetricPublishDel error %v", err)
			return err
		}
		var ymlByte []byte
		ymlByte, err = s.apmMetricYmlGenerator(c)
		if err != nil {
			log.Errorc(c, "apmMetricYmlGenerator error %v", err)
			return err
		}
		fileName := fmt.Sprintf("mobile_apm_metrics-%s.yaml", time.Now().Format("20060102150405"))
		localPath := path.Join(s.c.Prometheus.LocalPath.LocalDir, fileName)
		filePath := path.Join(s.c.LocalPath.LocalDir, localPath)
		selectorFilePath := path.Join(path.Dir(filePath), metricConfigFile)
		//nolint:gosec
		if err = ioutil.WriteFile(filePath, ymlByte, 0644); err != nil {
			log.Errorc(c, "ioutil.WriteFile error %v", err)
			return err
		}
		//nolint:gosec
		if err = ioutil.WriteFile(selectorFilePath, ymlByte, 0644); err != nil {
			log.Errorc(c, "ioutil.WriteFile error %v", err)
			return err
		}
		if add, err = s.fkDao.TxApmMetricPublish(tx, hex.EncodeToString(ymlByte[:]), localPath, req.Operator, req.Description); err != nil {
			log.Errorc(c, "TxApmMetricPublish error %v", err)
			return err
		}
		if err = s.fkDao.ApmMetricPublishStateUpdate(tx, apm.PrometheusPublish); err != nil {
			log.Errorc(c, "ApmMetricPublishStateUpdate error %v", err)
			return err
		}
		var activePublish *apm.PrometheusMetricPublish
		if activePublish, err = s.fkDao.ApmMetricPublishActive(c); err != nil {
			log.Errorc(c, "ApmMetricPublishActive error %v", err)
			return err
		}
		if activePublish == nil {
			log.Warnc(c, "activePublish is nil")
			return err
		}
		if err = s.fkDao.TxApmMetricPublishActiveVerUpdate(tx, activePublish.ID, apm.UnActiveVersion); err != nil {
			log.Errorc(c, "TxApmMetricPublishActiveVerUpdate error %v", err)
			return err
		}
		return err
	})
	return
}

// ApmMetricPublishList publish历史查询
func (s *Service) ApmMetricPublishList(c context.Context, req *apm.PrometheusMetricPublishListReq) (res *apm.PrometheusMetricPublishListRes, err error) {
	var (
		total       int
		publishList []*apm.PrometheusMetricPublish
	)
	if total, err = s.fkDao.ApmMetricPublishCount(c, req.MD5, req.LocalPath, req.Description, req.Operator); err != nil {
		log.Error("ApmMetricPublishList s.fkDao.ApmMetricPublishCount error(%v)", err)
		return
	}
	if publishList, err = s.fkDao.ApmMetricPublishList(c, req.MD5, req.LocalPath, req.Description, req.Operator, req.Pn, req.Ps); err != nil {
		log.Error("ApmMetricPublishList s.fkDao.ApmMetricPublishList error(%v)", err)
		return
	}
	for _, v := range publishList {
		v.LocalURL = s.c.LocalPath.LocalDomain + v.LocalPath
	}
	res = &apm.PrometheusMetricPublishListRes{
		Items: publishList,
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    req.Pn,
			Ps:    req.Ps,
		},
	}
	return
}

// ApmMetricPublishDiff publish diff
func (s *Service) ApmMetricPublishDiff(c context.Context) (res *apm.MetricDiff, err error) {
	localPathHistory, err := s.fkDao.ApmMetricPublishDiff(c)
	if err != nil {
		log.Error("ApmMetricPublishDiff s.fkDao.ApmMetricPublishDiff error(%v)", err)
		return
	}
	filePathHistory := s.c.LocalPath.LocalDir + localPathHistory
	curBuf, err := s.apmMetricYmlGenerator(c)
	if err != nil {
		log.Error("ApmMetricPublishDiff s.apmMetricYmlGenerator error(%v)", err)
		return
	}
	historyBuf, err := ioutil.ReadFile(filePathHistory)
	if err != nil {
		log.Error("ApmMetricPublishDiff ioutil.ReadFile error(%v)", err)
		return
	}
	res = &apm.MetricDiff{
		CurVersion:     string(curBuf),
		HistoryVersion: string(historyBuf),
	}
	return
}

func (s *Service) apmMetricYmlGenerator(c context.Context) (ymlRes []byte, err error) {
	var (
		metricList   []*apm.PrometheusMetric
		selectorList []*apm.YmlMetricSelector
	)
	if metricList, err = s.fkDao.ApmMetricList(c, "", "", "", "", -1, -1, -1, apm.MetricStatusOn, 0, true); err != nil {
		log.Error("ApmMetricPublish s.fkDao.ApmMetricList error(%v)", err)
		return
	}
	for _, metric := range metricList {
		selector := &apm.YmlMetricSelector{
			Metric:       metric.Metric,
			ApmType:      metric.MetricType,
			Sql:          strings.ReplaceAll(strings.Trim(metric.ExecSQL, " "), s.c.PrometheusTemplate.Key, s.c.PrometheusTemplate.Value),
			LabeledKeys:  strings.Split(strings.ReplaceAll(metric.LabeledKeys, " ", ""), ","),
			ValueKey:     metric.ValueKey,
			TimestampKey: metric.TimestampKey,
			Help:         strings.Trim(metric.Description, " "),
		}
		selectorList = append(selectorList, selector)
	}
	sort.Slice(selectorList, func(i, j int) bool {
		return selectorList[i].Metric < selectorList[j].Metric
	})
	metricDatabase := &apm.YmlMetricDatabase{
		Name:      s.c.Prometheus.Database.Name,
		Host:      s.c.Prometheus.Database.Host,
		Port:      s.c.Prometheus.Database.Port,
		User:      s.c.Prometheus.Database.User,
		Password:  s.c.Prometheus.Database.Password,
		Selectors: selectorList,
	}
	metricConfig := &apm.YmlMetricConfig{
		Databases: []*apm.YmlMetricDatabase{metricDatabase},
	}
	ymlRes, err = yamlV2.Marshal(metricConfig)
	if err != nil {
		log.Error("apmMetricYmlGenerator yaml.Marshal error(%v)", err)
	}
	return
}

func (s *Service) ApmMetricPublishRollback(c context.Context, req *apm.PrometheusMetricPublishRollbackReq) (err error) {
	var publish *apm.PrometheusMetricPublish
	if publish, err = s.fkDao.ApmMetricPublishById(c, req.Id); err != nil {
		log.Errorc(c, "ApmMetricPublishById error %v", err)
		return
	}
	if publish.IsActiveVersion == apm.ActiveVersion {
		return
	}
	dir := path.Dir(publish.LocalPath)
	rollbackFilePath := path.Join(s.c.LocalPath.LocalDir, publish.LocalPath)
	rollbackFile, err := ioutil.ReadFile(rollbackFilePath)
	if err != nil {
		log.Errorc(c, "ioutil.ReadFile error %v", err)
		return err
	}
	targetPath := path.Join(s.c.LocalPath.LocalDir, dir, metricConfigFile)
	//nolint:gosec
	if err = ioutil.WriteFile(targetPath, rollbackFile, 0644); err != nil {
		log.Error("ApmMetricPublish ioutil.WriteFile error(%v)", err)
		return
	}
	var activePublish *apm.PrometheusMetricPublish
	if activePublish, err = s.fkDao.ApmMetricPublishActive(c); err != nil {
		log.Errorc(c, "ApmMetricPublishActive error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmMetricPublishActiveVerUpdate(tx, activePublish.ID, apm.UnActiveVersion); err != nil {
			log.Errorc(c, "TxApmMetricPublishActiveVerUpdate error %v", err)
			return err
		}
		if err = s.fkDao.TxApmMetricPublishActiveVerUpdate(tx, publish.ID, apm.ActiveVersion); err != nil {
			log.Errorc(c, "TxApmMetricPublishActiveVerUpdate error %v", err)
			return err
		}
		return err
	})
	return
}
