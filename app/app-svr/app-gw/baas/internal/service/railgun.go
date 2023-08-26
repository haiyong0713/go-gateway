package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-gw/baas/api"
)

func (s *CommonService) initRailGun() {
	s.initExportRailGun()
	s.initImportsRailGun()
	s.initModelRuleRailGun()
	s.initFieldsRailGun()
}

func (s *CommonService) loadExportList() error {
	ctx := context.Background()
	list, err := s.dao.ExportList(ctx)
	if err != nil {
		return err
	}
	exportMap := make(map[string]*api.ExportItem, len(list))
	for _, item := range list {
		exportMap[item.ExportApi] = api.ConstructExportItem(item)
	}
	s.exports = exportMap
	return nil
}

func (s *CommonService) initExportRailGun() {
	if err := s.loadExportList(); err != nil {
		panic(fmt.Sprintf("Failed to loadExportList: %+v", err))
	}
	r := railgun.NewRailGun("获取导出表数据", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "@every 5s"}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			if err := s.loadExportList(); err != nil {
				log.Error("Failed to loadExportList: %+v", err)
			}
			return railgun.MsgPolicyNormal
		}))
	s.exportsRailGun = r
	r.Start()
}

func (s *CommonService) loadModelRule() error {
	result, err := s.dao.ModelFieldRule(context.Background())
	if err != nil {
		return err
	}
	s.rules = result
	return nil
}

func (s *CommonService) initModelRuleRailGun() {
	if err := s.loadModelRule(); err != nil {
		panic(fmt.Sprintf("Failed to loadModelRule: %+v", err))
	}
	r := railgun.NewRailGun("获取模型成员规则", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "@every 1s"}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			if err := s.loadModelRule(); err != nil {
				log.Error("Failed to loadModelRule: %+v", err)
			}
			return railgun.MsgPolicyNormal
		}))
	s.rulesRailGun = r
	r.Start()
}

func (s *CommonService) loadImportList() error {
	result, err := s.dao.ImportAll(context.Background())
	if err != nil {
		return err
	}
	s.imports = result
	return nil
}

func (s *CommonService) initImportsRailGun() {
	if err := s.loadImportList(); err != nil {
		panic(fmt.Sprintf("Failed to loadExportList: %+v", err))
	}
	r := railgun.NewRailGun("获取导入表", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "@every 5s"}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			if err := s.loadImportList(); err != nil {
				log.Error("Failed to loadImportList: %+v", err)
			}
			return railgun.MsgPolicyNormal
		}))
	s.importsRailGun = r
	r.Start()
}

func (s *CommonService) loadModelField() error {
	list, err := s.dao.ModelField(context.Background())
	if err != nil {
		return err
	}
	out := make(map[string][]*api.ModelField)
	for _, item := range list {
		out[item.ModelName] = append(out[item.ModelName], api.ConstructModelField(item))
	}
	s.fields = out
	return nil
}

func (s *CommonService) initFieldsRailGun() {
	if err := s.loadModelField(); err != nil {
		panic(fmt.Sprintf("Failed to loadModelField: %+v", err))
	}
	r := railgun.NewRailGun("获取模型成员", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "@every 1s"}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			if err := s.loadModelField(); err != nil {
				log.Error("Failed to loadModelField: %+v", err)
			}
			return railgun.MsgPolicyNormal
		}))
	s.fieldsRailGun = r
	r.Start()
}
