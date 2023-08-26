package apm

import (
	"time"

	xtime "go-common/library/time"
)

const (
	//	字段类型映射
	EventFieldTypeString_Keyword = 0 // es平台的字符串-keyword：用于精确匹配、排序、聚合
	EventFieldTypeByte           = 1
	EventFieldTypeInteger        = 2
	EventFieldTypeLong           = 3
	EventFieldTypeDouble         = 4
	EventFieldTypeDateTime       = 5
	EventFieldTypeFloat          = 6
	EventFieldTypeString_Text    = 7 // es平台的字符串-text: 分词存储用于全文检索, 不能聚合画图分析
	EventFieldTypeUInt64         = 8
	EventFieldTypeMap            = 9

	// 字段类型标识：0基础，1扩展
	EventFieldCommonType   = 0
	EventFieldExtendedType = 1

	EventDelete = 0
	// 字段状态
	EventFieldStateDelete   = -1
	EventFieldStateAdd      = 1
	EventFieldStateModify   = 2
	EventFieldStateReviewed = 3
	// log id
	LogIdTrackT  = "002312" // TrackT
	LogIdInfra   = "002980" // 基础埋点
	LogIdVeda    = "011130" // 堆栈解析
	LogIdIjk     = "002879" // 播放内核
	LogIdPolaris = "001538" // 北极星

	// 外部接口请求 ok status
	BillionsOkStatus   = 0
	DatacenterOkStatus = 200

	// 数据平台字段描述最大长度限制
	DatacenterFieldDescMaxLen = 50
	// 数据平台字段显示名最大长度限制
	DatacenterFieldNameMaxLen = 20

	// 技术埋点监测知
	EventMonitorNotifyOff     = 0 // 通知关
	EventMonitorNotifyOn      = 1 // 通知开
	EventMonitorNotifyMuteOff = 0 // 静默关
	EventMonitorNotifyMuteOn  = 1 // 静默开
)

// Event struct 监控事件
type Event struct {
	ID                     int64         `json:"id"`
	AppKeys                string        `json:"app_keys"`
	BusID                  int64         `json:"bus_id"`
	BusAppKeys             string        `json:"bus_app_keys"`
	Databases              string        `json:"db_name"`
	TableName              string        `json:"table_name"`
	DistributedTableName   string        `json:"distributed_table_name"`
	Name                   string        `json:"name"`
	Description            string        `json:"description"`
	Owner                  string        `json:"owner"`
	Operator               string        `json:"operator"`
	Shared                 int8          `json:"shared"`
	State                  int8          `json:"state"`
	EventFields            []*EventField `json:"event_fields"`
	CommonEventFields      []*EventField `json:"common_event_fields"`
	Ctime                  int64         `json:"ctime"`
	Mtime                  int64         `json:"mtime"`
	LogID                  string        `json:"log_id"`
	BusName                string        `json:"bus_name"`
	Topic                  string        `json:"kafka_topic"`
	Activity               int8          `json:"is_activity"`
	SampleRate             int           `json:"sample_rate"`
	LowestSampleRate       float64       `json:"lowest_sample_rate"`
	DatacenterEventID      int64         `json:"datacenter_event_id"`
	DatacenterAppID        int64         `json:"datacenter_app_id"`
	IsReviewed             bool          `json:"is_reviewed"`
	DataCount              int64         `json:"data_count"`
	Level                  int8          `json:"level"`
	DatacenterDwdTableName string        `json:"datacenter_dwd_table_name"`
	IsWideTable            int8          `json:"is_wide_table"`
	StorageCount           int64         `json:"storage_count"`
	StorageCapacity        int64         `json:"storage_capacity"`
}

