package http

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func apmBusList(c *bm.Context) {
	var (
		params            = c.Request.Form
		filterKey, appKey string
		pn, ps            int
		err               error
	)
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	appKey = params.Get("app_key")
	filterKey = params.Get("filter_key")
	c.JSON(s.ApmSvr.ApmBusList(c, appKey, filterKey, ps, pn))
}

func apmBusAdd(c *bm.Context) {
	var (
		params                                                                                      = c.Request.Form
		res                                                                                         = map[string]interface{}{}
		name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames string
		shared                                                                                      int
		err                                                                                         error
	)
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKeys = params.Get("app_keys"); appKeys == "" {
		res["message"] = "app_keys异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if owner = params.Get("owner"); owner == "" {
		res["message"] = "owner异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if datacenterBusinessKey = params.Get("datacenter_bus_key"); datacenterBusinessKey == "" {
		res["message"] = "datacenter_bus_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if shared, err = strconv.Atoi(params.Get("shared")); err != nil {
		shared = 0
	}
	if datacenterDwdTableNames = params.Get("datacenter_dwd_table_names"); datacenterDwdTableNames == "" {
		res["message"] = "datacenter_dwd_table_names 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmBusAdd(c, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames, shared))
}

func apmBusDel(c *bm.Context) {
	var (
		params   = c.Request.Form
		res      = map[string]interface{}{}
		userName string
		busId    int64
		err      error
	)
	if busId, err = strconv.ParseInt(params.Get("bus_id"), 10, 64); err != nil {
		res["message"] = "bus_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmBusDel(c, busId, userName))
}

func apmBusUpdate(c *bm.Context) {
	var (
		params                                                                                      = c.Request.Form
		res                                                                                         = map[string]interface{}{}
		name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames string
		BusId                                                                                       int64
		shared                                                                                      int
		err                                                                                         error
	)
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if owner = params.Get("owner"); owner == "" {
		res["message"] = "owner异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if datacenterBusinessKey = params.Get("datacenter_bus_key"); datacenterBusinessKey == "" {
		res["message"] = "datacenter_bus_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKeys = params.Get("app_keys"); appKeys == "" {
		res["message"] = "app_keys异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if BusId, err = strconv.ParseInt(params.Get("bus_id"), 10, 64); err != nil {
		res["message"] = "bus_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if shared, err = strconv.Atoi(params.Get("shared")); err != nil {
		shared = 0
	}
	if datacenterDwdTableNames = params.Get("datacenter_dwd_table_names"); datacenterDwdTableNames == "" {
		res["message"] = "datacenter_dwd_table_names 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmBusUpdate(c, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames, BusId, shared))
}

// Event
func apmEvent(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		eventID int64
		err     error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEvent(c, eventID))
}

func apmEventList(c *bm.Context) {
	var (
		params                                                                = c.Request.Form
		appKey, logID, busName, name, topic, tableName, orderBy, dwdTableName string
		pn, ps                                                                int
		busId, activityConv, dtConditionConv, appId, stateConv                int64
		activity, dtCondition, state                                          int8
		err                                                                   error
	)
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	if busId, err = strconv.ParseInt(params.Get("bus_id"), 10, 64); err != nil {
		busId = 0
	}
	if activityConv, err = strconv.ParseInt(params.Get("is_activity"), 10, 8); err != nil {
		activityConv = 0
	}
	if dtConditionConv, err = strconv.ParseInt(params.Get("db_table_condition"), 10, 8); err != nil {
		dtConditionConv = 0
	}
	if appId, err = strconv.ParseInt(params.Get("datacenter_app_id"), 10, 64); err != nil {
		appId = 0
	}
	if stateConv, err = strconv.ParseInt(params.Get("state"), 10, 64); err != nil {
		stateConv = 0
	}
	activity = int8(activityConv)
	dtCondition = int8(dtConditionConv)
	state = int8(stateConv)
	appKey = params.Get("app_key")
	logID = params.Get("log_id")
	busName = params.Get("bus_name")
	name = params.Get("name")
	topic = params.Get("kafka_topic")
	tableName = params.Get("table_name")
	orderBy = params.Get("order_by")
	dwdTableName = params.Get("datacenter_dwd_table_name")
	c.JSON(s.ApmSvr.ApmEventList(c, name, appKey, logID, busName, topic, tableName, orderBy, dwdTableName, ps, pn, busId, appId, activity, dtCondition, state))
}

func apmEventAdd(c *bm.Context) {
	var (
		res                                                                                                                      = map[string]interface{}{}
		name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string
		shared, sampleRate                                                                                                       int
		activity, level, isWideTable, isIgnoreBillions                                                                           int8
		busId, datacenterAppID, dataCount                                                                                        int64
		lowestSampleRate                                                                                                         float64
		bs                                                                                                                       []byte
		err                                                                                                                      error
	)
	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	// params
	var cs *apm.ParamsEvent
	if err = json.Unmarshal(bs, &cs); err != nil {
		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if logID = cs.LogID; logID == "" {
		res["message"] = "log_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = cs.Name; name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = cs.Description; description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if owner = cs.Owner; owner == "" {
		res["message"] = "owner异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if busId = cs.BusID; busId == 0 {
		res["message"] = "bus_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if datacenterAppID = cs.DatacenterAppID; datacenterAppID == 0 {
		res["message"] = "datacenter_app_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if sampleRate = cs.SampleRate; sampleRate == 0 {
		sampleRate = 10000
	}
	if lowestSampleRate = cs.LowestSampleRate; lowestSampleRate == 0 {
		lowestSampleRate = 1
	}
	dwdTableName = cs.DatacenterDwdTableName
	appKey = cs.AppKey
	appKeys = cs.AppKeys
	dbName = cs.Databases
	tableName = cs.TableName
	distributedTableName = cs.DistributedTableName
	shared = cs.Shared
	topic = cs.Topic
	activity = cs.Activity
	level = cs.Level
	dataCount = cs.DataCount
	isWideTable = cs.IsWideTable
	isIgnoreBillions = cs.IsIgnoreBillions
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.ApmSvr.ApmEventAdd(c, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName, shared, sampleRate, busId, datacenterAppID, dataCount, lowestSampleRate, activity, level, isWideTable, isIgnoreBillions))
}

func apmEventDel(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		eventID int64
		err     error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventDel(c, eventID))
}

func apmEventUpdate(c *bm.Context) {
	var (
		res                                                                                                              = map[string]interface{}{}
		appKeys, description, name, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string
		shared, sampleRate                                                                                               int
		activity, state, level, isWideTable                                                                              int8
		eventId, datacenterAppID, busId, dataCount                                                                       int64
		lowestSampleRate                                                                                                 float64
		bs                                                                                                               []byte
		err                                                                                                              error
	)
	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	// params
	var cs *apm.ParamsEvent
	if err = json.Unmarshal(bs, &cs); err != nil {
		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if eventId = cs.EventID; eventId == 0 {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if logID = cs.LogID; logID == "" {
		res["message"] = "log_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = cs.Name; name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = cs.Description; description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if owner = cs.Owner; owner == "" {
		res["message"] = "owner异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if busId = cs.BusID; busId == 0 {
		res["message"] = "bus_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	sampleRate = cs.SampleRate
	lowestSampleRate = cs.LowestSampleRate
	dwdTableName = cs.DatacenterDwdTableName
	appKeys = cs.AppKeys
	dbName = cs.Databases
	tableName = cs.TableName
	distributedTableName = cs.DistributedTableName
	shared = cs.Shared
	topic = cs.Topic
	activity = cs.Activity
	datacenterAppID = cs.DatacenterAppID
	state = cs.State
	level = cs.Level
	dataCount = cs.DataCount
	isWideTable = cs.IsWideTable
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventUpdate(c, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName, activity, state, level, isWideTable, shared, sampleRate, eventId, datacenterAppID, busId, dataCount, lowestSampleRate))
}

func apmEventFieldSet(c *bm.Context) {
	p := new(apm.EventFieldReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.EventID == 0 && p.CommonFieldsFlag != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmEventFieldSet(c, p))
}

func apmEventFieldList(c *bm.Context) {
	p := new(apm.EventFieldReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmEventFieldList(c, p.EventID))
}

func apmEventSql(c *bm.Context) {
	var (
		res     = map[string]interface{}{}
		params  = c.Request.Form
		eventID int64
		err     error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEventSql(c, eventID))
}

func apmEventAdvancedList(c *bm.Context) {
	var (
		res     = map[string]interface{}{}
		params  = c.Request.Form
		eventID int64
		err     error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEventAdvancedList(c, eventID))
}

func apmEventAdvancedAdd(c *bm.Context) {
	var (
		res                                                               = map[string]interface{}{}
		params                                                            = c.Request.Form
		eventID, displayType                                              int64
		fieldName, title, description, queryType, mappingGroup, customSql string
		err                                                               error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fieldName = params.Get("field_name"); fieldName == "" {
		res["message"] = "field_name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if title = params.Get("title"); title == "" {
		res["message"] = "title异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if displayType, err = strconv.ParseInt(params.Get("display_type"), 10, 64); err != nil {
		res["message"] = "display_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	queryType = params.Get("query_type")
	mappingGroup = params.Get("mapping_group")
	customSql = params.Get("custom_sql")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventAdvancedAdd(c, eventID, displayType, fieldName, title, description, queryType, mappingGroup, customSql, userName))
}

func apmEventAdvancedDel(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		params = c.Request.Form
		id     int64
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventAdvancedDel(c, id))
}

func apmEventAdvancedUpdate(c *bm.Context) {
	var (
		res                                                    = map[string]interface{}{}
		params                                                 = c.Request.Form
		id, displayType                                        int64
		title, description, queryType, mappingGroup, customSql string
		err                                                    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if title = params.Get("title"); title == "" {
		res["message"] = "title异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if displayType, err = strconv.ParseInt(params.Get("display_type"), 10, 64); err != nil {
		res["message"] = "display_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	queryType = params.Get("query_type")
	mappingGroup = params.Get("mapping_group")
	customSql = params.Get("custom_sql")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventAdvancedUpdate(c, id, displayType, title, description, queryType, mappingGroup, customSql, userName))
}

func apmCommandGroupAdvancedList(c *bm.Context) {
	var (
		res              = map[string]interface{}{}
		params           = c.Request.Form
		appKey           string
		eventId, groupId int64
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if groupId, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmCommandGroupAdvancedList(c, appKey, eventId, groupId))
}

func apmCommandGroupAdvancedAdd(c *bm.Context) {
	var (
		res                                                       = map[string]interface{}{}
		params                                                    = c.Request.Form
		appKey, fieldName, title, description, queryType, mapping string
		displayType                                               int
		eventId, groupId                                          int64
		err                                                       error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fieldName = params.Get("field_name"); fieldName == "" {
		res["message"] = "field_name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if title = params.Get("title"); title == "" {
		res["message"] = "title异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if displayType, err = strconv.Atoi(params.Get("display_type")); err != nil {
		res["message"] = "display_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if groupId, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	queryType = params.Get("query_type")
	mapping = params.Get("mapping")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCommandGroupAdvancedAdd(c, appKey, fieldName, title, description, queryType, mapping, userName, displayType, eventId, groupId))
}

func apmCommandGroupAdvancedDel(c *bm.Context) {
	var (
		params                   = c.Request.Form
		res                      = map[string]interface{}{}
		appKey                   string
		itemId, eventId, groupId int64
		err                      error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	if itemId, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if groupId, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCommandGroupAdvancedDel(c, appKey, eventId, groupId, itemId))
}

func apmCommandGroupAdvancedUpdate(c *bm.Context) {
	var (
		res                                            = map[string]interface{}{}
		params                                         = c.Request.Form
		appKey, title, description, queryType, mapping string
		displayType                                    int
		itemId, eventId, groupId                       int64
		err                                            error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if title = params.Get("title"); title == "" {
		res["message"] = "title异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if displayType, err = strconv.Atoi(params.Get("display_type")); err != nil {
		res["message"] = "display_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if groupId, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if itemId, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	queryType = params.Get("query_type")
	mapping = params.Get("mapping")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCommandGroupAdvancedUpdate(c, appKey, title, description, queryType, mapping, userName, displayType, eventId, groupId, itemId))
}

func apmMoniCalculate(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		cType       string
		matchOption *apm.MatchOption
	)
	if cType = params.Get("class_type"); cType == "" {
		res["message"] = "class_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmMoniCalculate(c, cType, matchOption))
}

func apmMoniLine(c *bm.Context) {
	var (
		params                         = c.Request.Form
		res                            = map[string]interface{}{}
		cType                          string
		eventID, busID, commandGroupId int64
		matchOption                    *apm.MatchOption
		err                            error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cType = params.Get("class_type"); cType == "" {
		res["message"] = "监控图统计类型class_type不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	busID, _ = strconv.ParseInt(params.Get("bus_id"), 10, 64)
	commandGroupId, _ = strconv.ParseInt(params.Get("command_group_id"), 10, 64)
	c.JSON(s.ApmSvr.ApmMoniLine(c, matchOption.AppKey, cType, eventID, busID, commandGroupId, matchOption))
}

func apmMoniPie(c *bm.Context) {
	var (
		params                         = c.Request.Form
		res                            = map[string]interface{}{}
		cType, column                  string
		eventID, busID, commandGroupId int64
		matchOption                    *apm.MatchOption
		err                            error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cType = params.Get("class_type"); cType == "" {
		res["message"] = "监控图统计类型class_type不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if column = params.Get("column"); column == "" {
		res["message"] = "维度字段column不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	busID, _ = strconv.ParseInt(params.Get("bus_id"), 10, 64)
	commandGroupId, _ = strconv.ParseInt(params.Get("command_group_id"), 10, 64)
	c.JSON(s.ApmSvr.ApmMoniPie(c, matchOption.AppKey, cType, column, eventID, busID, commandGroupId, matchOption))
}

func apmMoniNetInfoList(c *bm.Context) {
	var (
		res                            = map[string]interface{}{}
		params                         = c.Request.Form
		column                         string
		eventID, busID, commandGroupId int64
		matchOption                    *apm.MatchOption
		err                            error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.StartTime == 0 {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.IntervalTime == "" {
		matchOption.IntervalTime = "5 minute"
	}
	if matchOption.OrderBy == "" {
		matchOption.OrderBy = "count() DESC"
	}
	if matchOption.Limit == 0 {
		matchOption.Limit = 50
	}
	column = params.Get("column")
	busID, _ = strconv.ParseInt(params.Get("bus_id"), 10, 64)
	commandGroupId, _ = strconv.ParseInt(params.Get("command_group_id"), 10, 64)
	c.JSON(s.ApmSvr.ApmMoniNetInfoList(c, matchOption.AppKey, column, eventID, busID, commandGroupId, matchOption))
}

func apmMoniMetricInfoList(c *bm.Context) {
	var (
		res         = map[string]interface{}{}
		params      = c.Request.Form
		column      string
		matchOption *apm.MatchOption
		err         error
	)
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	column = params.Get("column")
	c.JSON(s.ApmSvr.ApmMoniMetricInfoList(c, column, matchOption))
}

func apmMoniCountInfoList(c *bm.Context) {
	var (
		res                            = map[string]interface{}{}
		params                         = c.Request.Form
		column                         string
		eventID, busID, commandGroupId int64
		matchOption                    *apm.MatchOption
		err                            error
	)
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	column = params.Get("column")
	busID, _ = strconv.ParseInt(params.Get("bus_id"), 10, 64)
	commandGroupId, _ = strconv.ParseInt(params.Get("command_group_id"), 10, 64)
	c.JSON(s.ApmSvr.ApmMoniCountInfoList(c, column, eventID, busID, commandGroupId, matchOption))
}

func apmMoniStatisticsInfoList(c *bm.Context) {
	var (
		res                            = map[string]interface{}{}
		params                         = c.Request.Form
		appKey, column                 string
		eventID, busID, commandGroupId int64
		matchOption                    *apm.MatchOption
		err                            error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventID, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.StartTime == 0 {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.IntervalTime == "" {
		matchOption.IntervalTime = "5 minute"
	}
	if matchOption.OrderBy == "" {
		matchOption.OrderBy = "count() DESC"
	}
	if matchOption.Limit == 0 {
		matchOption.Limit = 50
	}
	column = params.Get("column")
	busID, _ = strconv.ParseInt(params.Get("bus_id"), 10, 64)
	commandGroupId, _ = strconv.ParseInt(params.Get("command_group_id"), 10, 64)
	c.JSON(s.ApmSvr.ApmMoniStatisticsInfoList(c, matchOption.AppKey, column, eventID, busID, commandGroupId, matchOption))
}

func apmCommandGroupList(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, filterKey string
		busId, eventId    int64
		pn, ps            int
		err               error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	filterKey = params.Get("filter_key")
	if busId, err = strconv.ParseInt(params.Get("bus_id"), 10, 64); err != nil {
		log.Error("strconv.ParseInt error(%v)", err)
	}
	c.JSON(s.ApmSvr.ApmCommandGroupList(c, appKey, filterKey, eventId, busId, ps, pn))
}

func apmCommandGroupAdd(c *bm.Context) {
	var (
		params                                    = c.Request.Form
		res                                       = map[string]interface{}{}
		appKey, name, urls, description, userName string
		busId, eventId                            int64
		err                                       error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if urls = params.Get("urls"); urls == "" {
		res["message"] = "urls异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if busId, err = strconv.ParseInt(params.Get("bus_id"), 10, 64); err != nil {
		res["message"] = "bus_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCommandGroupAdd(c, appKey, name, urls, description, userName, busId, eventId))
}

func apmCommandGroupDel(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey      string
		id, eventId int64
		err         error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCommandGroupDel(c, appKey, id, eventId))
}

func apmCommandGroupUpdate(c *bm.Context) {
	var (
		params                                    = c.Request.Form
		res                                       = map[string]interface{}{}
		appKey, name, urls, description, userName string
		id, eventId                               int64
		err                                       error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if urls = params.Get("urls"); urls == "" {
		res["message"] = "urls异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCommandGroupUpdate(c, appKey, urls, description, userName, id, eventId))
}

func apmCommandList(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, filterKey string
		eventId, groupId  int64
		err               error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	filterKey = params.Get("filter_key")
	groupId, _ = strconv.ParseInt(params.Get("command_group_id"), 10, 64)
	c.JSON(s.ApmSvr.ApmCommandList(c, appKey, filterKey, eventId, groupId))
}

func apmAggregateNetList(c *bm.Context) {
	var (
		params                     = c.Request.Form
		res                        = map[string]interface{}{}
		appKey, command, queryType string
		startTime, endTime         int64
		err                        error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if command = params.Get("command"); command == "" {
		res["message"] = "command异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if queryType = params.Get("query_type"); queryType == "" {
		res["message"] = "query_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if startTime, err = strconv.ParseInt(params.Get("start_time"), 10, 64); err != nil {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if endTime, err = strconv.ParseInt(params.Get("end_time"), 10, 64); err != nil {
		res["message"] = "end_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmAggregateNetList(c, appKey, command, queryType, startTime, endTime))
}

func apmAggregateCrashList(c *bm.Context) {
	var (
		params                         = c.Request.Form
		res                            = map[string]interface{}{}
		appKey, versionCode, queryType string
		isAllVersion, dataType         int
		startTime, endTime             int64
		err                            error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if queryType = params.Get("query_type"); queryType == "" {
		res["message"] = "query_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isAllVersion, err = strconv.Atoi(params.Get("is_all_version")); err != nil {
		isAllVersion = 0
	}
	if dataType, err = strconv.Atoi(params.Get("data_type")); err != nil {
		dataType = 0
	}
	if startTime, err = strconv.ParseInt(params.Get("start_time"), 10, 64); err != nil {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if endTime, err = strconv.ParseInt(params.Get("end_time"), 10, 64); err != nil {
		res["message"] = "end_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	versionCode = params.Get("version_code")
	c.JSON(s.ApmSvr.ApmAggregateCrashList(c, appKey, versionCode, queryType, isAllVersion, dataType, startTime, endTime))
}

func apmAggregateANRList(c *bm.Context) {
	var (
		params                         = c.Request.Form
		res                            = map[string]interface{}{}
		appKey, versionCode, queryType string
		isAllVersion, dataType         int
		startTime, endTime             int64
		err                            error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if queryType = params.Get("query_type"); queryType == "" {
		res["message"] = "query_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isAllVersion, err = strconv.Atoi(params.Get("is_all_version")); err != nil {
		isAllVersion = 0
	}
	if startTime, err = strconv.ParseInt(params.Get("start_time"), 10, 64); err != nil {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if endTime, err = strconv.ParseInt(params.Get("end_time"), 10, 64); err != nil {
		res["message"] = "end_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if dataType, err = strconv.Atoi(params.Get("data_type")); err != nil {
		dataType = 0
	}
	versionCode = params.Get("version_code")
	c.JSON(s.ApmSvr.ApmAggregateANRList(c, appKey, versionCode, queryType, isAllVersion, dataType, startTime, endTime))
}

func apmAggregateSetupList(c *bm.Context) {
	var (
		params                         = c.Request.Form
		res                            = map[string]interface{}{}
		appKey, versionCode, queryType string
		isAllVersion                   int
		startTime, endTime             int64
		err                            error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if queryType = params.Get("query_type"); queryType == "" {
		res["message"] = "query_type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isAllVersion, err = strconv.Atoi(params.Get("is_all_version")); err != nil {
		isAllVersion = 0
	}
	if startTime, err = strconv.ParseInt(params.Get("start_time"), 10, 64); err != nil {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if endTime, err = strconv.ParseInt(params.Get("end_time"), 10, 64); err != nil {
		res["message"] = "end_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	versionCode = params.Get("version_code")
	c.JSON(s.ApmSvr.ApmAggregateSetupList(c, appKey, versionCode, queryType, isAllVersion, startTime, endTime))
}

func apmFlowmapRouteList(c *bm.Context) {
	var (
		res         = map[string]interface{}{}
		matchOption *apm.MatchOption
	)
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.AppKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.VersionCode == "" {
		res["message"] = "versionCode异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.StartTime == 0 {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.EndTime == 0 {
		res["message"] = "end_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlowmapRouteList(c, matchOption))
}

func apmFlowmapRouteAliasList(c *bm.Context) {
	var (
		params            = c.Request.Form
		appKey, filterKey string
		busID             int64
		res               = map[string]interface{}{}
		err               error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if busIDStr := params.Get("bus_id"); busIDStr != "" {
		if busID, err = strconv.ParseInt(busIDStr, 10, 64); err != nil {
			res["message"] = "bus_id异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		busID = -1
	}
	filterKey = params.Get("filter_key")
	c.JSON(s.ApmSvr.ApmFlowmapRouteAliasList(c, appKey, filterKey, busID))
}

func apmFlowmapRouteAliasAdd(c *bm.Context) {
	var (
		params                                  = c.Request.Form
		res                                     = map[string]interface{}{}
		appKey, routeName, routeAlias, userName string
		busID                                   int64
		err                                     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if routeName = params.Get("route_name"); routeName == "" {
		res["message"] = "route_name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if routeAlias = params.Get("route_alias"); routeAlias == "" {
		res["message"] = "route_alias异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if busIDStr := params.Get("bus_id"); busIDStr != "" {
		if busID, err = strconv.ParseInt(busIDStr, 10, 64); err != nil {
			res["message"] = "bus_id异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmFlowmapRouteAliasAdd(c, appKey, routeName, routeAlias, userName, busID))
}

func apmFlowmapRouteAliasUpdate(c *bm.Context) {
	var (
		params                          = c.Request.Form
		res                             = map[string]interface{}{}
		id, busID                       int64
		routeName, routeAlias, userName string
		err                             error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil && id != 0 {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if busID, err = strconv.ParseInt(params.Get("bus_id"), 10, 64); err != nil {
		res["message"] = "bus_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if routeName = params.Get("route_name"); routeName == "" {
		res["message"] = "route_name不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if routeAlias = params.Get("route_alias"); routeAlias == "" {
		res["message"] = "route_alias不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmFlowmapRouteAliasUpdate(c, id, routeName, routeAlias, userName, busID))
}

func apmFlowmapRouteAliasDel(c *bm.Context) {
	var (
		params   = c.Request.Form
		res      = map[string]interface{}{}
		id       int64
		userName string
		err      error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil && id != 0 {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmFlowmapRouteAliasDel(c, id, userName))
}

func apmWebTrack(c *bm.Context) {
	var (
		userName    string
		trackParams *apm.WebTrackParams
		bs          []byte
		err         error
	)
	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	// params
	if err = json.Unmarshal(bs, &trackParams); err != nil {
		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	for _, model := range trackParams.Models {
		model.Username = userName
	}
	c.JSON(nil, s.ApmSvr.ApmWebTrack(c, trackParams))
}

func apmEventSetting(c *bm.Context) {
	var (
		params  = c.Request.Form
		appKey  string
		eventId int64
		res     = map[string]interface{}{}
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEventSetting(c, appKey, eventId))
}

func apmDetailSetup(c *bm.Context) {
	var (
		matchOption *apm.MatchOption
		res         = map[string]interface{}{}
	)
	matchOption = new(apm.MatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.Buvid == "" && matchOption.Mid == "" {
		res["message"] = "buvid/mid 不能同时为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmDetailSetup(c, matchOption))
}

func apmMetricList(c *bm.Context) {
	p := new(apm.PrometheusMetricListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmMetricList(c, p))
}

func apmMetricSet(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.PrometheusMetric)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	if p.Metric == "" {
		res["message"] = "metric名称不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.ExecSQL == "" {
		res["message"] = "执行SQL不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.ApmDatabaseName == "" {
		res["message"] = "数据库名称不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.ApmTableName == "" {
		res["message"] = "数据表不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.LabeledKeys == "" {
		res["message"] = "聚合字段不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.BusID == 0 {
		res["message"] = "业务组ID不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmMetricSet(c, p))
}

func apmMetricDel(c *bm.Context) {
	var (
		params    = c.Request.Form
		IsUndoDel int64
		res       = map[string]interface{}{}
		err       error
	)
	p := new(apm.PrometheusMetric)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Metric == "" {
		res["message"] = "metric不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if IsUndoDel, err = strconv.ParseInt(params.Get("is_undo_del"), 10, 64); err != nil {
		IsUndoDel = 0
	}
	c.JSON(nil, s.ApmSvr.ApmMetricDel(c, p, IsUndoDel))
}

func apmMetricPublish(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.PrometheusMetricPublishReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	if p.Description == "" {
		res["message"] = "备注不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmMetricPublish(c, p))
}

func apmMetricPublishList(c *bm.Context) {
	p := new(apm.PrometheusMetricPublishListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmMetricPublishList(c, p))
}

func apmMetricPublishDiff(c *bm.Context) {
	c.JSON(s.ApmSvr.ApmMetricPublishDiff(c))
}

func apmMetricPublishRollback(c *bm.Context) {
	p := new(apm.PrometheusMetricPublishRollbackReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(nil, s.ApmSvr.ApmMetricPublishRollback(c, p))
}

func apmFlinkJobList(c *bm.Context) {
	p := new(apm.FlinkJobReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobList(c, p))
}

func apmFlinkJobAdd(c *bm.Context) {
	var (
		res = map[string]interface{}{}
		err error
	)
	p := new(apm.FlinkJobReq)
	if err = c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	if _, err = strconv.ParseInt(p.LogID, 10, 64); err != nil {
		res["message"] = "log_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.Name == "" {
		res["message"] = "name不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobAdd(c, p))
}

func apmFlinkJobUpdate(c *bm.Context) {
	var (
		res = map[string]interface{}{}
		err error
	)
	p := new(apm.FlinkJobReq)
	if err = c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	if p.ID == 0 {
		res["message"] = "id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if _, err = strconv.ParseInt(p.LogID, 10, 64); err != nil {
		res["message"] = "log_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobUpdate(c, p))
}

func apmFlinkJobDel(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.FlinkJobReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.ID == 0 {
		res["message"] = "id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobDel(c, p))
}

func apmFlinkJobRelationList(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.EventFlinkRelReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.JobID == 0 {
		res["message"] = "flink_job_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobRelationList(c, p))
}

func apmFlinkJobRelationAdd(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.EventFlinkRelReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	if p.JobID == 0 {
		res["message"] = "flink_job_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.EventID == 0 {
		res["message"] = "event_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmFlinkJobRelationAdd(c, p))
}

func apmFlinkJobRelationDel(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.EventFlinkRelReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.JobID == 0 {
		res["message"] = "flink_job_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.EventID == 0 {
		res["message"] = "event_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmFlinkJobRelationDel(c, p))
}

func apmFlinkJobPublish(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.EventFlinkRelReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	if p.Description == "" {
		res["message"] = "备注不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.JobID == 0 {
		res["message"] = "flink_job_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobPublish(c, p))
}

func apmFlinkJobPublishList(c *bm.Context) {
	var res = map[string]interface{}{}
	p := new(apm.EventFlinkRelPublishListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.FlinkJobID == 0 {
		res["message"] = "flink_job_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobPublishList(c, p))
}

func apmFlinkJobPublishDiff(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		params = c.Request.Form
		jobID  int64
		err    error
	)
	if jobID, err = strconv.ParseInt(params.Get("flink_job_id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmFlinkJobPublishDiff(c, jobID))
}

func apmCrashRule(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		params = c.Request.Form
		ruleID int64
		err    error
	)
	if ruleID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmCrashRule(c, ruleID))
}

func apmCrashRuleList(c *bm.Context) {
	var (
		res = map[string]interface{}{}
		p   = new(apm.CrashRuleReq)
	)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.Pn < 1 {
		p.Pn = 1
	}
	if p.Ps < 1 {
		p.Ps = 20
	}
	c.JSON(s.ApmSvr.ApmCrashRuleList(c, p))
}

func apmCrashRuleAdd(c *bm.Context) {
	var (
		res = map[string]interface{}{}
		p   = new(apm.CrashRuleReq)
	)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.AppKeys == "" {
		res["message"] = "app_keys不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.BusID == 0 {
		res["message"] = "bus_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.RuleName == "" {
		res["message"] = "rule_name不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.KeyWords == "" {
		res["message"] = "key_words不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.Description == "" {
		res["message"] = "description不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmCrashRuleAdd(c, p))
}

func apmCrashRuleDel(c *bm.Context) {
	var (
		res = map[string]interface{}{}
		p   = new(apm.CrashRuleReq)
	)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.ID == 0 {
		res["message"] = "id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmCrashRuleDel(c, p))
}

func apmCrashRuleUpdate(c *bm.Context) {
	var (
		res = map[string]interface{}{}
		p   = new(apm.CrashRuleReq)
	)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.ID == 0 {
		res["message"] = "id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.AppKeys == "" {
		res["message"] = "app_keys不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.BusID == 0 {
		res["message"] = "bus_id不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.RuleName == "" {
		res["message"] = "rule_name不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.KeyWords == "" {
		res["message"] = "key_words不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if p.Description == "" {
		res["message"] = "description不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmCrashRuleUpdate(c, p))
}

func apmAppEventList(c *bm.Context) {
	var (
		res                                   = map[string]interface{}{}
		appKey, name, busName, logId, orderBy string
		pn, ps                                int
		stateConv                             int64
		state                                 int8
		params                                = c.Request.Form
		err                                   error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if stateConv, err = strconv.ParseInt(params.Get("state"), 10, 64); err != nil {
		stateConv = 0
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	state = int8(stateConv)
	name = params.Get("name")
	busName = params.Get("bus_name")
	logId = params.Get("log_id")
	orderBy = params.Get("order_by")
	c.JSON(s.ApmSvr.ApmAppEventList(c, appKey, name, busName, logId, orderBy, pn, ps, state))
}

func apmAppEventRelAdd(c *bm.Context) {
	p := new(apm.EventDatacenterRel)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAppEventRelAdd(c, p))
}

func apmEventFieldBillionsSync(c *bm.Context) {
	var (
		res     = map[string]interface{}{}
		eventId int64
		params  = c.Request.Form
		err     error
	)
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventFieldBillionsSync(c, eventId))
}

func apmEventFieldPublish(c *bm.Context) {
	p := new(apm.EventFieldReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.EventID == 0 && p.CommonFieldsFlag != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmEventFieldPublish(c, p.EventID, p.IsIgnoreBillions, p.Operator))
}

func apmEventFieldPublishDiff(c *bm.Context) {
	p := new(apm.EventFieldReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.EventID == 0 && p.CommonFieldsFlag != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEventFieldPublishDiff(c, p.EventID))
}

func apmEventFieldPublishHistory(c *bm.Context) {
	p := new(apm.EventFieldReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.EventID == 0 && p.CommonFieldsFlag != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if p.Pn < 1 {
		p.Pn = 1
	}
	if p.Ps < 1 {
		p.Ps = 20
	}
	c.JSON(s.ApmSvr.ApmEventFieldPublishHistory(c, p.EventID, p.Pn, p.Ps))
}

func apmEventFieldDiff(c *bm.Context) {
	var (
		res              = map[string]interface{}{}
		eventId, version int64
		params           = c.Request.Form
		err              error
	)
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version, err = strconv.ParseInt(params.Get("version"), 10, 64); err != nil {
		res["message"] = "version异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEventFieldDiff(c, eventId, version))
}

func apmEventFieldTypeSync(c *bm.Context) {
	var (
		res                       = map[string]interface{}{}
		eventId, commonFieldsFlag int64
		params                    = c.Request.Form
		err                       error
	)
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if commonFieldsFlag, err = strconv.ParseInt(params.Get("common_fields_flag"), 10, 64); err != nil {
		res["message"] = "common_fields_flag异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId == 0 && commonFieldsFlag != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventFieldTypeSync(c, eventId))
}

func apmAppCommonFieldGroupAdd(c *bm.Context) {
	p := new(apm.EventCommonFieldGroupReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAppCommonFieldGroupAdd(c, p.AppKey, p.Name, p.Description, p.Operator, p.IsDefault, p.Fields))
}

func apmAppCommonFieldGroupUpdate(c *bm.Context) {
	p := new(apm.EventCommonFieldGroupReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAppCommonFieldGroupUpdate(c, p.AppKey, p.Name, p.Description, p.Operator, p.IsDefault, p.Id, p.Fields))
}

func apmAppCommonFieldGroupDel(c *bm.Context) {
	var (
		res     = map[string]interface{}{}
		groupId int64
		params  = c.Request.Form
		err     error
	)
	if groupId, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmAppCommonFieldGroupDel(c, groupId))
}

func apmAppCommonFieldGroupList(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		appKey string
		pn, ps int
		params = c.Request.Form
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	c.JSON(s.ApmSvr.ApmAppCommonFieldGroupList(c, appKey, pn, ps))
}

func apmAppCommonFieldGroup(c *bm.Context) {
	var (
		res     = map[string]interface{}{}
		groupId int64
		params  = c.Request.Form
		err     error
	)
	if groupId, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmAppCommonFieldGroup(c, groupId))
}

func apmEventAlertAdd(c *bm.Context) {
	p := new(apm.EventAlertAddReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	p.Creator = userName
	c.JSON(nil, s.ApmSvr.ApmEventAlertAdd(c, p))
}

func apmEventAlertUpdate(c *bm.Context) {
	var (
		res = map[string]interface{}{}
	)
	p := new(apm.EventAlertUpdateReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Id == 0 {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmEventAlertUpdate(c, p))
}

func apmEventAlertDel(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		id     int64
		params = c.Request.Form
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmEventAlertDel(c, id))
}

func apmEventAlertList(c *bm.Context) {
	p := new(apm.EventAlertQueryReq)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmEventAlertList(c, p))
}

func apmEventAlert(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		id     int64
		params = c.Request.Form
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ApmSvr.ApmEventAlert(c, id))
}

func apmEventAlertSwitch(c *bm.Context) {
	var (
		res              = map[string]interface{}{}
		id, isEnableConv int64
		isEnable         int8
		params           = c.Request.Form
		err              error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isEnableConv, err = strconv.ParseInt(params.Get("is_enable"), 10, 8); err != nil {
		res["message"] = "is_enable 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isEnableConv > 0 {
		isEnableConv = 1
	} else {
		isEnableConv = -1
	}
	isEnable = int8(isEnableConv)
	c.JSON(nil, s.ApmSvr.ApmEventAlertSwitch(c, isEnable, id))
}

func apmEventCKTableCreate(c *bm.Context) {
	p := new(apm.CKTableCreateReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmEventCKTableCreate(c, p))
}

func apmAlertRuleList(c *bm.Context) {
	p := new(apm.AlertRuleListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmAlertRuleList(c, p))
}

func apmAlertRuleSet(c *bm.Context) {
	p := new(apm.AlertRuleSetReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAlertRuleSet(c, p))
}

func apmAlertRuleDel(c *bm.Context) {
	var (
		id     int64
		err    error
		params = c.Request.Form
		res    = map[string]interface{}{}
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ApmSvr.ApmAlertRuleDel(c, id))
}

func apmAlertList(c *bm.Context) {
	p := new(apm.AlertListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmAlertList(c, p))
}

func apmAlertAdd(c *bm.Context) {
	p := new(apm.AlertAddReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAlertAdd(c, p))
}

func apmAlertUpdate(c *bm.Context) {
	p := new(apm.AlertUpdateReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAlertUpdate(c, p))
}

func apmAlertIndicatorInfo(c *bm.Context) {
	p := new(apm.AlertIndicatorReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmAlertIndicatorInfo(c, p))
}

func apmAlertReason(c *bm.Context) {
	var (
		alertMd5 string
		params   = c.Request.Form
		res      = map[string]interface{}{}
	)
	if alertMd5 = params.Get("alert_md5"); alertMd5 == "" {
		res["message"] = "alert_md5 异常"
		c.JSONMap(res, ecode.RequestErr)
	}
	c.JSON(s.ApmSvr.ApmAlertReason(c, alertMd5))
}

func apmAlertReasonConfig(c *bm.Context) {
	p := new(apm.AlertReasonConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmAlertReasonConfig(c, p))

}

func apmAlertReasonConfigAdd(c *bm.Context) {
	p := new(apm.AlertReasonConfigAddReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAlertReasonConfigAdd(c, p))
}

func apmAlertReasonConfigUpdate(c *bm.Context) {
	p := new(apm.AlertReasonConfigUpdateReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.ApmAlertReasonConfigUpdate(c, p))
}

func apmAlertReasonConfigDelete(c *bm.Context) {
	p := new(apm.AlertReasonConfigDeleteReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(nil, s.ApmSvr.ApmAlertReasonConfigDelete(c, p))
}

func apmEventSampleRateAdd(ctx *bm.Context) {
	p := new(apm.AddEventSampleRateReq)
	if err := ctx.BindWith(p, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	ctx.JSON(nil, s.ApmSvr.ApmEventSampleRateAdd(ctx, p))
}

func apmEventSampleRateDel(ctx *bm.Context) {
	p := new(apm.DeleteEventSampleRateReq)
	if err := ctx.BindWith(p, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	ctx.JSON(nil, s.ApmSvr.ApmEventSampleRateDelete(ctx, p))
}

func apmEventSampleRateList(ctx *bm.Context) {
	p := new(apm.EventSampleRateListReq)
	if err := ctx.BindWith(p, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	ctx.JSON(s.ApmSvr.ApmEventSampleRateList(ctx, p))
}

func apmEventSampleRateConfig(ctx *bm.Context) {
	p := new(apm.EventSampleRateConfigReq)
	if err := ctx.BindWith(p, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	ctx.JSON(s.ApmSvr.ApmEventSampleRateConfig(ctx, p))
}

func apmEventMonitorNotifyConfig(c *bm.Context) {
	p := new(apm.EventMonitorNotifyConfigReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmEventMonitorNotifyConfig(c, p))
}

func apmEventMonitorNotifyConfigList(c *bm.Context) {
	p := new(apm.EventMonitorNotifyConfigListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.ApmSvr.ApmEventMonitorNotifyConfigList(c, p))

}

func apmEventMonitorNotifyConfigSet(c *bm.Context) {
	p := new(apm.EventMonitorNotifyConfigSetReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(s.ApmSvr.ApmEventMonitorNotifyConfigSet(c, p))

}
