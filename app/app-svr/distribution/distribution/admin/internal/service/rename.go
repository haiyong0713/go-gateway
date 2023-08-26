package service

import (
	"context"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model/rename"
	tmm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultiple"
	tme "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultipleedit"
	"go-gateway/app/app-svr/distribution/distribution/admin/tool"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/log"

	"github.com/pkg/errors"
)

var renameTypeValidator = map[int64]func(in string) bool{
	1: validateAbtestFlagValue,
	2: validateMultipleTus,
}

func (s *Service) Rename(ctx context.Context, req *rename.Rename) error {
	validate, ok := renameTypeValidator[req.Type]
	if !ok {
		return errors.Errorf("wrong type")
	}
	if !validate(req.ID) {
		return errors.Errorf("wrong id")
	}
	renameInfo, err := s.renameDao.FetchRenameInfo(ctx, req.ID)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	mergeRenameInfo(renameInfo, req)
	if err := s.renameDao.Rename(ctx, req); err != nil {
		log.Error("%+v", err)
		return err
	}
	s.AsyncLog(ctx)
	return nil
}

func mergeRenameInfo(old *rename.Rename, new *rename.Rename) {
	if old.Title != "" && new.Title == "" {
		new.Title = old.Title
	}
	if len(old.TabNames) != 0 && len(new.TabNames) == 0 {
		new.TabNames = make(map[string]string, len(old.TabNames))
	}
	for tusValue, v := range old.TabNames {
		if _, ok := new.TabNames[tusValue]; !ok {
			new.TabNames[tusValue] = v
		}
	}
}

func validateMultipleTus(fieldName string) bool {
	meta, ok := preferenceproto.TryGetPreference(multipleTusPreference)
	if !ok {
		return false
	}
	for _, v := range meta.ProtoDesc.GetFields() {
		if v.GetFullyQualifiedName() == fieldName {
			return true
		}
	}
	return false
}

func validateAbtestFlagValue(flagValue string) bool {
	flagValues, err := tool.GetFiledOptionValuesFromPreferenceproto(abtestPreference, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsABTestFlagValue)
	if err != nil {
		return false
	}
	for _, v := range flagValues {
		if flagValue == v {
			return true
		}
	}
	return false
}

func (s *Service) rename(ctx context.Context, fieldNames []string, renameHandler renameHandler) {
	renameMap, err := s.renameDao.BatchFetchRenameInfo(ctx, fieldNames)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	renameHandler.Handle(renameMap)
}

type renameHandler interface {
	Handle(renameMap map[string]*rename.Rename)
}

type multipleTusTitleRenamer struct {
	fieldInfos []*tmm.FieldInfo
}

func (m *multipleTusTitleRenamer) Handle(renameMap map[string]*rename.Rename) {
	for _, v := range m.fieldInfos {
		ri, ok := renameMap[v.Name]
		if !ok || ri.Title == "" { //没有更改过或者没有修改title
			continue
		}
		v.Descriptor = ri.Title
	}
}

type multipleTusTabsRenamer struct {
	fieldName string
	details   []*tmm.Detail
}

func (m *multipleTusTabsRenamer) Handle(renameMap map[string]*rename.Rename) {
	rename, ok := renameMap[m.fieldName]
	if !ok {
		return
	}
	if rename == nil || len(rename.TabNames) == 0 {
		return
	}
	for _, detail := range m.details {
		tabName, ok := rename.TabNames[detail.TusValue]
		if !ok || tabName == "" {
			continue
		}
		detail.TusName = tabName
	}
}

type multipleTusEditRenamer struct {
	overviews []*tme.Overview
}

func (m *multipleTusEditRenamer) Handle(renameMap map[string]*rename.Rename) {
	for _, v := range m.overviews {
		renameInfo, ok := renameMap[v.FieldName]
		if !ok {
			continue
		}
		if renameInfo.Title != "" {
			v.Name = renameInfo.Title
		}
		if len(renameInfo.TabNames) == 0 {
			return
		}
		for _, performance := range v.Performances {
			tusName, ok := renameInfo.TabNames[performance.TusValue]
			if !ok {
				continue
			}
			if tusName == "" {
				continue
			}
			performance.Text = tusName
		}
	}
}