// EventField struct 监控事件字段表
type EventField struct {
	ID                     int64  `json:"id" form:"id"`
	EventID                int64  `json:"event_id" form:"event_id"`
	Key                    string `json:"field_key" form:"field_key"`
	Description            string `json:"description" form:"description"`
	Example                string `json:"example" form:"example"`
	Type                   int8   `json:"field_type" form:"field_type"`
	Mode                   int8   `json:"mode" form:"mode"`
	DefaultValue           string `json:"default_value" form:"default_value"`
	Index                  int64  `json:"field_index" form:"field_index"`
	State                  int8   `json:"state" form:"state"`
	Operator               string `json:"operator" form:"operator"`
	Ctime                  int64  `json:"ctime" form:"ctime"`
	Mtime                  int64  `json:"mtime" form:"mtime"`
	IsClickhouse           int8   `json:"is_clickhouse" form:"is_clickhouse"`
	ISElasticsearchIndex   int8   `json:"is_elasticsearch_index" form:"is_elasticsearch_index"`
	ElasticSearchFieldType int8   `json:"elasticsearch_field_type" form:"elasticsearch_field_type"`
}

// EventFieldReq struct
type EventFieldReq struct {
	EventID          int64         `json:"event_id" form:"event_id"`
	Fields           []*EventField `json:"event_fields" form:"event_fields"`
	CommonFieldsFlag int64         `json:"common_fields_flag" form:"common_fields_flag"`
	IsIgnoreBillions int8          `json:"is_ignore_billions" form:"is_ignore_billions"`
	Operator         string        `json:"operator" form:"operator"`
	Pn               int           `json:"pn" form:"pn"`
	Ps               int           `json:"ps" form:"ps"`
}

// ParamsEvent struct.
type ParamsEvent struct {
	BusID                  int64         `json:"bus_id"`
	EventID                int64         `json:"event_id"`
	Name                   string        `json:"name"`
	AppKeys                string        `json:"app_keys"`
	AppKey                 string        `json:"app_key"`
	Description            string        `json:"description"`
	Owner                  string        `json:"owner"`
	LogID                  string        `json:"log_id"`
	Databases              string        `json:"db_name"`
	TableName              string        `json:"table_name"`
	DistributedTableName   string        `json:"distributed_table_name"`
	Topic                  string        `json:"kafka_topic"`
	Activity               int8          `json:"is_activity"`
	Shared                 int           `json:"shared"`
	Fields                 []*EventField `json:"event_fields"`
	State                  int8          `json:"state"`
	SampleRate             int           `json:"sample_rate"`
	LowestSampleRate       float64       `json:"lowest_sample_rate"`
	DatacenterEventID      int64         `json:"datacenter_event_id"`
	DatacenterAppID        int64         `json:"datacenter_app_id"`
	DatacenterDwdTableName string        `json:"datacenter_dwd_table_name"`
	DataCount              int64         `json:"data_count"`
	Level                  int8          `json:"level"`
	IsWideTable            int8          `json:"is_wide_table"`
	IsIgnoreBillions       int8          `json:"is_ignore_billions"`
}

// EventAdvanced struct
type EventAdvanced struct {
	ID           int64      `json:"id" form:"id"`
	EventID      int64      `json:"event_id" form:"event_id"`
	FieldName    string     `json:"field_name" form:"field_name"`
	Title        string     `json:"title" form:"title"`
	Description  string     `json:"description" form:"description"`
	DisplayType  int8       `json:"display_type" form:"display_type"`
	QueryType    string     `json:"query_type" form:"query_type"`
	MappingGroup string     `json:"mapping_group" form:"mapping_group"`
	CustomSql    string     `json:"custom_sql" form:"custom_sql"`
	Operator     string     `json:"operator" form:"operator"`
	Ctime        xtime.Time `json:"ctime" form:"ctime"`
	Mtime        xtime.Time `json:"mtime" form:"mtime"`
}

// ResultEventList struct.
type ResultEventList struct {
	PageInfo *Page    `json:"page,omitempty"`
	Items    []*Event `json:"items,omitempty"`
}

// DatacenterEvent struct
type DatacenterEvent struct {
	ID                 int64              `json:"id,omitempty" form:"id"`
	LogID              string             `json:"logId" form:"logId"`
	EventCode          string             `json:"eventCode" form:"eventCode"`
	EventName          string             `json:"eventName" form:"eventName"`
	EventType          string             `json:"eventType" form:"eventType"`
	ProID              string             `json:"proId" form:"proId"`
	BizLine            string             `json:"bizLine" form:"bizLine"`
	Topic              string             `json:"topic" form:"topic"`
	DataWarehouseTable string             `json:"dataWarehouseTable" form:"dataWarehouseTable"`
	EventStatus        string             `json:"eventStatus" form:"eventStatus"`
	Fields             []*DatacenterField `json:"appEventSysFields"`
}

