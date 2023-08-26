package service

import (
	"context"
	"sync"

	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/internal/dao"
	"go-gateway/app/app-svr/app-gw/baas/internal/model"
	"go-gateway/app/app-svr/app-gw/baas/utils/sets"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	//"github.com/dop251/goja"
)

type CommonService struct {
	ac             *paladin.Map
	dao            dao.Dao
	exports        map[string]*api.ExportItem
	rules          map[string]*api.MapperModelFieldRule
	imports        map[int64][]*api.ImportItem
	fields         map[string][]*api.ModelField
	exportsRailGun *railgun.Railgun
	importsRailGun *railgun.Railgun
	rulesRailGun   *railgun.Railgun
	fieldsRailGun  *railgun.Railgun
	customConfig   *CustomConfig
}

type CustomConfig struct {
	JSVM string
}

func newOuterService(d dao.Dao) *CommonService {
	s := &CommonService{
		ac:  &paladin.TOML{},
		dao: d,
	}
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}
	if err := s.ac.Get("customConfig").UnmarshalTOML(&s.customConfig); err != nil {
		panic(err)
	}
	s.initRailGun()
	return s
}

type baasFanoutResult struct {
	fields     []*api.ModelField
	baasImport []*api.ImportItem
	datasource map[string]string
}

type datasourceCtr struct {
	store   map[string]string
	_keySet sets.String
}

func newDatasourceCtr(in map[string]string) *datasourceCtr {
	out := &datasourceCtr{
		store:   in,
		_keySet: sets.StringKeySet(in),
	}
	return out
}

func (d *datasourceCtr) KeySet() sets.String {
	return d._keySet
}

func (d *datasourceCtr) KeyList() []string {
	return d._keySet.List()
}

