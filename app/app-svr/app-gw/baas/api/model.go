package api

import (
	database_sql "database/sql"
	go_common_library_time "go-common/library/time"
)

type MapperModelFieldRule struct {
	Id            int64                       `json:"id,omitempty"`
	ModelName     string                      `json:"model_name,omitempty"`
	FieldName     string                      `json:"field_name,omitempty"`
	DatasourceApi string                      `json:"datasource_api,omitempty"`
	ExternalRule  database_sql.NullString     `json:"external_rule,omitempty"`
	RuleType      string                      `json:"rule_type,omitempty"`
	ValueSource   string                      `json:"value_source,omitempty"`
	Ctime         go_common_library_time.Time `json:"ctime,omitempty"`
	IsDeleted     int32                       `json:"is_deleted,omitempty"`
}

type MapperModel struct {
	Id          int64                       `json:"id,omitempty"`
	Name        string                      `json:"name,omitempty"`
	Description string                      `json:"description,omitempty"`
	TreeId      int64                       `json:"tree_id,omitempty"`
	Ctime       go_common_library_time.Time `json:"ctime,omitempty"`
	Mtime       go_common_library_time.Time `json:"mtime,omitempty"`
	IsDeleted   int32                       `json:"is_deleted,omitempty"`
}

type BaasExport struct {
	Id        int64                       `json:"id,omitempty"`
	ExportApi string                      `json:"export_api,omitempty"`
	ModelName string                      `json:"model_name,omitempty"`
	Ctime     go_common_library_time.Time `json:"ctime,omitempty"`
	Mtime     go_common_library_time.Time `json:"mtime,omitempty"`
	State     int32                       `json:"state,omitempty"`
	IsDeleted int32                       `json:"is_deleted,omitempty"`
	TreeId    int64                       `json:"tree_id,omitempty"`
}

type BaasImport struct {
	Id             int64  `json:"id,omitempty"`
	BaasExportId   int64  `json:"baas_export_id,omitempty"`
	DatasourceApi  string `json:"datasource_api,omitempty"`
	DatasourceType string `json:"datasource_type,omitempty"`
}

type MapperModelField struct {
	Id        int64                       `json:"id,omitempty"`
	ModelName string                      `json:"model_name,omitempty"`
	FieldName string                      `json:"field_name,omitempty"`
	FieldType string                      `json:"field_type,omitempty"`
	JsonAlias string                      `json:"json_alias,omitempty"`
	Ctime     go_common_library_time.Time `json:"ctime,omitempty"`
	Mtime     go_common_library_time.Time `json:"mtime,omitempty"`
	IsDeleted int32                       `json:"is_deleted,omitempty"`
}

func ConstructModelField(in *MapperModelField) *ModelField {
	return &ModelField{
		Id:        in.Id,
		ModelName: in.ModelName,
		FieldName: in.FieldName,
		FieldType: in.FieldType,
		JsonAlias: in.JsonAlias,
		Ctime:     in.Ctime,
		Mtime:     in.Mtime,
		IsDeleted: in.IsDeleted,
	}
}

func ConstructImportItem(in *BaasImport) *ImportItem {
	return &ImportItem{
		Id:             in.Id,
		BaasExportId:   in.BaasExportId,
		DatasourceApi:  in.DatasourceApi,
		DatasourceType: in.DatasourceType,
	}
}

func ConstructExportItem(in *BaasExport) *ExportItem {
	return &ExportItem{
		Id:        in.Id,
		ExportApi: in.ExportApi,
		ModelName: in.ModelName,
		Ctime:     in.Ctime,
		Mtime:     in.Mtime,
		State:     in.State,
		IsDeleted: in.IsDeleted,
		TreeId:    in.TreeId,
	}
}

func ConstructMapperModelItem(in *MapperModel) *MapperModelItem {
	return &MapperModelItem{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.Description,
		TreeId:      in.TreeId,
		Ctime:       in.Ctime,
		Mtime:       in.Mtime,
		IsDeleted:   in.IsDeleted,
	}
}