type DatacenterField struct {
	FieldId   string `json:"fieldId"`
	FieldName string `json:"fieldName"`
	FieldType string `json:"fieldType"`
	FieldDesc string `json:"fieldDesc"`
}

func (df *DatacenterField) String() string {
	return "{fieldId:" + df.FieldId + "," + "fieldName:" + df.FieldName + "," + "fieldType:" + df.FieldType + "," + "fieldDesc:" + df.FieldDesc + "}"
}

type EventFieldGroup struct {
	CommonFields   []*EventField `json:"common_fields"`
	ExtendedFields []*EventField `json:"event_fields"`
}

// BillionsEvent struct
type BillionsEvent struct {
	TreeID           string `json:"treeId" form:"treeId"`
	AppID            string `json:"appId" form:"appId"`
	AppName          string `json:"appName" form:"appName"`
	ServicePrincipal string `json:"servicePrincipal" form:"servicePrincipal"`
	DeployLocations  string `json:"deployLocations" form:"deployLocations"`
}

// BillionsEventField struct
type BillionsEventField struct {
	Name string `json:"name" form:"name"`
	Type string `json:"type" form:"type"`
}

func (bef *BillionsEventField) String() string {
	return "{name:" + bef.Name + "," + "type:" + bef.Type + "}"
}

// BillionsEventFieldMapping struct
type BillionsEventFieldMapping struct {
	AppID  string                `json:"appId" form:"appId"`
	Fields []*BillionsEventField `json:"fields" form:"fields"`
}

// EventFieldSql struct
type EventFieldSql struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Index int64  `json:"index"`
	Desc  string `json:"desc"`
}

type EventSqlTemplate struct {
	DBName       string
	TableName    string
	DisTableName string
	Fields       []*EventFieldSql
}

// EventSqlRes struct
type EventSqlRes struct {
	CreateSql    string `json:"create_sql"`
	CreateSqlDis string `json:"create_sql_dis"`
}

type EventMonitor struct {
	EventId         int64  `json:"event_id" form:"event_id"`
	EventName       string `json:"event_name" form:"event_name"`
	DatacenterAppId string `json:"datacenter_app_id" form:"datacenter_app_id"`
	StorageCount    int64  `json:"storage_count" form:"storage_count"`
	StorageCapacity int64  `json:"storage_capacity" form:"storage_capacity"`
}

type EventMonitorDB struct {
	ID               int64     `json:"id" form:"id"`
	EventId          int64     `json:"event_id" form:"event_id"`
	BusId            int64     `json:"bus_id" form:"bus_id"`
	BusName          string    `json:"bus_name" form:"bus_name"`
	LogCount         int64     `json:"log_count" form:"log_count"`
	LogTime          time.Time `json:"log_time" form:"log_time"`
	LogCapacity      int64     `json:"log_capacity" form:"log_capacity"`
	LogIndexOpenTime int64     `json:"log_index_open_time" form:"log_index_open_time"`
	LogRetentionTime int64     `json:"log_retention_time" form:"log_retention_time"`
	Operator         string    `json:"operator" form:"operator"`
	Ctime            time.Time `json:"ctime" form:"ctime"`
	Mtime            time.Time `json:"mtime" form:"mtime"`
}

type EventMonitorResp struct {
	PageInfo *Page             `json:"page,omitempty"`
	Items    []*EventMonitorDB `json:"items,omitempty"`
}

type BillionsQueryBodyTemplate struct {
	From       int64
	Size       int64
	Sort       *BillionsQuerySort
	Query      string
	RangeFiled *BillionsQueryRange
	Aggs       []*BillionsQueryAggs
}

type BillionsQuerySort struct {
	Name  string
	Value string
	Type  string
}

type BillionsQueryRange struct {
	StartTime int64
	EndTime   int64
}

type BillionsQueryAggs struct {
	Type      string
	FieldName string
}