func (s *CommonService) BaasImpl(ctx context.Context, exportAPI string) (string, error) {
	export, ok := s.exports[exportAPI]
	if !ok {
		return "", ecode.Error(ecode.NothingFound, "找不到对应的导出配置")
	}
	fanoutResult, err := s.doBaasFanoutResult(ctx, export)
	if err != nil {
		log.Error("Failed to doBaasFanoutResult: %+v", err)
		return "", err
	}
	dsCtr := newDatasourceCtr(fanoutResult.datasource)
	out, err := s.exportModel(export.ModelName, dsCtr)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (s *CommonService) generateByOtto(js, source string) (gjson.Result, error) {
	vm := otto.New()
	_, err := vm.Run(js)
	if err != nil {
		return gjson.Result{}, err
	}
	out, err := vm.Call("convert", nil, source)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(out.String()), nil
}

//func (s *CommonService) generateByGoja(js, source string) (gjson.Result, error) {
//	vm := goja.New()
//	if _, err := vm.RunString(js); err != nil {
//		return gjson.Result{}, err
//	}
//	func_, ok := goja.AssertFunction(vm.Get("convert"))
//	if !ok {
//		return gjson.Result{}, errors.New("Failed to AssertFunction")
//	}
//	result, err := func_(goja.Undefined(), vm.ToValue(source))
//	if err != nil {
//		return gjson.Result{}, err
//	}
//	return gjson.Parse(result.String()), nil
//}

func (s *CommonService) exportModel(modelName string, dsCtr *datasourceCtr) (gjson.Result, error) {
	if model.IsGeneric(modelName) {
		return gjson.Result{}, errors.Errorf("请使用 javascript 模式构造该类型: %s", modelName)
	}
	fields, ok := s.fields[modelName]
	if !ok {
		return gjson.Result{}, errors.Errorf("找不到对应的模型成员: %s", modelName)
	}
	frMeta := s.constructFieldRuleMetadatas(fields, dsCtr.KeySet())

	output := ""
	for _, fr := range frMeta {
		apiResp := dsCtr.store[fr.DatasourceApi]
		fieldType := fr.FieldType

		var sourceV gjson.Result
		switch fr.ValueSource {
		case "$all", "":
			sourceV = gjson.Parse(apiResp)
		default:
			sourceV = gjson.Get(apiResp, fr.ValueSource)
		}

		fieldResult := gjson.Result{}
		switch fr.RuleType {
		case model.RuleTypeJavascript:
			result, err := s.generateByJS(fr.ExternalRule, sourceV.String())
			if err != nil {
				log.Error("Failed to parse js: %+v", err)
				continue
			}
			// TODO: format to fr.FieldType
			fieldResult, err = s.formatAsType(result, fieldType)
			if err != nil {
				log.Error("Failed to format js type: %+v", err)
				continue
			}
		case model.RuleTypeLiteral:
			fieldResult = gjson.Parse(fr.ExternalRule)
		case model.RuleTypePrimary:
			fieldResult = sourceV
			if model.IsReference(fieldType) {
				result, err := s.exportModel(fieldType, dsCtr)
				if err != nil {
					log.Error("Failed to exportModel: %+v", err)
					continue
				}
				fieldResult = result
			}
		}
		if !fieldResult.Exists() {
			continue
		}

		jsonFieldName := fr.FieldName
		if fr.JsonAlias != "" {
			jsonFieldName = fr.JsonAlias
		}
		outputPart, err := sjson.SetRaw(output, jsonFieldName, fieldResult.Raw)
		if err != nil {
			log.Error("Failed to set json field: %+v", err)
			continue
		}
		output = outputPart
	}
	return gjson.Parse(output), nil
}

func (s *CommonService) generateByJS(js, source string) (gjson.Result, error) {
	switch s.customConfig.JSVM {
	case "otto":
		return s.generateByOtto(js, source)
	//case "goja":
	//	return s.generateByGoja(js, source)
	default:
		return s.generateByOtto(js, source)
	}
}

func asFieldMap(in []*api.ModelField) map[string]*api.ModelField {
	out := make(map[string]*api.ModelField, len(in))
	for _, v := range in {
		out[v.FieldName] = v
	}
	return out
}

func (s *CommonService) formatAsType(in gjson.Result, dstType string) (gjson.Result, error) {
	if model.IsGeneric(dstType) {
		if !in.IsArray() {
			return gjson.Result{}, errors.Errorf("输入类型应该为 array：%q", in.Type)
		}

		innerType, err := model.SplitGenericType(dstType)
		if err != nil {
			return gjson.Result{}, err
		}
		fields, ok := s.fields[innerType]
		if !ok {
			return gjson.Result{}, errors.Errorf("找不到对应的模型成员: %s", dstType)
		}
		nameToField := asFieldMap(fields)

		arrOutput := "[]"
		in.ForEach(func(key, item gjson.Result) bool {
			if !model.IsReference(innerType) {
				arrOutputP, err := sjson.SetRaw(arrOutput, "-1", item.Raw)
				if err != nil {
					log.Error("Failed to set raw json to array: %+v", err)
					return true
				}
				arrOutput = arrOutputP
				return true
			}
			if !item.IsObject() {
				log.Error("Abort to iterate input, item should be an object: %q", item.Type)
				return false
			}

			itemOutput := ResultForEach(item, nameToField)
			arrOutputP, err := sjson.SetRaw(arrOutput, "-1", itemOutput)
			if err != nil {
				log.Error("Failed to set raw json to array: %+v", err)
				return true
			}
			arrOutput = arrOutputP
			return true
		})
		return gjson.Parse(arrOutput), nil
	}
	// 暂时认为 dstType 不可能是基本类型
	// 目前只有可能是 object
	if !in.IsObject() {
		return gjson.Result{}, errors.Errorf("输入类型应该为 object：%q", in.Type)
	}
	fields, ok := s.fields[dstType]
	if !ok {
		return gjson.Result{}, errors.Errorf("找不到对应的模型成员: %s", dstType)
	}
	nameToField := asFieldMap(fields)
	output := ResultForEach(in, nameToField)
	return gjson.Parse(output), nil
}

func ResultForEach(item gjson.Result, nameToField map[string]*api.ModelField) string {
	itemOutput := ""
	item.ForEach(func(key, value gjson.Result) bool {
		keyS := key.String()
		f, ok := nameToField[keyS]
		if !ok {
			outputP, err := sjson.SetRaw(itemOutput, keyS, value.Raw)
			if err != nil {
				log.Error("Failed to set raw json: %+v", err)
				return true
			}
			itemOutput = outputP
			return true
		}
		jsonFieldName := f.FieldName
		if f.JsonAlias != "" {
			jsonFieldName = f.JsonAlias
		}
		outputP, err := sjson.SetRaw(itemOutput, jsonFieldName, value.Raw)
		if err != nil {
			log.Error("Failed to set raw json: %+v", err)
			return true
		}
		itemOutput = outputP
		return true
	})
	return itemOutput
}

func (s *CommonService) constructFieldRuleMetadatas(fields []*api.ModelField, datasourceAPI sets.String) []*api.FieldRuleMetadata {
	list := make([]*api.FieldRuleMetadata, 0, len(fields))
	for _, field := range fields {
		item := setModelItem(field, s.rules, datasourceAPI)
		list = append(list, item)
	}
	return list
}

func (s *CommonService) doBaasFanoutResult(ctx context.Context, export *api.ExportItem) (*baasFanoutResult, error) {
	out := &baasFanoutResult{
		fields:     s.fields[export.ModelName],
		baasImport: s.imports[export.Id],
	}
	eg := errgroup.WithContext(ctx)
	datasource := make(map[string]string, len(out.baasImport))
	mutex := sync.Mutex{}
	for _, v := range out.baasImport {
		v := v
		switch v.DatasourceType {
		case "HTTP":
			eg.Go(func(ctx context.Context) error {
				response, err := s.dao.RawHttpImpl(ctx, v.DatasourceApi)
				if err != nil {
					log.Error("Failed to raw http: %s, %+v", v.DatasourceApi, err)
					return nil
				}
				mutex.Lock()
				datasource[v.DatasourceApi] = string(response)
				mutex.Unlock()
				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	out.datasource = datasource
	return out, nil
}