type BillionsLifecycle struct {
	AppId            string `json:"appId,omitempty"`
	Capacity         string `json:"capacity,omitempty"`
	IndexOpenTime    int64  `json:"indexOpenTime,omitempty"`
	LogRetentionTime int64  `json:"logRetentionTime,omitempty"`
}

type EventFieldFile struct {
	Id                     int64     `json:"id"`
	EventId                int64     `json:"event_id"`
	FieldId                int64     `json:"field_id"`
	FieldKey               string    `json:"field_key"`
	FieldType              int8      `json:"field_type"`
	Description            string    `json:"description"`
	Example                string    `json:"example"`
	Type                   int8      `json:"type"`
	DefaultValue           string    `json:"default_value"`
	IsClickhouse           int8      `json:"is_clickhouse"`
	IsElasticsearchIndex   int8      `json:"is_elasticsearch_index"`
	ElasticsearchFieldType int8      `json:"elasticsearch_field_type"`
	FieldState             int8      `json:"field_state"`
	FieldIndex             int64     `json:"field_index"`
	FieldVersion           int64     `json:"field_version"`
	Operator               string    `json:"operator"`
	CTime                  time.Time `json:"ctime"`
	MTime                  time.Time `json:"mtime"`
}

func (file *EventFieldFile) FileConvertToField() *EventField {
	return &EventField{
		ID:                     file.FieldId,
		EventID:                file.EventId,
		Key:                    file.FieldKey,
		Description:            file.Description,
		Example:                file.Example,
		Type:                   file.FieldType,
		Mode:                   file.Type,
		DefaultValue:           file.DefaultValue,
		Index:                  file.FieldIndex,
		State:                  file.FieldState,
		Operator:               file.Operator,
		IsClickhouse:           file.IsClickhouse,
		ISElasticsearchIndex:   file.IsElasticsearchIndex,
		ElasticSearchFieldType: file.ElasticsearchFieldType,
		Ctime:                  file.CTime.Unix(),
		Mtime:                  file.MTime.Unix(),
	}
}

type EventFieldDiff struct {
	Old   *EventField `json:"old"`
	New   *EventField `json:"new"`
	State int8        `json:"state"`
}

type EventFieldPublish struct {
	Id           int64     `json:"id"`
	EventId      int64     `json:"event_id"`
	FieldVersion int64     `json:"version"`
	Operator     string    `json:"operator"`
	CTime        time.Time `json:"ctime"`
	MTime        time.Time `json:"mtime"`
}

type EventFieldPublishResp struct {
	PageInfo *Page                `json:"page,omitempty"`
	Items    []*EventFieldPublish `json:"items,omitempty"`
}

type AppEventCommonField struct {
	Id                     int64     `json:"id"`
	AppKey                 string    `json:"app_key"`
	GroupId                int64     `json:"group_id"`
	Key                    string    `json:"field_key"`
	Type                   int8      `json:"field_type"`
	Index                  int64     `json:"field_index"`
	Description            string    `json:"description"`
	DefaultValue           string    `json:"default_value"`
	State                  int8      `json:"state"`
	IsClickhouse           int8      `json:"is_clickhouse"`
	IsElasticsearchIndex   int8      `json:"is_elasticsearch_index"`
	ElasticsearchFieldType int8      `json:"elasticsearch_field_type"`
	Operator               string    `json:"operator"`
	CTime                  time.Time `json:"ctime"`
	MTime                  time.Time `json:"mtime"`
}

type EventCommonFieldGroup struct {
	Id          int64                  `json:"id"`
	AppKey      string                 `json:"app_key"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsDefault   int8                   `json:"is_default"`
	Fields      []*AppEventCommonField `json:"fields"`
	Operator    string                 `json:"operator"`
	CTime       time.Time              `json:"ctime"`
	MTime       time.Time              `json:"mtime"`
}

type EventCommonFieldGroupResp struct {
	PageInfo *Page                    `json:"page,omitempty"`
	Items    []*EventCommonFieldGroup `json:"items,omitempty"`
}

type EventCommonFieldGroupReq struct {
	Id          int64                  `json:"group_id"`
	AppKey      string                 `json:"app_key" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description" validate:"required"`
	IsDefault   int8                   `json:"is_default"`
	Fields      []*AppEventCommonField `json:"fields"`
	Operator    string                 `json:"operator"`
}

// CKTableCreateReq struct 数据平台建表请求
type CKTableCreateReq struct {
	EventId     int64  `json:"event_id" validate:"required"`
	TTLDur      int64  `json:"ttl_duration" validate:"required"`
	TTLExp      string `json:"ttl_dateTimeExp" validate:"required"`
	OrderBy     string `json:"order_by" validate:"required"`
	PartitionBy string `json:"partition_by" validate:"required"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
}

// DatacenterOpenAPI struct 数据平台的openapi
type DatacenterOpenAPI struct {
	Account   string `json:"account"`
	APIName   string `json:"apiName"`
	AppId     string `json:"appId"`
	Data      string `json:"data"`
	GroupName string `json:"groupName"`
	RequestId string `json:"requestId"`
	Signature string `json:"signature"`
}

// CKTableCreateData struct 数据平台clickhouse建表数据
type CKTableCreateData struct {
	BasicModule    *CKBasicModule   `json:"basicModule"`
	Cols           []*CKCol         `json:"cols"`
	CfgModule      *CKCfgModule     `json:"configurationModule"`
	BusModule      *CKBusModule     `json:"businessModule"`
	ContentModule  *CKContentModule `json:"contentModule"`
	ModelModule    *CKModelModule   `json:"modelModule"`
	DataSourceType string           `json:"dsType"`
}

type CKBasicModule struct {
	DSName      string `json:"dsName"`
	DBName      string `json:"dbName"`
	TabName     string `json:"tabName"`
	TabDesc     string `json:"tabDesc,omitempty"`
	TTLUnit     string `json:"unit"`
	TTLDuration int64  `json:"dataDuration"`
	TTLDataExpr string `json:"dateTimeExpr"`
	Operator    string `json:"userName"`
}

type CKCol struct {
	Name string `json:"colName"`
	Type string `json:"colType"`
	Desc string `json:"colDesc,omitempty"`
}

type CKCfgModule struct {
	CustomSet   []*CKCustomCfg    `json:"customSettings"`
	OrderBy     []*CKCfgExp       `json:"orderBy"`
	PartitionBy []*CKCfgExp       `json:"partitionBy"`
	ShardingKey *CKCfgShardingKey `json:"shardingKey"`
	Engine      *CKCfgEngine      `json:"engine"`
}

type CKCustomCfg struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CKCfgExp struct {
	Expression string `json:"expression"`
}

type CKCfgShardingKey struct {
	Func string `json:"func"`
}

type CKCfgEngine struct {
	Type string `json:"type"`
}

type CKBusModule struct {
	BusTag   *CKBusTag `json:"businessTags"`
	PubLevel int64     `json:"publicLevel"`
}

type CKBusTag struct {
	Items []string `json:"item"`
}

type CKContentModule struct {
	DataLevel string `json:"dataLevel"`
}

type CKModelModule struct {
	ModeLevel int64 `json:"modelLevel"`
}

type EventDatacenterRel struct {
	EventId           int64  `json:"event_id" form:"event_id" validate:"required"`
	DatacenterAppId   int64  `json:"datacenter_app_id" form:"datacenter_app_id" validate:"required"`
	DatacenterEventId int64  `json:"datacenter_event_id" form:"datacenter_event_id" validate:"required"`
	Operator          string `json:"operator" form:"operator"`
}

type EventCompletion struct {
	Id                  int64  `json:"id"`
	DatacenterEventName string `json:"datacenter_event_name"`
	DatacenterAppId     int64  `json:"datacenter_app_id"`
	Count               int64  `json:"cnt"`
	LogDate             string `json:"log_date"`
}

// AddEventSampleRateReq 新增采样率Req
type AddEventSampleRateReq struct {
	DatacenterAppId int64   `json:"datacenter_app_id" form:"datacenter_app_id"`
	AppKey          string  `json:"app_key" form:"app_key"`
	EventId         string  `json:"event_id" form:"event_id"`
	EventName       string  `json:"event_name" form:"event_name"`
	Rate            float64 `json:"rate" form:"rate"`
	LogId           string  `json:"log_id" form:"log_id"`
}

// DeleteEventSampleRateReq 删除采样率Req
type DeleteEventSampleRateReq struct {
	Items []*DeleteSampleItem `protobuf:"bytes,1,opt,name=items,proto3" json:"items" form:"items"`
}

type DeleteSampleItem struct {
	DatacenterAppId int64  `protobuf:"varint,1,opt,name=DatacenterAppId,proto3" json:"datacenter_app_id" form:"datacenter_app_id"`
	AppKey          string `protobuf:"varint,2,opt,name=AppKey,proto3" json:"app_key" form:"app_key"`
	EventId         string `protobuf:"bytes,3,opt,name=EventId,proto3" json:"event_id" form:"event_id"`
	LogId           string `protobuf:"bytes,4,opt,name=LogId,proto3" json:"log_id" form:"log_id"`
}

// EventSampleRateListReq 列表页Req
type EventSampleRateListReq struct {
	AppKey  string `json:"app_key,omitempty" form:"app_key"`
	EventId string `json:"event_id,omitempty" form:"event_id"`
	LogId   string `json:"log_id,omitempty" form:"log_id"`
}

// EventSampleRateListResp 列表页Resp
type EventSampleRateListResp struct {
	Items []*EventSampleRateItem `json:"items"`
}

type EventSampleRateItem struct {
	Id              int64   `json:"id,omitempty"`
	AppKey          string  `json:"app_key"`
	DatacenterAppId int64   `json:"datacenter_app_id"`
	EventId         string  `json:"event_id"`
	EventName       string  `json:"event_name"`
	Rate            float64 `json:"rate"`
	Ctime           int64   `json:"ctime" form:"ctime"`
	Mtime           int64   `json:"mtime" form:"mtime"`
	LogId           string  `json:"log_id" form:"log_id"`
	Operator        int64   `json:"operator"`
	IsTemporary     int64   `json:"is_temporary"`
}

// EventSampleRateConfigReq 获取采样率配置req
type EventSampleRateConfigReq struct {
	AppKey string `json:"app_key" form:"app_key"`
}

// EventSampleRate apm_event_sample_rate表 事件上报采样率表
type EventSampleRate struct {
	Id              uint      `json:"id"`                // 自增ID
	DatacenterAppId int64     `json:"datacenter_app_id"` // 数据平台appid
	AppKey          string    `json:"app_key"`           // appkey
	EventId         string    `json:"event_id"`          // 事件id
	SampleRate      float64   `json:"sample_rate"`       // 采样率
	EventName       string    `json:"event_name"`        // 事件名称
	Mtime           time.Time `json:"mtime"`             // 修改时间
	Ctime           time.Time `json:"ctime"`             // 创建时间
	LogId           string    `json:"log_id"`            // 日志id
}

// EventSampleRateApp apm_event_sample_rate_app表 事件上报采样率表(app维度)
type EventSampleRateApp struct {
	Id          uint      `json:"id"`           // 自增ID
	AppKey      string    `json:"app_key"`      // APP在平台内唯一标识,多个英文逗号隔开
	EventId     string    `json:"event_id"`     // 事件id
	SampleRate  float64   `json:"sample_rate"`  // 采样率
	EventName   string    `json:"event_name"`   // 事件名称
	Mtime       time.Time `json:"mtime"`        // 修改时间
	Ctime       time.Time `json:"ctime"`        // 创建时间
	LogId       string    `json:"log_id"`       // 日志id
	IsTemporary int64     `json:"is_temporary"` // 是否临时注入
}

type EventSampleRateConfigResp struct {
	EventRates string `json:"event_rates,omitempty"`
}

type EventMonitorNotifyConfig struct {
	Id            int64     `json:"id"`
	EventId       int64     `json:"event_id"`
	AppKey        string    `json:"app_key"`
	IsNotify      int8      `json:"is_notify"`
	IsMute        int8      `json:"is_mute"`
	MuteStartTime time.Time `json:"mute_start_time"`
	MuteEndTime   time.Time `json:"mute_end_time"`
	Operator      string    `json:"operator"`
	CTime         time.Time `json:"ctime"`
	MTime         time.Time `json:"mtime"`
}

type EventMonitorNotifyConfigReq struct {
	EventId int64  `json:"event_id" form:"event_id" validate:"required"`
	AppKey  string `json:"app_key" form:"app_key" validate:"required"`
}

type EventMonitorNotifyConfigListReq struct {
	EventId  int64  `json:"event_id" form:"event_id"`
	AppKey   string `json:"app_key" form:"app_key"`
	IsNotify int8   `json:"is_notify" form:"is_notify"`
	IsMute   int8   `json:"is_mute" form:"is_mute"`
	Pn       int    `json:"pn" form:"pn" default:"1" validate:"min=1"`
	Ps       int    `json:"ps" form:"ps" default:"20" validate:"min=1"`
}

type EventMonitorNotifyConfigListResp struct {
	PageInfo *Page                       `json:"page,omitempty"`
	Items    []*EventMonitorNotifyConfig `json:"items,omitempty"`
}

type EventMonitorNotifyConfigSetReq struct {
	EventId       int64     `json:"event_id" validate:"required"`
	AppKey        string    `json:"app_key" validate:"required"`
	IsNotify      int8      `json:"is_notify"`
	IsMute        int8      `json:"is_mute"`
	MuteStartTime time.Time `json:"mute_start_time"`
	MuteEndTime   time.Time `json:"mute_end_time"`
	Operator      string    `json:"operator"`
}

const (
	// template
	EventTemplateCreateSql = `CREATE TABLE {{.DBName}}.{{.TableName}} ON CLUSTER 'Clickhouse_datacenter_olap_ck_mobile_infra_replica'
(
{{- $mIndex := maxIndex}}
{{range $index, $field := .Fields}}
	{{- if eq $mIndex $index}}	{{$field.Name}} {{$field.Type}}
	{{- else}}	{{$field.Name}} {{$field.Type}},
	{{- end}}
{{end -}}
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/{{.TableName}}','{replica}')
PARTITION BY (【分区字段】)
ORDER BY (【一级索引】)
TTL toDate(【TTL字段】) + toIntervalDay(【TTL周期】)
SETTINGS index_granularity = 8192, storage_policy = 'hot_and_cold'`
	EventTemplateCreateSqlDis = `CREATE TABLE {{.DBName}}.{{.DisTableName}} ON CLUSTER 'Clickhouse_datacenter_olap_ck_mobile_infra_replica'
(
{{- $mIndex := maxIndex}}
{{range $index, $field := .Fields -}}
	{{- if eq $mIndex $index}}	{{$field.Name}} {{$field.Type}}
	{{- else}}	{{$field.Name}} {{$field.Type}},
	{{- end}}
{{end -}}
)
ENGINE = Distributed('Clickhouse_datacenter_olap_ck_mobile_infra_replica', '{{.DBName}}', '{{.TableName}}', rand())`

	BillionsTemplateQueryBody = `{
	"from": {{.From}},
    "size": {{.Size}},
	{{- if .Sort}}
	"sort": [
			{
				"{{.Sort.Name}}": {
					"order": "{{.Sort.Value}}",
					"unmapped_type": "{{.Sort.Type}}"
				}
			}
	],
	{{- end}}
	"_source": "True",
    "query": {
        "bool": {
            "filter": [
			{{- if .Query}}
                {
                    "query_string": {
                        "query": "{{.Query}}"
                    }
                },
			{{- end}}
                {
                    "range": {
                        "@timestamp": {
                            "gte": {{.RangeFiled.StartTime}},
                            "lte": {{.RangeFiled.EndTime}},
							"format": "epoch_millis"
                        }
                    }
                }
            ]
        }
    }{{- if .Aggs}},{{end}}
	{{- $mIndex := maxIndex}}
	{{- range $index,$agg := .Aggs}}
	{{- if eq $mIndex $index}}
	"aggs": {
		"{{$agg.FieldName}}": {
			"{{$agg.Type}}": {
				"field": "{{$agg.FieldName}}"
				}
			}
	}
	{{- else}}
	"aggs": {
		"{{$agg.FieldName}}": {
			"{{$agg.Type}}": {
				"field": "{{$agg.FieldName}}"
				}
			}
	},
	{{- end}}
	{{- end}}
}`
)
